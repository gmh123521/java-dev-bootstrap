package config

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/gmh123521/java-dev-bootstrap/internal/model"
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

func TestLoadManifestPreservesCommaInsideQuotedListItem(t *testing.T) {
	dir := t.TempDir()
	path := filepath.Join(dir, "manifest.yaml")
	content := "version: 1\npackages:\n  - id: tool\n    name: Tool\n    description: Tool\n    platforms: [windows]\n    manager: winget\n    manager_id: Example.Tool\n    check_program: tool\n    check_args: [\"--format\", \"name,version\"]\n"
	if err := os.WriteFile(path, []byte(content), 0o600); err != nil {
		t.Fatal(err)
	}

	manifest, err := LoadManifest(path)
	if err != nil {
		t.Fatalf("加载带逗号参数失败: %v", err)
	}
	args := manifest.Packages[0].CheckArgs
	if len(args) != 2 || args[1] != "name,version" {
		t.Fatalf("引号内逗号不应拆分参数: %#v", args)
	}
}

func TestDefaultManifestUsesStrongDesktopAndJDKChecks(t *testing.T) {
	manifest, err := LoadDefaultManifest()
	if err != nil {
		t.Fatalf("加载默认清单失败: %v", err)
	}
	packages := map[string]model.Package{}
	for _, pkg := range manifest.Packages {
		packages[pkg.ID] = pkg
	}
	if packages["jdk"].CheckProgram != "javac" {
		t.Fatalf("JDK 必须通过 javac 检测: %#v", packages["jdk"])
	}
	if packages["docker"].CheckProgram != "" || len(packages["docker"].CheckPaths) == 0 {
		t.Fatalf("Docker Desktop 不应仅通过 docker CLI 检测: %#v", packages["docker"])
	}
	for _, path := range packages["intellij"].CheckPaths {
		if strings.HasPrefix(path, "%APPDATA%") {
			t.Fatalf("IntelliJ 不应通过可能残留的用户配置目录判断安装: %s", path)
		}
	}
}
