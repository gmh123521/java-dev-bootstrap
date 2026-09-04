package platform

import (
	"context"
	"errors"
	"testing"

	"github.com/gmh123521/java-dev-bootstrap/internal/model"
	"github.com/gmh123521/java-dev-bootstrap/internal/ports"
)

type prerequisiteRunner struct {
	result ports.Result
}

func (r prerequisiteRunner) Run(_ context.Context, command ports.Command) ports.Result {
	result := r.result
	result.Command = command
	return result
}

func TestCheckPrerequisitesReportsWindowsWinget(t *testing.T) {
	items, err := CheckPrerequisites(context.Background(), model.PlatformWindows, "amd64", prerequisiteRunner{result: ports.Result{Output: "v1.29.290"}})
	if err != nil {
		t.Fatalf("检查 Windows 前置条件失败: %v", err)
	}
	if len(items) != 2 || items[0].Name != "操作系统" || items[1].Name != "包管理器 winget" {
		t.Fatalf("前置条件项目错误: %#v", items)
	}
	if !items[1].OK || items[1].Current != "v1.29.290" {
		t.Fatalf("winget 正常状态错误: %#v", items[1])
	}
}

func TestCheckPrerequisitesReportsMissingHomebrew(t *testing.T) {
	items, err := CheckPrerequisites(context.Background(), model.PlatformDarwin, "arm64", prerequisiteRunner{result: ports.Result{Err: errors.New("命令不存在")}})
	if err != nil {
		t.Fatalf("检查 macOS 前置条件失败: %v", err)
	}
	if len(items) != 2 || items[1].Name != "包管理器 brew" {
		t.Fatalf("前置条件项目错误: %#v", items)
	}
	if items[1].OK || items[1].Suggestion == "" {
		t.Fatalf("brew 缺失时应提供修复建议: %#v", items[1])
	}
}

func TestCheckPrerequisitesRejectsUnsupportedPlatform(t *testing.T) {
	if _, err := CheckPrerequisites(context.Background(), model.Platform("linux"), "amd64", prerequisiteRunner{}); err == nil {
		t.Fatal("不支持的平台应返回错误")
	}
}
