package config

import (
	"os"
	"path/filepath"
	"testing"
)

func TestLoadManifestReadsYAML(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "manifest.yaml")
	content := "version: 1\npackages:\n  - id: jdk\n    name: JDK\n    description: Java\n    platforms: [windows]\n    manager: winget\n    manager_id: EclipseAdoptium.Temurin.21.JDK\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}
	manifest, err := LoadManifest(path)
	if err != nil {
		t.Fatalf("加载失败: %v", err)
	}
	if len(manifest.Packages) != 1 || manifest.Packages[0].ID != "jdk" {
		t.Fatalf("清单内容错误: %#v", manifest)
	}
}
