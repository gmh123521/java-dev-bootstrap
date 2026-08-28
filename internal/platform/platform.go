package platform

import (
	"context"
	"fmt"
	"os/exec"
	"runtime"
	"strings"

	"github.com/gmh123521/java-dev-bootstrap/internal/model"
	"github.com/gmh123521/java-dev-bootstrap/internal/ports"
)

type ExecRunner struct{}

func (ExecRunner) Run(ctx context.Context, command ports.Command) ports.Result {
	cmd := exec.CommandContext(ctx, command.Program, command.Args...)
	output, err := cmd.CombinedOutput()
	return ports.Result{Command: command, Output: strings.TrimSpace(string(output)), Err: err}
}

func Current() (model.Platform, error) { return model.CurrentPlatform() }

func ManagerFor(platform model.Platform) (string, error) {
	switch platform {
	case model.PlatformWindows:
		return "winget", nil
	case model.PlatformDarwin:
		return "brew", nil
	default:
		return "", fmt.Errorf("不支持的平台: %s（运行时为 %s）", platform, runtime.GOOS)
	}
}

func InstallCommand(platform model.Platform, pkg model.Package) (ports.Command, error) {
	if pkg.Manager == "" || pkg.ManagerID == "" {
		return ports.Command{}, fmt.Errorf("软件包 %q 缺少安装器信息", pkg.ID)
	}
	switch platform {
	case model.PlatformWindows:
		return ports.Command{Program: "winget", Args: []string{"install", "--exact", "--id", pkg.ManagerIDFor(platform), "--accept-source-agreements", "--accept-package-agreements"}}, nil
	case model.PlatformDarwin:
		if pkg.Kind == "formula" {
			return ports.Command{Program: "brew", Args: []string{"install", pkg.ManagerIDFor(platform)}}, nil
		}
		return ports.Command{Program: "brew", Args: []string{"install", "--cask", pkg.ManagerIDFor(platform)}}, nil
	default:
		return ports.Command{}, fmt.Errorf("暂不支持平台: %s", platform)
	}
}

func CheckCommand(platform model.Platform, pkg model.Package) (ports.Command, error) {
	switch platform {
	case model.PlatformWindows:
		return ports.Command{Program: "winget", Args: []string{"list", "--exact", "--id", pkg.ManagerIDFor(platform)}}, nil
	case model.PlatformDarwin:
		if pkg.Kind == "formula" {
			return ports.Command{Program: "brew", Args: []string{"list", "--formula", pkg.ManagerIDFor(platform)}}, nil
		}
		return ports.Command{Program: "brew", Args: []string{"list", "--cask", pkg.ManagerIDFor(platform)}}, nil
	default:
		return ports.Command{}, fmt.Errorf("暂不支持平台: %s", platform)
	}
}
