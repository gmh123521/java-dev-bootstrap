package cli

import (
	"context"
	"strings"
	"testing"
)

func TestRunHelpContainsCommands(t *testing.T) {
	var output strings.Builder
	if err := Run(context.Background(), nil, &output, &output); err != nil {
		t.Fatalf("帮助命令失败: %v", err)
	}
	for _, command := range []string{"list", "plan", "install", "doctor"} {
		if !strings.Contains(output.String(), command) {
			t.Fatalf("帮助缺少命令 %q: %s", command, output.String())
		}
	}
}
