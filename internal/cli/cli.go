package cli

import (
	"context"
	"fmt"
	"io"
	"os"
	"runtime"
	"strconv"
	"strings"
	"time"

	"github.com/gmh123521/java-dev-bootstrap/internal/config"
	"github.com/gmh123521/java-dev-bootstrap/internal/detection"
	"github.com/gmh123521/java-dev-bootstrap/internal/logging"
	"github.com/gmh123521/java-dev-bootstrap/internal/model"
	"github.com/gmh123521/java-dev-bootstrap/internal/platform"
	"github.com/gmh123521/java-dev-bootstrap/internal/ports"
	"github.com/gmh123521/java-dev-bootstrap/internal/service"
	"github.com/gmh123521/java-dev-bootstrap/internal/version"
)

const DefaultManifest = "configs/default.yaml"

func Run(ctx context.Context, args []string, out, errOut io.Writer) error {
	if len(args) == 0 {
		return help(out)
	}
	manifestPath := DefaultManifest
	yes := false
	dryRun := false
	jsonOutput := false
	retry := 0
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
		case "--json":
			jsonOutput = true
		case "--retry":
			if i+1 >= len(args) {
				return fmt.Errorf("--retry 后必须提供非负整数")
			}
			parsedRetry, parseErr := strconv.Atoi(args[i+1])
			if parseErr != nil || parsedRetry < 0 {
				return fmt.Errorf("--retry 后必须提供非负整数")
			}
			retry = parsedRetry
			i++
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
	// 版本和 profile 列表不依赖操作系统，允许在 Linux CI 等环境中执行。
	if command == "version" {
		fmt.Fprintf(out, "版本：%s\n", version.Display())
		return nil
	}
	if command == "profiles" {
		for _, name := range service.ProfileNames() {
			fmt.Fprintln(out, "-", name)
		}
		return nil
	}
	if command == "prerequisites" {
		current, err := platform.Current()
		if err != nil {
			return err
		}
		items, err := platform.CheckPrerequisites(ctx, current, runtime.GOARCH, platform.ExecRunner{})
		if err != nil {
			return err
		}
		fmt.Fprintln(out, "前置条件：")
		ready := true
		for _, item := range items {
			fmt.Fprintln(out, formatPrerequisite(item))
			ready = ready && item.OK
		}
		if !ready {
			return fmt.Errorf("前置条件不满足，请按提示处理后重试")
		}
		return nil
	}
	if command == "setup" {
		current, err := platform.Current()
		if err != nil {
			return err
		}
		guide, err := platform.SetupGuide(ctx, current, runtime.GOARCH, platform.ExecRunner{})
		if err != nil {
			return err
		}
		fmt.Fprintln(out, formatSetupGuide(guide))
		if !guide.Ready {
			return fmt.Errorf("前置条件不满足，请按提示处理后重试")
		}
		return nil
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
	case "list":
		packages, err := manifest.PackagesForPlatform(current)
		if err != nil {
			return err
		}
		if jsonOutput {
			formatted, formatErr := formatPackagesJSON(packages)
			if formatErr != nil {
				return formatErr
			}
			fmt.Fprintln(out, formatted)
			return nil
		}
		for _, pkg := range packages {
			fmt.Fprintf(out, "- %s（%s）：%s\n", pkg.Name, pkg.ID, pkg.Description)
		}
		return nil
	case "plan", "install":
		execRunner := platform.ExecRunner{}
		var runner ports.Runner
		var detectionRunner ports.Runner = execRunner
		var packageDetector detection.Detector
		if command == "install" && !dryRun {
			logFile, openErr := os.OpenFile(logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o600)
			if openErr != nil {
				return fmt.Errorf("打开日志文件失败: %w", openErr)
			}
			defer logFile.Close()
			runner = logging.Runner{Inner: platform.ExecRunner{}, Logger: logging.New(logFile)}
			detectionRunner = runner
		}
		packageDetector = packageDetectorFor(command, current, detectionRunner)
		bootstrap := service.Bootstrap{
			Runner:                runner,
			Detector:              packageDetector,
			Timeout:               timeout,
			Retry:                 retry,
			IgnoreDetectionErrors: command == "install" && dryRun,
		}
		operationCtx, operationCancel := operationContext(ctx, command, dryRun, timeout)
		defer operationCancel()
		preflightCtx, preflightCancel := context.WithTimeout(operationCtx, 30*time.Second)
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
		items, err := bootstrap.Plan(operationCtx, manifest, current)
		if err != nil {
			return err
		}
		pending := 0
		for _, item := range items {
			if !item.Skipped {
				pending++
			}
			if !jsonOutput {
				fmt.Fprintln(out, formatPlanItem(item))
			}
		}
		if command == "plan" || dryRun || pending == 0 {
			if jsonOutput {
				formatted, formatErr := formatPlanJSON(items)
				if formatErr != nil {
					return formatErr
				}
				fmt.Fprintln(out, formatted)
			}
			return nil
		}
		if jsonOutput && !yes {
			return fmt.Errorf("JSON 模式执行真实安装必须同时使用 --yes")
		}
		if !yes {
			fmt.Fprint(out, "\n继续安装？请输入 yes 确认：")
			var answer string
			if _, err := fmt.Fscan(os.Stdin, &answer); err != nil || strings.ToLower(answer) != "yes" {
				return fmt.Errorf("未确认安装")
			}
		}
		report := bootstrap.InstallReport(operationCtx, items)
		if jsonOutput {
			formatted, formatErr := formatInstallReportJSON(report)
			if formatErr != nil {
				return formatErr
			}
			fmt.Fprintln(out, formatted)
		} else {
			fmt.Fprintf(out, "\n安装汇总：成功 %d，跳过 %d，失败 %d，重试 %d\n", report.Succeeded, report.Skipped, report.Failed, report.Retried)
		}
		if packageDetector != nil {
			fmt.Fprintf(out, "安装后复查：通过 %d，失败 %d\n", report.Verified, report.VerificationFailed)
		}
		if report.Failed > 0 {
			for _, installErr := range report.Errors {
				fmt.Fprintln(errOut, "-", installErr)
			}
			return fmt.Errorf("有 %d 个软件安装失败", report.Failed)
		}
		if report.VerificationFailed > 0 {
			for _, verificationErr := range report.VerificationErrors {
				fmt.Fprintln(errOut, "-", verificationErr)
			}
			return fmt.Errorf("有 %d 个软件安装后复查失败", report.VerificationFailed)
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
		fmt.Fprintf(out, "操作系统：%s\n清单：%s\n", current, manifestPath)
		fmt.Fprintln(out, formatDiagnostic(managerDiagnostic(manager, result)))
		fmt.Fprintln(out, "环境诊断：")
		environmentDetector := detection.EnvironmentDetector{Runner: platform.ExecRunner{}}
		diagnostics := environmentDetector.Diagnose(doctorCtx, current)
		diagnostics = append(diagnostics, environmentDetector.DiagnoseTools(doctorCtx, manifest.Packages)...)
		applicationDetector := detection.PackageDetector{Runner: platform.ExecRunner{}, Platform: current}
		diagnostics = append(diagnostics, environmentDetector.DiagnoseApplications(doctorCtx, manifest.Packages, applicationDetector)...)
		for _, diagnostic := range diagnostics {
			fmt.Fprintln(out, formatDiagnostic(diagnostic))
		}
		return nil
	default:
		return fmt.Errorf("未知命令: %s\n%s", command, helpText())
	}
}

func formatCommand(command ports.Command) string {
	parts := append([]string{command.Program}, command.Args...)
	return strings.Join(parts, " ")
}

func formatPlanItem(item service.PlanItem) string {
	detail := "未检测到"
	action := "待安装"
	switch item.Detection.Status {
	case detection.StatusInstalled:
		detail = "已安装"
		if item.Detection.Version != "" {
			detail += " " + item.Detection.Version
		}
		action = "跳过"
	case detection.StatusOutdated:
		detail = "版本过低"
		if item.Detection.Version != "" {
			detail += " " + item.Detection.Version
		}
		action = "待升级"
	case detection.StatusMissing:
		detail = "未检测到"
	case detection.StatusError:
		detail = "检测失败"
	}
	if item.Detection.Source != "" {
		detail += "，来源 " + item.Detection.Source
	}
	if item.Detection.Path != "" {
		detail += "，路径 " + item.Detection.Path
	}
	if !item.Skipped {
		detail += "；执行 " + formatCommand(item.Command)
	}
	return fmt.Sprintf("- %s：%s [%s]", item.Package.Name, detail, action)
}

func formatDiagnostic(item detection.Diagnostic) string {
	level := "正常"
	if item.Level == detection.LevelWarning {
		level = "警告"
	}
	result := fmt.Sprintf("- [%s] %s：%s", level, item.Name, item.Current)
	if item.Suggestion != "" {
		result += "；建议：" + item.Suggestion
	}
	return result
}

func formatPrerequisite(item platform.PrerequisiteItem) string {
	level := "失败"
	if item.OK {
		level = "正常"
	}
	result := fmt.Sprintf("- [%s] %s：%s", level, item.Name, item.Current)
	if item.Suggestion != "" {
		result += "；建议：" + item.Suggestion
	}
	return result
}

func packageDetectorFor(command string, current model.Platform, runner ports.Runner) detection.Detector {
	if (command != "plan" && command != "install") || runner == nil {
		return nil
	}
	return detection.PackageDetector{Runner: runner, Platform: current}
}

func formatSetupGuide(guide platform.SetupResult) string {
	var builder strings.Builder
	builder.WriteString("准备检查：\n")
	fmt.Fprintf(&builder, "- 平台：%s\n", guide.Platform)
	fmt.Fprintf(&builder, "- 包管理器 %s：%s\n", guide.Manager, guide.ManagerState)
	if guide.Ready {
		fmt.Fprintf(&builder, "- 下一步：%s", guide.NextCommand)
	} else {
		fmt.Fprintf(&builder, "- 建议：%s\n", guide.Suggestion)
		fmt.Fprintf(&builder, "- 处理后执行：%s", guide.NextCommand)
	}
	return builder.String()
}

func managerDiagnostic(manager string, result ports.Result) detection.Diagnostic {
	if result.Err != nil {
		return detection.Diagnostic{
			Level:      detection.LevelWarning,
			Name:       "包管理器 " + manager,
			Current:    "不可用",
			Suggestion: "先安装或修复 " + manager + "，再执行软件安装",
		}
	}
	return detection.Diagnostic{Level: detection.LevelOK, Name: "包管理器 " + manager, Current: result.Output}
}

func operationContext(ctx context.Context, command string, dryRun bool, timeout time.Duration) (context.Context, context.CancelFunc) {
	if command == "install" && !dryRun {
		return context.WithTimeout(ctx, timeout)
	}
	if command == "plan" {
		return context.WithTimeout(ctx, 2*time.Minute)
	}
	return context.WithCancel(ctx)
}

func help(out io.Writer) error {
	_, err := io.WriteString(out, helpText())
	return err
}

func helpText() string {
	return "Java Dev Bootstrap\n\n用法：jdb <version|list|profiles|prerequisites|setup|plan|install|doctor> [--manifest 路径] [--profile 名称] [--yes] [--dry-run] [--json] [--retry 次数] [--log 路径]\n"
}
