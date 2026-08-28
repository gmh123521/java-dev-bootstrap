package platform

import (
	"testing"

	"github.com/gmh123521/java-dev-bootstrap/internal/model"
)

func TestManagerCheckCommandUsesVersion(t *testing.T) {
	command, err := ManagerCheckCommand(model.PlatformWindows)
	if err != nil {
		t.Fatalf("生成包管理器检查命令失败: %v", err)
	}
	if command.Program != "winget" || len(command.Args) != 1 || command.Args[0] != "--version" {
		t.Fatalf("检查命令错误: %#v", command)
	}
}
