package detection

import (
	"context"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/gmh123521/java-dev-bootstrap/internal/model"
	"github.com/gmh123521/java-dev-bootstrap/internal/ports"
)

type environmentRunner struct {
	outputs map[string]string
}

func (r environmentRunner) Run(_ context.Context, command ports.Command) ports.Result {
	return ports.Result{Command: command, Output: r.outputs[command.Program]}
}

func TestEnvironmentDetectorWarnsWhenJavaHomeMissing(t *testing.T) {
	detector := EnvironmentDetector{
		Getenv:   func(string) string { return "" },
		LookPath: func(string) (string, error) { return `C:\Java\bin\java.exe`, nil },
		Runner:   environmentRunner{outputs: map[string]string{`C:\Java\bin\java.exe`: `java version "23.0.2"`}},
	}

	items := detector.Diagnose(context.Background(), model.PlatformWindows)

	if !hasDiagnostic(items, "JAVA_HOME", LevelWarning) {
		t.Fatalf("JAVA_HOME 缺失应给出警告: %#v", items)
	}
	if !hasDiagnostic(items, "Java PATH", LevelOK) {
		t.Fatalf("JAVA_HOME 缺失时仍应诊断 PATH 中的 Java: %#v", items)
	}
}

func TestEnvironmentDetectorAcceptsMatchingJavaVersions(t *testing.T) {
	home := t.TempDir()
	homeJava := filepath.Join(home, "bin", "java.exe")
	if err := os.MkdirAll(filepath.Dir(homeJava), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(homeJava, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, "bin", "javac.exe"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	pathJava := `C:\Program Files\Common Files\Oracle\Java\javapath\java.exe`
	detector := EnvironmentDetector{
		Getenv: func(name string) string {
			if name == "JAVA_HOME" {
				return home
			}
			return ""
		},
		LookPath: func(string) (string, error) { return pathJava, nil },
		Runner: environmentRunner{outputs: map[string]string{
			homeJava: `java version "23.0.2"`,
			pathJava: `java version "23.0.2"`,
		}},
	}

	items := detector.Diagnose(context.Background(), model.PlatformWindows)

	if !hasDiagnostic(items, "Java 版本", LevelOK) {
		t.Fatalf("相同 Java 版本应通过诊断: %#v", items)
	}
}

func TestEnvironmentDetectorWarnsWhenJavaVersionsDiffer(t *testing.T) {
	home := t.TempDir()
	homeJava := filepath.Join(home, "bin", "java.exe")
	if err := os.MkdirAll(filepath.Dir(homeJava), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(homeJava, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(home, "bin", "javac.exe"), nil, 0o600); err != nil {
		t.Fatal(err)
	}
	pathJava := `C:\Java17\bin\java.exe`
	detector := EnvironmentDetector{
		Getenv: func(name string) string {
			if name == "JAVA_HOME" {
				return home
			}
			return ""
		},
		LookPath: func(string) (string, error) { return pathJava, nil },
		Runner: environmentRunner{outputs: map[string]string{
			homeJava: `java version "23.0.2"`,
			pathJava: `openjdk version "17.0.12"`,
		}},
	}

	items := detector.Diagnose(context.Background(), model.PlatformWindows)

	if !hasDiagnostic(items, "Java 版本", LevelWarning) {
		t.Fatalf("不同 Java 版本应给出警告: %#v", items)
	}
}

func TestEnvironmentDetectorRejectsJavaHomeWithoutCompiler(t *testing.T) {
	home := t.TempDir()
	homeJava := filepath.Join(home, "bin", "java.exe")
	if err := os.MkdirAll(filepath.Dir(homeJava), 0o700); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(homeJava, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	detector := EnvironmentDetector{
		Getenv: func(name string) string {
			if name == "JAVA_HOME" {
				return home
			}
			return ""
		},
		LookPath: func(string) (string, error) { return homeJava, nil },
		Runner:   environmentRunner{outputs: map[string]string{homeJava: `java version "23.0.2"`}},
	}

	items := detector.Diagnose(context.Background(), model.PlatformWindows)

	if !hasDiagnostic(items, "JDK 完整性", LevelWarning) {
		t.Fatalf("缺少 javac 时不应认定为完整 JDK: %#v", items)
	}
}

func TestEnvironmentDetectorReportsToolAvailability(t *testing.T) {
	detector := EnvironmentDetector{
		LookPath: func(program string) (string, error) { return program, nil },
		Runner: packageRunner{results: map[string]ports.Result{
			"git":    {Output: "git version 2.51.1.windows.1"},
			"gradle": {Err: exec.ErrNotFound},
		}},
	}
	packages := []model.Package{
		{ID: "git", Name: "Git", CheckProgram: "git", CheckArgs: []string{"--version"}},
		{ID: "gradle", Name: "Gradle", CheckProgram: "gradle", CheckArgs: []string{"--version"}},
	}

	items := detector.DiagnoseTools(context.Background(), packages)

	if !hasDiagnostic(items, "Git", LevelOK) || !hasDiagnostic(items, "Gradle", LevelWarning) {
		t.Fatalf("工具可用性诊断错误: %#v", items)
	}
}

type applicationDetector struct {
	results map[string]Result
}

func (d applicationDetector) Detect(_ context.Context, pkg model.Package) Result {
	return d.results[pkg.ID]
}

func TestEnvironmentDetectorReportsDesktopApplications(t *testing.T) {
	packages := []model.Package{
		{ID: "intellij", Name: "IntelliJ IDEA"},
		{ID: "docker", Name: "Docker Desktop"},
	}
	detector := applicationDetector{results: map[string]Result{
		"intellij": {Status: StatusInstalled, Source: "path", Path: `D:\idea\idea64.exe`},
		"docker":   {Status: StatusMissing, Source: "winget"},
	}}

	items := (EnvironmentDetector{}).DiagnoseApplications(context.Background(), packages, detector)

	if !hasDiagnostic(items, "IntelliJ IDEA", LevelOK) || !hasDiagnostic(items, "Docker Desktop", LevelWarning) {
		t.Fatalf("桌面软件诊断错误: %#v", items)
	}
}

func hasDiagnostic(items []Diagnostic, name string, level Level) bool {
	for _, item := range items {
		if item.Name == name && item.Level == level {
			return true
		}
	}
	return false
}
