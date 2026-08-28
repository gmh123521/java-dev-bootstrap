package service

import (
	"context"
	"fmt"
	"time"

	"github.com/gmh123521/java-dev-bootstrap/internal/model"
	platformexec "github.com/gmh123521/java-dev-bootstrap/internal/platform"
	"github.com/gmh123521/java-dev-bootstrap/internal/ports"
)

type PlanItem struct {
	Package model.Package
	Command ports.Command
	Skipped bool
}

type InstallReport struct {
	Succeeded int
	Skipped   int
	Failed    int
	Errors    []error
}

type Bootstrap struct {
	Runner  ports.Runner
	Timeout time.Duration
}

func (b Bootstrap) commandContext(ctx context.Context) (context.Context, context.CancelFunc) {
	if b.Timeout <= 0 {
		return context.WithCancel(ctx)
	}
	return context.WithTimeout(ctx, b.Timeout)
}

func (b Bootstrap) Plan(ctx context.Context, manifest model.Manifest, platform model.Platform) ([]PlanItem, error) {
	packages, err := manifest.PackagesForPlatform(platform)
	if err != nil {
		return nil, err
	}
	items := make([]PlanItem, 0, len(packages))
	for _, pkg := range packages {
		command, err := platformexec.InstallCommand(platform, pkg)
		if err != nil {
			return nil, err
		}
		item := PlanItem{Package: pkg, Command: command}
		if b.Runner != nil {
			check, checkErr := platformexec.CheckCommand(platform, pkg)
			if checkErr != nil {
				return nil, checkErr
			}
			checkCtx, cancel := b.commandContext(ctx)
			result := b.Runner.Run(checkCtx, check)
			cancel()
			if result.Err == nil {
				item.Skipped = true
			} else if result.Output == "" {
				return nil, fmt.Errorf("检查 %s 安装状态失败: %w", pkg.Name, result.Err)
			}
		}
		items = append(items, item)
	}
	return items, nil
}

func (b Bootstrap) Install(ctx context.Context, items []PlanItem) error {
	report := b.InstallReport(ctx, items)
	if report.Failed > 0 {
		return report.Errors[0]
	}
	return nil
}

func (b Bootstrap) InstallReport(ctx context.Context, items []PlanItem) InstallReport {
	report := InstallReport{}
	for _, item := range items {
		if err := ctx.Err(); err != nil {
			report.Failed++
			report.Errors = append(report.Errors, fmt.Errorf("安装已取消或超时: %w", err))
			break
		}
		if item.Skipped {
			report.Skipped++
			continue
		}
		if b.Runner == nil {
			report.Failed++
			report.Errors = append(report.Errors, fmt.Errorf("未配置命令执行器"))
			continue
		}
		commandCtx, cancel := b.commandContext(ctx)
		result := b.Runner.Run(commandCtx, item.Command)
		cancel()
		if result.Err != nil {
			report.Failed++
			report.Errors = append(report.Errors, fmt.Errorf("安装 %s 失败: %w\n%s", item.Package.Name, result.Err, result.Output))
			continue
		}
		report.Succeeded++
	}
	return report
}
