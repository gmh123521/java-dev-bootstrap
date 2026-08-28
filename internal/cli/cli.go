package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/gmh123521/java-dev-bootstrap/internal/config"
	"github.com/gmh123521/java-dev-bootstrap/internal/logging"
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
	dryRun := false
	profile := ""
	logPath := "jdb.log"
	timeout := 30 * time.Minute
	command := args[0]
	for i := 1; i < len(args); i++ {
		switch args[i] {
		case "--yes":
			yes = true
		case "--dry-run":
			dryRun = true
		case "--profile":
			if i+1 >= len(args) {
				return fmt.Errorf("--profile 后必须提供名称")
			}
			profile = args[i+1]
			i++
		case "--log":
			if i+1 >= len(args) {
				return fmt.Errorf("--log 后必须提供文件路径")
			}
			logPath = args[i+1]
			i++
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
	if profile != "" {
		manifest, err = service.ApplyProfile(manifest, profile)
		if err != nil {
			return err
		}
	}
	current, err := platform.Current()
	if err != nil {
		return err
	}
	switch command {
	case "profiles":
		for _, name := range service.ProfileNames() {
			fmt.Fprintln(out, "-", name)
		}
		return nil
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
		if command == "install" && !dryRun {
			logFile, openErr := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
			if openErr != nil {
				return fmt.Errorf("打开日志文件失败: %w", openErr)
			}
			defer logFile.Close()
			runner = logging.Runner{Inner: platform.ExecRunner{}, Logger: logging.New(logFile)}
		}
		bootstrap := service.Bootstrap{Runner: runner, Timeout: timeout}
		preflightCtx, preflightCancel := context.WithTimeout(ctx, 30*time.Second)
		defer preflightCancel()
		if command == "install" && !dryRun {
			managerCheck, checkErr := platform.ManagerCheckCommand(current)
			if checkErr != nil {
				return checkErr
			}
			managerResult := runner.Run(preflightCtx, managerCheck)
			if managerResult.Err != nil {
				return fmt.Errorf("包管理器不可用，请先安装或修复：%w", managerResult.Err)
			}
		}
		items, err := bootstrap.Plan(preflightCtx, manifest, current)
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
		if command == "plan" || dryRun || pending == 0 {
			return nil
		}
		if !yes {
			fmt.Fprint(out, "\n继续安装？请输入 yes 确认：")
			var answer string
			if _, err := fmt.Fscan(os.Stdin, &answer); err != nil || strings.ToLower(answer) != "yes" {
				return fmt.Errorf("未确认安装")
			}
		}
		report := bootstrap.InstallReport(ctx, items)
		fmt.Fprintf(out, "\n安装汇总：成功 %d，跳过 %d，失败 %d\n", report.Succeeded, report.Skipped, report.Failed)
		if report.Failed > 0 {
			for _, installErr := range report.Errors {
				fmt.Fprintln(errOut, "-", installErr)
			}
			return fmt.Errorf("有 %d 个软件安装失败", report.Failed)
		}
		return nil
	case "doctor":
		manager, err := platform.ManagerFor(current)
		if err != nil {
			return err
		}
		check, err := platform.ManagerCheckCommand(current)
		if err != nil {
			return err
		}
		doctorCtx, cancel := context.WithTimeout(ctx, 15*time.Second)
		defer cancel()
		result := (platform.ExecRunner{}).Run(doctorCtx, check)
		if result.Err != nil {
			return fmt.Errorf("包管理器 %s 不可用，请先安装或修复：%w", manager, result.Err)
		}
		fmt.Fprintf(out, "操作系统：%s\n清单：%s\n包管理器：%s（%s）\n", current, manifestPath, manager, result.Output)
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
	return "Java Dev Bootstrap\n\n用法：jdb <list|profiles|plan|install|doctor> [--manifest 路径] [--profile 名称] [--yes] [--dry-run] [--log 路径]\n"
}
