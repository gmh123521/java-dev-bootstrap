package cli

import (
	"context"
	"strings"
	"testing"
)

func TestRunProfilesListsBuiltInProfiles(t *testing.T) {
	var output strings.Builder
	if err := Run(context.Background(), []string{"profiles"}, &output, &output); err != nil {
		t.Fatalf("profiles 命令失败: %v", err)
	}
	if !strings.Contains(output.String(), "java-basic") || !strings.Contains(output.String(), "spring-backend") {
		t.Fatalf("profile 列表不完整: %s", output.String())
	}
}

func TestRunVersionPrintsVersion(t *testing.T) {
	var output strings.Builder
	if err := Run(context.Background(), []string{"version"}, &output, &output); err != nil {
		t.Fatalf("version 命令失败: %v", err)
	}
	if !strings.Contains(output.String(), "版本") {
		t.Fatalf("版本输出不完整: %s", output.String())
	}
}
