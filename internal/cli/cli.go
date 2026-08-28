package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/gmh123521/java-dev-bootstrap/internal/config"
	"github.com/gmh123521/java-dev-bootstrap/internal/model"
	"github.com/gmh123521/java-dev-bootstrap/internal/platform"
	"github.com/gmh123521/java-dev-bootstrap/internal/ports"
	"github.com/gmh123521/java-dev-bootstrap/internal/service"
)

const DefaultManifest = "configs/default.yaml"

func Run(ctx context.Context, args []string, out, errOut io.Writer) error {
	if len(args) == 0 {
		return help(out)
	}
	manifestPath := DefaultManifest
	yes := false
	command := args[0]
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--yes":
			yes = true
		case "--manifest":
			if i+1 >= len(args) {
				return fmt.Errorf("--manifest 后必须提供文件路径")
			}
			manifestPath = args[i+1]
			i++
		case "--help", "-h":
			return help(out)
		default:
			return fmt.Errorf("未知参数: %s", args[i])
		}
	}
	var manifest model.Manifest
	var err error
	if manifestPath == DefaultManifest {
		manifest, err = config.LoadDefaultManifest()
	} else {
		manifest, err = config.LoadManifest(manifestPath)
	}
	if err != nil {
		return err
	}
	current, err := platform.Current()
	if err != nil {
		return err
	}
	switch command {
	case "list":
		packages, err := manifest.PackagesForPlatform(current)
		if err != nil {
			return err
		}
		for _, pkg := range packages {
			fmt.Fprintf(out, "- %s（%s）：%s\n", pkg.Name, pkg.ID, pkg.Description)
		}
		return nil
	case "plan", "install":
		var runner ports.Runner
		if command == "install" {
			runner = platform.ExecRunner{}
		}
		bootstrap := service.Bootstrap{Runner: runner}
		items, err := bootstrap.Plan(ctx, manifest, current)
		if err != nil {
			return err
		}
		pending := 0
		for _, item := range items {
			status := "待安装"
			if item.Skipped {
				status = "已安装，跳过"
			} else {
				pending++
			}
			fmt.Fprintf(out, "- %s：%s [%s]\n", item.Package.Name, formatCommand(item.Command), status)
		}
		if command == "plan" || pending == 0 {
			return nil
		}
		if !yes {
			fmt.Fprint(out, "\n继续安装？请输入 yes 确认：")
			var answer string
			if _, err := fmt.Fscan(os.Stdin, &answer); err != nil || strings.ToLower(answer) != "yes" {
				return fmt.Errorf("未确认安装")
			}
		}
		return bootstrap.Install(ctx, items)
	case "doctor":
		if _, err := platform.ManagerFor(current); err != nil {
			return err
		}
		fmt.Fprintf(out, "操作系统：%s\n清单：%s\n包管理器：可用性将在安装时检查\n", current, manifestPath)
		return nil
	default:
		return fmt.Errorf("未知命令: %s\n%s", command, helpText())
	}
}

func formatCommand(command ports.Command) string {
	parts := append([]string{command.Program}, command.Args...)
	return strings.Join(parts, " ")
}

func help(out io.Writer) error {
	_, err := io.WriteString(out, helpText())
	return err
}

func helpText() string {
	return "Java Dev Bootstrap\n\n用法：jdb <list|plan|install|doctor> [--manifest 路径] [--yes]\n"
}
