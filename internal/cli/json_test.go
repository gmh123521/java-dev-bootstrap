package cli

import (
	"strings"
	"testing"

	"github.com/gmh123521/java-dev-bootstrap/internal/detection"
	"github.com/gmh123521/java-dev-bootstrap/internal/model"
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
	actual, err := formatInstallReportJSON(service.InstallReport{Succeeded: 1, Skipped: 2, Failed: 0, Verified: 1, VerificationFailed: 0})
	if err != nil {
		t.Fatalf("生成安装报告 JSON 失败: %v", err)
	}
	for _, expected := range []string{"\"succeeded\": 1", "\"skipped\": 2", "\"verified\": 1", "\"verification_failed\": 0"} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("安装报告 JSON 缺少 %q: %s", expected, actual)
		}
	}
}
