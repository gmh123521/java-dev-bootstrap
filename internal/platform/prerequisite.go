package platform

import (
	"context"
	"fmt"
	"strings"

	"github.com/gmh123521/java-dev-bootstrap/internal/model"
	"github.com/gmh123521/java-dev-bootstrap/internal/ports"
)

type PrerequisiteItem struct {
	Name       string
	Current    string
	Suggestion string
	OK         bool
}

func CheckPrerequisites(ctx context.Context, current model.Platform, arch string, runner ports.Runner) ([]PrerequisiteItem, error) {
	manager, err := ManagerFor(current)
	if err != nil {
		return nil, err
	}
	if strings.TrimSpace(arch) == "" {
		arch = "未知"
	}
	items := []PrerequisiteItem{{
		Name:    "操作系统",
		Current: fmt.Sprintf("%s（%s）", current, arch),
		OK:      true,
	}}
	managerItem := PrerequisiteItem{Name: "包管理器 " + manager}
	if runner == nil {
		managerItem.Current = "未配置检测执行器"
		managerItem.Suggestion = "请使用可执行命令环境重新运行"
		items = append(items, managerItem)
		return items, nil
	}
	check, err := ManagerCheckCommand(current)
	if err != nil {
		return nil, err
	}
	result := runner.Run(ctx, check)
	if result.Err == nil {
		managerItem.Current = result.Output
		managerItem.OK = true
	} else {
		managerItem.Current = strings.TrimSpace(result.Output)
		if managerItem.Current == "" {
			managerItem.Current = "不可用"
		}
		managerItem.Suggestion = prerequisiteSuggestion(current)
	}
	items = append(items, managerItem)
	return items, nil
}

func prerequisiteSuggestion(current model.Platform) string {
	if current == model.PlatformWindows {
		return "请安装或修复 Microsoft 应用安装程序（App Installer）中的 winget"
	}
	return "请先安装 Homebrew，并确认 brew 已加入 PATH"
}
