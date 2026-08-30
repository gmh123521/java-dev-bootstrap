package cli

import (
	"context"
	"runtime"
	"strings"
	"testing"

	"github.com/gmh123521/java-dev-bootstrap/internal/detection"
	"github.com/gmh123521/java-dev-bootstrap/internal/model"
	"github.com/gmh123521/java-dev-bootstrap/internal/ports"
	"github.com/gmh123521/java-dev-bootstrap/internal/service"
)

func TestRunHelpContainsCommands(t *testing.T) {
	var output strings.Builder
	if err := Run(context.Background(), nil, &output, &output); err != nil {
		t.Fatalf("帮助命令失败: %v", err)
	}
	for _, command := range []string{"list", "plan", "install", "doctor"} {
		if !strings.Contains(output.String(), command) {
			t.Fatalf("帮助缺少命令 %q: %s", command, output.String())
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

func TestRunHelpContainsDryRunOption(t *testing.T) {
	var output strings.Builder
	if err := Run(context.Background(), nil, &output, &output); err != nil {
		t.Fatalf("帮助命令失败: %v", err)
	}
	if !strings.Contains(output.String(), "--dry-run") {
		t.Fatalf("帮助缺少 --dry-run: %s", output.String())
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
