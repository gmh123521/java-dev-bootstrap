package detection

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gmh123521/java-dev-bootstrap/internal/model"
	"github.com/gmh123521/java-dev-bootstrap/internal/ports"
)

type Level string

const (
	LevelOK      Level = "ok"
	LevelWarning Level = "warning"
)

type Diagnostic struct {
	Level      Level
	Name       string
	Current    string
	Suggestion string
}

type EnvironmentDetector struct {
	Runner   ports.Runner
	Getenv   func(string) string
	LookPath func(string) (string, error)
	Stat     func(string) (os.FileInfo, error)
}

func (d EnvironmentDetector) Diagnose(ctx context.Context, current model.Platform) []Diagnostic {
	getenv := d.Getenv
	if getenv == nil {
		getenv = os.Getenv
	}
	lookPath := d.LookPath
	if lookPath == nil {
		lookPath = exec.LookPath
	}
	stat := d.Stat
	if stat == nil {
		stat = os.Stat
	}

	home := getenv("JAVA_HOME")
	if home == "" {
		return []Diagnostic{{
			Level:      LevelWarning,
			Name:       "JAVA_HOME",
			Current:    "未设置",
			Suggestion: "将 JAVA_HOME 指向当前使用的 JDK 根目录",
		}}
	}
	if info, err := stat(home); err != nil || !info.IsDir() {
		return []Diagnostic{{
			Level:      LevelWarning,
			Name:       "JAVA_HOME",
			Current:    home,
			Suggestion: "JAVA_HOME 指向的目录不存在，请改为有效 JDK 根目录",
		}}
	}

	javaName := "java"
	if current == model.PlatformWindows {
		javaName = "java.exe"
	}
	homeJava := filepath.Join(home, "bin", javaName)
	if _, err := stat(homeJava); err != nil {
		return []Diagnostic{{
			Level:      LevelWarning,
			Name:       "JAVA_HOME",
			Current:    home,
			Suggestion: fmt.Sprintf("没有找到 %s，请确认 JAVA_HOME 指向 JDK 而不是 JRE", homeJava),
		}}
	}

	items := []Diagnostic{{Level: LevelOK, Name: "JAVA_HOME", Current: homeJava}}
	pathJava, err := lookPath("java")
	if err != nil {
		return append(items, Diagnostic{
			Level:      LevelWarning,
			Name:       "Java PATH",
			Current:    "未找到 java 命令",
			Suggestion: "将 JAVA_HOME 的 bin 目录加入 PATH",
		})
	}
	if d.Runner == nil {
		return append(items, Diagnostic{Level: LevelWarning, Name: "Java 版本", Current: "未配置诊断命令执行器"})
	}

	homeVersion, homeErr := d.javaVersion(ctx, homeJava)
	pathVersion, pathErr := d.javaVersion(ctx, pathJava)
	if homeErr != nil || pathErr != nil {
		return append(items, Diagnostic{
			Level:      LevelWarning,
			Name:       "Java 版本",
			Current:    fmt.Sprintf("JAVA_HOME=%s，PATH=%s", homeVersion, pathVersion),
			Suggestion: "确认两个 Java 命令都可以正常执行",
		})
	}
	if homeVersion != pathVersion {
		return append(items, Diagnostic{
			Level:      LevelWarning,
			Name:       "Java 版本",
			Current:    fmt.Sprintf("JAVA_HOME=%s（%s），PATH=%s（%s）", homeVersion, homeJava, pathVersion, pathJava),
			Suggestion: "调整 PATH 顺序，使 java 与 JAVA_HOME 使用同一 JDK",
		})
	}
	return append(items, Diagnostic{
		Level:   LevelOK,
		Name:    "Java 版本",
		Current: fmt.Sprintf("%s（JAVA_HOME 与 PATH 一致）", homeVersion),
	})
}

func (d EnvironmentDetector) DiagnoseTools(ctx context.Context, packages []model.Package) []Diagnostic {
	items := make([]Diagnostic, 0, len(packages))
	toolDetector := ToolDetector{Runner: d.Runner, LookPath: d.LookPath}
	for _, pkg := range packages {
		if pkg.ID == "jdk" || pkg.CheckProgram == "" {
			continue
		}
		result := toolDetector.Detect(ctx, pkg)
		switch result.Status {
		case StatusInstalled:
			current := result.Version
			if result.Path != "" {
				current += "（" + result.Path + "）"
			}
			items = append(items, Diagnostic{Level: LevelOK, Name: pkg.Name, Current: current})
		case StatusOutdated:
			items = append(items, Diagnostic{
				Level:      LevelWarning,
				Name:       pkg.Name,
				Current:    "版本 " + result.Version + " 低于要求",
				Suggestion: "升级到清单要求的最低版本",
			})
		case StatusMissing:
			items = append(items, Diagnostic{
				Level:      LevelWarning,
				Name:       pkg.Name,
				Current:    "未找到 " + pkg.CheckProgram + " 命令",
				Suggestion: "确认软件已安装并将命令目录加入 PATH",
			})
		case StatusError:
			items = append(items, Diagnostic{
				Level:      LevelWarning,
				Name:       pkg.Name,
				Current:    "检测失败",
				Suggestion: result.Detail,
			})
		}
	}
	return items
}

func (d EnvironmentDetector) javaVersion(ctx context.Context, program string) (string, error) {
	result := d.Runner.Run(ctx, ports.Command{Program: program, Args: []string{"-version"}})
	if result.Err != nil {
		return "", result.Err
	}
	return VersionText(result.Output)
}
