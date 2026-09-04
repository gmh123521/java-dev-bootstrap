package platform

import (
	"context"

	"github.com/gmh123521/java-dev-bootstrap/internal/model"
	"github.com/gmh123521/java-dev-bootstrap/internal/ports"
)

// SetupResult 描述 setup 命令对当前机器的准备检查结果。
type SetupResult struct {
	Ready        bool
	Platform     string
	Architecture string
	Manager      string
	ManagerState string
	Suggestion   string
	NextCommand  string
}

// SetupGuide 检查安装所需前置条件，并生成不修改系统的处理建议。
func SetupGuide(ctx context.Context, current model.Platform, arch string, runner ports.Runner) (SetupResult, error) {
	items, err := CheckPrerequisites(ctx, current, arch, runner)
	if err != nil {
		return SetupResult{}, err
	}
	manager, err := ManagerFor(current)
	if err != nil {
		return SetupResult{}, err
	}
	result := SetupResult{
		Platform:     items[0].Current,
		Architecture: arch,
		Manager:      manager,
		ManagerState: items[1].Current,
		NextCommand:  "jdb prerequisites",
	}
	result.Ready = items[1].OK
	if result.Ready {
		result.NextCommand = "jdb doctor"
	} else {
		result.Suggestion = prerequisiteSuggestion(current)
	}
	return result, nil
}
