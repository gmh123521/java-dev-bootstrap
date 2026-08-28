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

func TestRunHelpContainsDryRunOption(t *testing.T) {
	var output strings.Builder
	if err := Run(context.Background(), nil, &output, &output); err != nil {
		t.Fatalf("帮助命令失败: %v", err)
	}
	if !strings.Contains(output.String(), "--dry-run") {
		t.Fatalf("帮助缺少 --dry-run: %s", output.String())
	}
}

func TestRunDryRunDoesNotAskForConfirmation(t *testing.T) {
	var output strings.Builder
	if err := Run(context.Background(), []string{"install", "--dry-run"}, &output, &output); err != nil {
		t.Fatalf("dry-run 不应失败: %v", err)
	}
	if strings.Contains(output.String(), "继续安装") {
		t.Fatalf("dry-run 不应要求确认: %s", output.String())
	}
}
