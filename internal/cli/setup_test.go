package cli

import (
	"strings"
	"testing"

	"github.com/gmh123521/java-dev-bootstrap/internal/platform"
)

func TestFormatSetupGuideShowsNextCommand(t *testing.T) {
	actual := formatSetupGuide(platform.SetupResult{
		Ready:        true,
		Platform:     "windows（amd64）",
		Manager:      "winget",
		ManagerState: "v1.29.290",
		NextCommand:  "jdb doctor",
	})
	for _, expected := range []string{"准备检查：", "平台：windows（amd64）", "包管理器 winget：v1.29.290", "下一步：jdb doctor"} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("setup 正常输出缺少 %q: %s", expected, actual)
		}
	}
}

func TestFormatSetupGuideShowsSuggestion(t *testing.T) {
	actual := formatSetupGuide(platform.SetupResult{
		Platform:     "darwin（arm64）",
		Manager:      "brew",
		ManagerState: "不可用",
		Suggestion:   "请先安装 Homebrew",
		NextCommand:  "jdb prerequisites",
	})
	for _, expected := range []string{"准备检查：", "包管理器 brew：不可用", "建议：请先安装 Homebrew", "处理后执行：jdb prerequisites"} {
		if !strings.Contains(actual, expected) {
			t.Fatalf("setup 异常输出缺少 %q: %s", expected, actual)
		}
	}
}
