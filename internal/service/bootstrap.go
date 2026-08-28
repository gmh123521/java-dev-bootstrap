package service

import (
	"context"
	"fmt"

	"github.com/gmh123521/java-dev-bootstrap/internal/model"
	platformexec "github.com/gmh123521/java-dev-bootstrap/internal/platform"
	"github.com/gmh123521/java-dev-bootstrap/internal/ports"
)

type PlanItem struct {
	Package model.Package
	Command ports.Command
	Skipped bool
}

type Bootstrap struct {
	Runner ports.Runner
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
			result := b.Runner.Run(ctx, check)
			item.Skipped = result.Err == nil
		}
		items = append(items, item)
	}
	return items, nil
}

func (b Bootstrap) Install(ctx context.Context, items []PlanItem) error {
	for _, item := range items {
		if item.Skipped {
			continue
		}
		if b.Runner == nil {
			return fmt.Errorf("未配置命令执行器")
		}
		result := b.Runner.Run(ctx, item.Command)
		if result.Err != nil {
			return fmt.Errorf("安装 %s 失败: %w\n%s", item.Package.Name, result.Err, result.Output)
		}
	}
	return nil
}
