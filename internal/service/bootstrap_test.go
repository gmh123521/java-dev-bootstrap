package service

import (
	"context"
	"testing"

	"github.com/gmh123521/java-dev-bootstrap/internal/model"
	"github.com/gmh123521/java-dev-bootstrap/internal/ports"
)

type fakeRunner struct{ commands []ports.Command }

func (f *fakeRunner) Run(_ context.Context, command ports.Command) ports.Result {
	f.commands = append(f.commands, command)
	return ports.Result{Command: command, Err: nil}
}

func TestPlanMarksDetectedPackageAsSkipped(t *testing.T) {
	runner := &fakeRunner{}
	manifest := model.Manifest{Version: 1, Packages: []model.Package{{ID: "jdk", Name: "JDK", Platforms: []model.Platform{model.PlatformWindows}, Manager: "winget", ManagerID: "EclipseAdoptium.Temurin.21.JDK"}}}
	items, err := (Bootstrap{Runner: runner}).Plan(context.Background(), manifest, model.PlatformWindows)
	if err != nil {
		t.Fatalf("生成计划失败: %v", err)
	}
	if len(items) != 1 || !items[0].Skipped {
		t.Fatalf("已检测软件应跳过: %#v", items)
	}
}
