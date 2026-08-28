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

func ManagerCheckCommand(platform model.Platform) (ports.Command, error) {
	manager, err := ManagerFor(platform)
	if err != nil {
		return ports.Command{}, err
	}
	return ports.Command{Program: manager, Args: []string{"--version"}}, nil
}

func InstallCommand(platform model.Platform, pkg model.Package) (ports.Command, error) {
	manager, managerID, err := packageManager(platform, pkg)
	if err != nil {
		return ports.Command{}, err
	}
	switch manager {
	case "winget":
		return ports.Command{Program: "winget", Args: []string{"install", "--exact", "--id", managerID, "--accept-source-agreements", "--accept-package-agreements"}}, nil
	case "brew":
		if pkg.Kind == "formula" {
			return ports.Command{Program: "brew", Args: []string{"install", managerID}}, nil
		}
		return ports.Command{Program: "brew", Args: []string{"install", "--cask", managerID}}, nil
	default:
		return ports.Command{}, fmt.Errorf("软件包 %q 使用了不支持的安装器: %s", pkg.ID, manager)
	}
}

func packageManager(platform model.Platform, pkg model.Package) (string, string, error) {
	manager := strings.ToLower(strings.TrimSpace(pkg.Manager))
	managerID := strings.TrimSpace(pkg.ManagerIDFor(platform))
	if platform == model.PlatformDarwin {
		if strings.TrimSpace(pkg.DarwinID) == "" {
			return "", "", fmt.Errorf("软件包 %q 缺少 macOS 安装器 ID", pkg.ID)
		}
		if manager != "" && manager != "brew" && manager != "winget" {
			return "", "", fmt.Errorf("软件包 %q 的 macOS 安装器不受支持: %s", pkg.ID, manager)
		}
		manager = "brew"
	}
	if platform == model.PlatformWindows && manager != "winget" {
		return "", "", fmt.Errorf("软件包 %q 的 Windows 安装器必须是 winget", pkg.ID)
	}
	if manager == "" || managerID == "" {
		return "", "", fmt.Errorf("软件包 %q 缺少当前平台的安装器或安装器 ID", pkg.ID)
	}
	if manager == "brew" && pkg.Kind != "formula" && pkg.Kind != "cask" {
		return "", "", fmt.Errorf("软件包 %q 的 Homebrew 类型必须是 formula 或 cask", pkg.ID)
	}
	return manager, managerID, nil
}

func CheckCommand(platform model.Platform, pkg model.Package) (ports.Command, error) {
	manager, managerID, err := packageManager(platform, pkg)
	if err != nil {
		return ports.Command{}, err
	}
	switch manager {
	case "winget":
		return ports.Command{Program: "winget", Args: []string{"list", "--exact", "--id", managerID}}, nil
	case "brew":
		if pkg.Kind == "formula" {
			return ports.Command{Program: "brew", Args: []string{"list", "--formula", managerID}}, nil
		}
		return ports.Command{Program: "brew", Args: []string{"list", "--cask", managerID}}, nil
	default:
		return ports.Command{}, fmt.Errorf("软件包 %q 使用了不支持的安装器: %s", pkg.ID, manager)
	}
}
