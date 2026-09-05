package cli

import (
	"strings"
	"testing"

	"github.com/gmh123521/java-dev-bootstrap/internal/detection"
	"github.com/gmh123521/java-dev-bootstrap/internal/model"
	"github.com/gmh123521/java-dev-bootstrap/internal/platform"
	"github.com/gmh123521/java-dev-bootstrap/internal/ports"
	"github.com/gmh123521/java-dev-bootstrap/internal/service"
)

func TestFormatPackagesJSONIncludesPackageFields(t *testing.T) {
	actual, err := formatPackagesJSON([]model.Package{{ID: "jdk", Name: "JDK", Description: "Java 开发工具包"}})
	if err != nil {
		t.Fatalf("生成软件清单 JSON 失败: %v", err)
	}
	for _, expected := range []string{`"id": "jdk"`, `"name": "JDK"`, `"description": "Java 开发工具包"`} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("软件清单 JSON 缺少 %q: %s", expected, actual)
		}
	}
}

func TestFormatPlanJSONIncludesDetectionAndCommand(t *testing.T) {
	actual, err := formatPlanJSON([]service.PlanItem{{
		Package:   model.Package{ID: "gradle", Name: "Gradle"},
		Command:   ports.Command{Program: "winget", Args: []string{"install", "gradle"}},
		Detection: detection.Result{Status: detection.StatusMissing, Source: "winget"},
	}})
	if err != nil {
		t.Fatalf("生成安装计划 JSON 失败: %v", err)
	}
	for _, expected := range []string{"\"id\": \"gradle\"", "\"status\": \"missing\"", "\"program\": \"winget\"", "\"skipped\": false"} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("安装计划 JSON 缺少 %q: %s", expected, actual)
		}
	}
}

func TestFormatInstallReportJSONIncludesVerification(t *testing.T) {
	actual, err := formatInstallReportJSON(service.InstallReport{Succeeded: 1, Skipped: 2, Failed: 0, Retried: 1, Verified: 1, VerificationFailed: 0})
	if err != nil {
		t.Fatalf("生成安装报告 JSON 失败: %v", err)
	}
	for _, expected := range []string{"\"succeeded\": 1", "\"skipped\": 2", "\"retried\": 1", "\"verified\": 1", "\"verification_failed\": 0"} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("安装报告 JSON 缺少 %q: %s", expected, actual)
		}
	}
}

func TestFormatDiagnosticsJSONIncludesSuggestions(t *testing.T) {
	actual, err := formatDiagnosticsJSON([]detection.Diagnostic{{Level: detection.LevelWarning, Name: "JAVA_HOME", Current: "未设置", Suggestion: "设置 JDK 路径"}})
	if err != nil {
		t.Fatalf("生成诊断 JSON 失败: %v", err)
	}
	for _, expected := range []string{"\"level\": \"warning\"", "\"name\": \"JAVA_HOME\"", "\"suggestion\": \"设置 JDK 路径\""} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("诊断 JSON 缺少 %q: %s", expected, actual)
		}
	}
}

func TestFormatPrerequisitesJSONIncludesReadiness(t *testing.T) {
	actual, err := formatPrerequisitesJSON([]platform.PrerequisiteItem{{Name: "包管理器 winget", Current: "v1.29.290", OK: true}})
	if err != nil {
		t.Fatalf("生成前置条件 JSON 失败: %v", err)
	}
	for _, expected := range []string{"\"name\": \"包管理器 winget\"", "\"current\": \"v1.29.290\"", "\"ok\": true"} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("前置条件 JSON 缺少 %q: %s", expected, actual)
		}
	}
}

func TestFormatSetupJSONIncludesNextCommand(t *testing.T) {
	actual, err := formatSetupJSON(platform.SetupResult{Ready: false, Platform: "windows（amd64）", Manager: "winget", ManagerState: "不可用", Suggestion: "安装 winget", NextCommand: "jdb prerequisites"})
	if err != nil {
		t.Fatalf("生成 setup JSON 失败: %v", err)
	}
	for _, expected := range []string{"\"ready\": false", "\"manager\": \"winget\"", "\"suggestion\": \"安装 winget\"", "\"next_command\": \"jdb prerequisites\""} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("setup JSON 缺少 %q: %s", expected, actual)
		}
	}
}
