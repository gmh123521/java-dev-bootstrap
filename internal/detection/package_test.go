package detection

import (
	"context"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"testing"

	"github.com/gmh123521/java-dev-bootstrap/internal/model"
	"github.com/gmh123521/java-dev-bootstrap/internal/ports"
)

type packageRunner struct {
	results map[string]ports.Result
}

func (r packageRunner) Run(_ context.Context, command ports.Command) ports.Result {
	result := r.results[command.Program]
	result.Command = command
	return result
}

func TestPackageDetectorFindsDesktopApplicationByPath(t *testing.T) {
	dir := t.TempDir()
	application := filepath.Join(dir, "idea64.exe")
	if err := os.WriteFile(application, nil, 0o600); err != nil {
		t.Fatal(err)
	}
	pkg := model.Package{ID: "intellij", CheckPaths: []string{application}}

	result := (PackageDetector{}).Detect(context.Background(), pkg)

	if result.Status != StatusInstalled || result.Source != "path" || result.Path != application {
		t.Fatalf("应通过路径识别桌面软件: %#v", result)
	}
}

func TestPackageDetectorFallsBackToPackageManager(t *testing.T) {
	pkg := model.Package{ID: "intellij", Manager: "winget", ManagerID: "JetBrains.IntelliJIDEA.Community"}
	runner := packageRunner{results: map[string]ports.Result{"winget": {Output: "IntelliJ IDEA Community"}}}

	result := (PackageDetector{Runner: runner, Platform: model.PlatformWindows}).Detect(context.Background(), pkg)

	if result.Status != StatusInstalled || result.Source != "winget" {
		t.Fatalf("应通过 winget 识别软件: %#v", result)
	}
}

func TestPackageDetectorReportsMissingAfterAllChecks(t *testing.T) {
	pkg := model.Package{ID: "gradle", Manager: "winget", ManagerID: "Gradle.Gradle", CheckProgram: "gradle", CheckArgs: []string{"--version"}}
	runner := packageRunner{results: map[string]ports.Result{
		"gradle": {Err: exec.ErrNotFound},
		"winget": {Output: "没有找到符合条件的已安装程序包", Err: errors.New("退出代码 1")},
	}}

	result := (PackageDetector{Runner: runner, Platform: model.PlatformWindows}).Detect(context.Background(), pkg)

	if result.Status != StatusMissing {
		t.Fatalf("所有检查均未发现时应标记缺失: %#v", result)
	}
}

func TestPackageDetectorReportsErrorWhenManagerCannotRun(t *testing.T) {
	pkg := model.Package{ID: "gradle", Manager: "winget", ManagerID: "Gradle.Gradle", CheckProgram: "gradle", CheckArgs: []string{"--version"}}
	runner := packageRunner{results: map[string]ports.Result{
		"gradle": {Err: exec.ErrNotFound},
		"winget": {Err: errors.New("系统无法访问此文件")},
	}}

	result := (PackageDetector{Runner: runner, Platform: model.PlatformWindows}).Detect(context.Background(), pkg)

	if result.Status != StatusError || result.Source != "winget" {
		t.Fatalf("包管理器查询失败且无输出时必须阻止自动安装: %#v", result)
	}
}

func TestPackageDetectorUsesInstalledManagerPackageWhenPathCommandIsOld(t *testing.T) {
	pkg := model.Package{ID: "jdk", Manager: "winget", ManagerID: "EclipseAdoptium.Temurin.21.JDK", CheckProgram: "javac", CheckArgs: []string{"-version"}, MinVersion: 21}
	runner := packageRunner{results: map[string]ports.Result{
		"javac":  {Output: "javac 17.0.12"},
		"winget": {Output: "Eclipse Temurin JDK 21"},
	}}

	result := (PackageDetector{Runner: runner, Platform: model.PlatformWindows}).Detect(context.Background(), pkg)

	if result.Status != StatusInstalled || result.Source != "winget" {
		t.Fatalf("包管理器中的满足版本软件应覆盖 PATH 旧版本: %#v", result)
	}
}

func TestPackageDetectorDoesNotTreatManagerErrorsAsMissing(t *testing.T) {
	pkg := model.Package{ID: "intellij", Manager: "winget", ManagerID: "JetBrains.IntelliJIDEA.Community"}
	runner := packageRunner{results: map[string]ports.Result{
		"winget": {Output: "Failed when opening source; database is corrupted", Err: errors.New("退出代码 1")},
	}}

	result := (PackageDetector{Runner: runner, Platform: model.PlatformWindows}).Detect(context.Background(), pkg)

	if result.Status != StatusError {
		t.Fatalf("包管理器故障不应解释为软件缺失: %#v", result)
	}
}

func TestCandidatePatternsExpandsWindowsDrivePlaceholder(t *testing.T) {
	patterns := candidatePatterns(`{drive}:\idea\IntelliJ IDEA*\bin\idea64.exe`, model.PlatformWindows)
	foundD := false
	for _, pattern := range patterns {
		if pattern == `D:\idea\IntelliJ IDEA*\bin\idea64.exe` {
			foundD = true
		}
	}
	if !foundD {
		t.Fatalf("Windows 路径模式应覆盖 D 盘: %#v", patterns)
	}
}
