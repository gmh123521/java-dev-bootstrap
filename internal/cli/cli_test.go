package cli

import (
	"context"
	"encoding/json"
	"errors"
	"runtime"
	"strings"
	"testing"
	"time"

	"github.com/gmh123521/java-dev-bootstrap/internal/detection"
	"github.com/gmh123521/java-dev-bootstrap/internal/model"
	"github.com/gmh123521/java-dev-bootstrap/internal/platform"
	"github.com/gmh123521/java-dev-bootstrap/internal/ports"
	"github.com/gmh123521/java-dev-bootstrap/internal/service"
)

func TestRunHelpContainsCommands(t *testing.T) {
	var output strings.Builder
	if err := Run(context.Background(), nil, &output, &output); err != nil {
		t.Fatalf("帮助命令失败: %v", err)
	}
	for _, command := range []string{"list", "plan", "install", "doctor", "prerequisites", "setup"} {
		if !strings.Contains(output.String(), command) {
			t.Fatalf("帮助缺少命令 %q: %s", command, output.String())
		}
	}
}

func TestFormatPrerequisiteShowsFailureSuggestion(t *testing.T) {
	item := platform.PrerequisiteItem{
		Name:       "包管理器 winget",
		Current:    "不可用",
		Suggestion: "请安装 App Installer",
	}
	actual := formatPrerequisite(item)
	for _, expected := range []string{"[失败]", "包管理器 winget", "不可用", "建议：请安装 App Installer"} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("前置条件输出缺少 %q: %s", expected, actual)
		}
	}
}

func TestFormatPlanItemShowsInstalledDetails(t *testing.T) {
	item := service.PlanItem{
		Package:   model.Package{Name: "JDK"},
		Command:   ports.Command{Program: "winget", Args: []string{"install", "jdk"}},
		Skipped:   true,
		Detection: detection.Result{Status: detection.StatusInstalled, Version: "23.0.2", Source: "java", Path: `D:\java\jdk\bin\java.exe`},
	}

	actual := formatPlanItem(item)
	for _, expected := range []string{"已安装 23.0.2", "来源 java", `D:\java\jdk\bin\java.exe`, "[跳过]"} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("计划输出缺少 %q: %s", expected, actual)
		}
	}
}

func TestFormatPlanItemShowsMissingPackage(t *testing.T) {
	item := service.PlanItem{
		Package:   model.Package{Name: "Gradle"},
		Command:   ports.Command{Program: "winget", Args: []string{"install", "gradle"}},
		Detection: detection.Result{Status: detection.StatusMissing, Source: "gradle"},
	}

	actual := formatPlanItem(item)
	if !strings.Contains(actual, "未检测到") || !strings.Contains(actual, "[待安装]") {
		t.Fatalf("缺失软件计划输出错误: %s", actual)
	}
}

func TestFormatDiagnosticShowsLevelAndSuggestion(t *testing.T) {
	warning := detection.Diagnostic{Level: detection.LevelWarning, Name: "JAVA_HOME", Current: "未设置", Suggestion: "指向 JDK 根目录"}
	actual := formatDiagnostic(warning)
	for _, expected := range []string{"[警告]", "JAVA_HOME", "未设置", "建议：指向 JDK 根目录"} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("诊断输出缺少 %q: %s", expected, actual)
		}
	}

	ok := formatDiagnostic(detection.Diagnostic{Level: detection.LevelOK, Name: "Java 版本", Current: "23.0.2"})
	if !strings.Contains(ok, "[正常]") {
		t.Fatalf("正常诊断输出错误: %s", ok)
	}
}

func TestManagerDiagnosticTurnsManagerFailureIntoWarning(t *testing.T) {
	item := managerDiagnostic("winget", ports.Result{Err: errors.New("命令不存在")})
	if item.Level != detection.LevelWarning || item.Name != "包管理器 winget" {
		t.Fatalf("包管理器失败应转换为警告: %#v", item)
	}
	if item.Suggestion == "" {
		t.Fatal("包管理器失败应提供修复建议")
	}
}

func TestOperationContextUsesSingleInstallDeadline(t *testing.T) {
	timeout := 30 * time.Minute
	ctx, cancel := operationContext(context.Background(), "install", false, timeout)
	defer cancel()
	deadline, ok := ctx.Deadline()
	if !ok {
		t.Fatal("实际安装必须有统一截止时间")
	}
	remaining := time.Until(deadline)
	if remaining <= 29*time.Minute || remaining > timeout {
		t.Fatalf("安装截止时间错误: %s", remaining)
	}

	dryRunCtx, dryRunCancel := operationContext(context.Background(), "install", true, timeout)
	defer dryRunCancel()
	if _, ok := dryRunCtx.Deadline(); ok {
		t.Fatal("纯 dry-run 不应增加安装超时")
	}

	planCtx, planCancel := operationContext(context.Background(), "plan", false, timeout)
	defer planCancel()
	planDeadline, ok := planCtx.Deadline()
	if !ok || time.Until(planDeadline) > 2*time.Minute {
		t.Fatal("只读计划应有不超过 2 分钟的统一检测超时")
	}
}

func TestRunHelpContainsDryRunOption(t *testing.T) {
	var output strings.Builder
	if err := Run(context.Background(), nil, &output, &output); err != nil {
		t.Fatalf("帮助命令失败: %v", err)
	}
	if !strings.Contains(output.String(), "--dry-run") {
		t.Fatalf("帮助缺少 --dry-run: %s", output.String())
	}
}

func TestRunHelpContainsJSONOption(t *testing.T) {
	var output strings.Builder
	if err := Run(context.Background(), nil, &output, &output); err != nil {
		t.Fatalf("帮助命令失败: %v", err)
	}
	if !strings.Contains(output.String(), "--json") {
		t.Fatalf("帮助缺少 --json: %s", output.String())
	}
}

func TestRunListJSONIsMachineReadable(t *testing.T) {
	var output strings.Builder
	if err := Run(context.Background(), []string{"list", "--json"}, &output, &output); err != nil {
		t.Fatalf("list JSON 命令失败: %v", err)
	}
	var value []map[string]any
	if err := json.Unmarshal([]byte(output.String()), &value); err != nil {
		t.Fatalf("list JSON 输出不可解析: %v; 输出: %s", err, output.String())
	}
	if len(value) == 0 {
		t.Fatal("list JSON 应至少返回一项软件清单")
	}
}

func TestRunVersionJSONIsMachineReadable(t *testing.T) {
	var output strings.Builder
	if err := Run(context.Background(), []string{"version", "--json"}, &output, &output); err != nil {
		t.Fatalf("version JSON 命令失败: %v", err)
	}
	var value map[string]string
	if err := json.Unmarshal([]byte(output.String()), &value); err != nil {
		t.Fatalf("version JSON 输出不可解析: %v; 输出: %s", err, output.String())
	}
	if value["version"] == "" {
		t.Fatal("version JSON 应包含非空版本号")
	}
}

func TestRunProfilesJSONIsMachineReadable(t *testing.T) {
	var output strings.Builder
	if err := Run(context.Background(), []string{"profiles", "--json"}, &output, &output); err != nil {
		t.Fatalf("profiles JSON 命令失败: %v", err)
	}
	var value []string
	if err := json.Unmarshal([]byte(output.String()), &value); err != nil {
		t.Fatalf("profiles JSON 输出不可解析: %v; 输出: %s", err, output.String())
	}
	if len(value) == 0 {
		t.Fatal("profiles JSON 应至少返回一个预设")
	}
}

func TestRunRejectsInvalidRetry(t *testing.T) {
	var output strings.Builder
	if err := Run(context.Background(), []string{"version", "--retry", "-1"}, &output, &output); err == nil {
		t.Fatal("负数重试次数应被拒绝")
	}
}

func TestRunRejectsOptionsForWrongCommand(t *testing.T) {
	tests := []struct {
		name string
		args []string
	}{
		{name: "version 不支持 yes", args: []string{"version", "--yes"}},
		{name: "list 不支持 dry-run", args: []string{"list", "--dry-run"}},
		{name: "plan 不支持 retry", args: []string{"plan", "--retry", "2"}},
		{name: "doctor 不支持 log", args: []string{"doctor", "--log", "doctor.log"}},
		{name: "setup 不支持 manifest", args: []string{"setup", "--manifest", "custom.yaml"}},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var output strings.Builder
			if err := Run(context.Background(), test.args, &output, &output); err == nil {
				t.Fatalf("参数应被拒绝: %v", test.args)
			}
		})
	}
}

func TestPackageDetectorIsEnabledForInstallDryRun(t *testing.T) {
	detector := packageDetectorFor("install", model.PlatformWindows, platform.ExecRunner{})
	if detector == nil {
		t.Fatal("install --dry-run 应执行只读软件检测")
	}
}

func TestRunDryRunDoesNotAskForConfirmation(t *testing.T) {
	if runtime.GOOS != "windows" && runtime.GOOS != "darwin" {
		t.Skip("dry-run 的平台执行测试只在产品支持的平台运行")
	}
	var output strings.Builder
	if err := Run(context.Background(), []string{"install", "--dry-run"}, &output, &output); err != nil {
		t.Fatalf("dry-run 不应失败: %v", err)
	}
	if strings.Contains(output.String(), "继续安装") {
		t.Fatalf("dry-run 不应要求确认: %s", output.String())
	}
}
