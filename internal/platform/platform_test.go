package platform

import (
	"testing"

	"github.com/gmh123521/java-dev-bootstrap/internal/model"
)

func TestInstallCommandUsesHomebrewFormulaForGit(t *testing.T) {
	pkg := model.Package{ID: "git", Manager: "winget", ManagerID: "Git.Git", DarwinID: "git", Kind: "formula"}
	command, err := InstallCommand(model.PlatformDarwin, pkg)
	if err != nil {
		t.Fatalf("生成 macOS 安装命令失败: %v", err)
	}
	if command.Program != "brew" || len(command.Args) != 2 || command.Args[0] != "install" || command.Args[1] != "git" {
		t.Fatalf("命令错误: %#v", command)
	}
}

func TestInstallCommandUsesHomebrewCaskForJDK(t *testing.T) {
	pkg := model.Package{ID: "jdk", Manager: "winget", ManagerID: "EclipseAdoptium.Temurin.21.JDK", DarwinID: "temurin@21", Kind: "cask"}
	command, err := InstallCommand(model.PlatformDarwin, pkg)
	if err != nil {
		t.Fatalf("生成 macOS cask 命令失败: %v", err)
	}
	if command.Program != "brew" || len(command.Args) != 3 || command.Args[0] != "install" || command.Args[1] != "--cask" || command.Args[2] != "temurin@21" {
		t.Fatalf("命令错误: %#v", command)
	}
}
