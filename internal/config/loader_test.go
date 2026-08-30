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

func TestLoadManifestReadsDetectionRules(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "manifest.yaml")
	content := "version: 1\npackages:\n  - id: jdk\n    name: JDK\n    description: Java\n    platforms: [windows]\n    manager: winget\n    manager_id: EclipseAdoptium.Temurin.21.JDK\n    check_program: java\n    check_args: [-version]\n    check_paths: [\"%APPDATA%\\\\JetBrains\\\\IdeaIC*\"]\n    min_version: 21\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	manifest, err := LoadManifest(path)
	if err != nil {
		t.Fatalf("加载检测规则失败: %v", err)
	}
	pkg := manifest.Packages[0]
	if pkg.CheckProgram != "java" || len(pkg.CheckArgs) != 1 || pkg.CheckArgs[0] != "-version" || pkg.MinVersion != 21 {
		t.Fatalf("检测规则内容错误: %#v", pkg)
	}
	if len(pkg.CheckPaths) != 1 || pkg.CheckPaths[0] != `%APPDATA%\JetBrains\IdeaIC*` {
		t.Fatalf("Windows 检测路径解析错误: %#v", pkg.CheckPaths)
	}
}
