package service

import (
	"testing"

	"github.com/gmh123521/java-dev-bootstrap/internal/model"
)

func TestProfileFiltersManifestPackages(t *testing.T) {
	manifest := model.Manifest{Version: 1, Packages: []model.Package{
		{ID: "jdk", Name: "JDK", Manager: "winget", ManagerID: "jdk"},
		{ID: "git", Name: "Git", Manager: "winget", ManagerID: "git"},
		{ID: "docker", Name: "Docker", Manager: "winget", ManagerID: "docker"},
	}}
	filtered, err := ApplyProfile(manifest, "java-basic")
	if err != nil {
		t.Fatalf("应用 profile 失败: %v", err)
	}
	if len(filtered.Packages) != 2 || filtered.Packages[0].ID != "jdk" || filtered.Packages[1].ID != "git" {
		t.Fatalf("profile 筛选错误: %#v", filtered.Packages)
	}
}

func TestProfileRejectsUnknownName(t *testing.T) {
	if _, err := ApplyProfile(model.Manifest{Version: 1}, "unknown"); err == nil {
		t.Fatal("未知 profile 应该失败")
	}
}
