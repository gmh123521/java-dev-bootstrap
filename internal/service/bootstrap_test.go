package service

import (
	"context"
	"testing"

	"github.com/gmh123521/java-dev-bootstrap/internal/detection"
	"github.com/gmh123521/java-dev-bootstrap/internal/model"
	"github.com/gmh123521/java-dev-bootstrap/internal/ports"
)

type fakeRunner struct{ commands []ports.Command }

func (f *fakeRunner) Run(_ context.Context, command ports.Command) ports.Result {
	f.commands = append(f.commands, command)
	return ports.Result{Command: command, Err: nil}
}

type fakeDetector struct {
	result detection.Result
}

func (d fakeDetector) Detect(_ context.Context, _ model.Package) detection.Result {
	return d.result
}

func TestPlanMarksDetectedPackageAsSkipped(t *testing.T) {
	manifest := model.Manifest{Version: 1, Packages: []model.Package{{ID: "jdk", Name: "JDK", Platforms: []model.Platform{model.PlatformWindows}, Manager: "winget", ManagerID: "EclipseAdoptium.Temurin.21.JDK"}}}
	detector := fakeDetector{result: detection.Result{Status: detection.StatusInstalled, Version: "23.0.2", Source: "java"}}
	items, err := (Bootstrap{Detector: detector}).Plan(context.Background(), manifest, model.PlatformWindows)
	if err != nil {
		t.Fatalf("生成计划失败: %v", err)
	}
	if len(items) != 1 || !items[0].Skipped || items[0].Detection.Version != "23.0.2" {
		t.Fatalf("已检测软件应跳过: %#v", items)
	}
}

func TestPlanKeepsMissingPackagePending(t *testing.T) {
	manifest := model.Manifest{Version: 1, Packages: []model.Package{{ID: "gradle", Name: "Gradle", Platforms: []model.Platform{model.PlatformWindows}, Manager: "winget", ManagerID: "Gradle.Gradle"}}}
	detector := fakeDetector{result: detection.Result{Status: detection.StatusMissing, Source: "gradle"}}

	items, err := (Bootstrap{Detector: detector}).Plan(context.Background(), manifest, model.PlatformWindows)
	if err != nil {
		t.Fatalf("生成计划失败: %v", err)
	}
	if len(items) != 1 || items[0].Skipped || items[0].Detection.Status != detection.StatusMissing {
		t.Fatalf("未安装软件应保留待安装状态: %#v", items)
	}
}
