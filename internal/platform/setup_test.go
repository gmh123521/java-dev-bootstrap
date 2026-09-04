package platform

import (
	"context"
	"errors"
	"testing"

	"github.com/gmh123521/java-dev-bootstrap/internal/model"
	"github.com/gmh123521/java-dev-bootstrap/internal/ports"
)

type setupRunner struct {
	result ports.Result
}

func (r setupRunner) Run(_ context.Context, command ports.Command) ports.Result {
	result := r.result
	result.Command = command
	return result
}

func TestSetupGuideForReadyWindows(t *testing.T) {
	guide, err := SetupGuide(context.Background(), model.PlatformWindows, "amd64", setupRunner{result: ports.Result{Output: "v1.29.290"}})
	if err != nil {
		t.Fatalf("生成 Windows setup 指引失败: %v", err)
	}
	if !guide.Ready || guide.NextCommand != "jdb doctor" {
		t.Fatalf("winget 正常时应可继续诊断: %#v", guide)
	}
	if guide.Suggestion != "" {
		t.Fatalf("前置条件正常时不应显示修复建议: %#v", guide)
	}
}

func TestSetupGuideForMissingHomebrew(t *testing.T) {
	guide, err := SetupGuide(context.Background(), model.PlatformDarwin, "arm64", setupRunner{result: ports.Result{Err: errors.New("命令不存在")}})
	if err != nil {
		t.Fatalf("生成 macOS setup 指引失败: %v", err)
	}
	if guide.Ready || guide.Suggestion == "" || guide.NextCommand != "jdb prerequisites" {
		t.Fatalf("brew 缺失时应给出处理建议: %#v", guide)
	}
}

func TestSetupGuideRejectsUnsupportedPlatform(t *testing.T) {
	if _, err := SetupGuide(context.Background(), model.Platform("linux"), "amd64", setupRunner{}); err == nil {
		t.Fatal("不支持的平台应返回错误")
	}
}
