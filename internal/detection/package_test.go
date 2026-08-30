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

func TestPackageDetectorKeepsMissingResultWhenManagerCannotRun(t *testing.T) {
	pkg := model.Package{ID: "gradle", Manager: "winget", ManagerID: "Gradle.Gradle", CheckProgram: "gradle", CheckArgs: []string{"--version"}}
	runner := packageRunner{results: map[string]ports.Result{
		"gradle": {Err: exec.ErrNotFound},
		"winget": {Err: errors.New("系统无法访问此文件")},
	}}

	result := (PackageDetector{Runner: runner, Platform: model.PlatformWindows}).Detect(context.Background(), pkg)

	if result.Status != StatusMissing || result.Source != "gradle" {
		t.Fatalf("工具命令明确不存在时应保留缺失结论: %#v", result)
	}
}
