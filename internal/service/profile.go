package service

import (
	"fmt"

	"github.com/gmh123521/java-dev-bootstrap/internal/model"
)

var profiles = map[string][]string{
	"java-basic":     {"jdk", "git", "maven", "gradle", "intellij"},
	"spring-backend": {"jdk", "git", "maven", "gradle", "intellij", "docker"},
	"java-fullstack": {"jdk", "git", "maven", "gradle", "intellij", "docker", "vscode"},
}

func ProfileNames() []string {
	return []string{"java-basic", "spring-backend", "java-fullstack"}
}

func ApplyProfile(manifest model.Manifest, name string) (model.Manifest, error) {
	ids, ok := profiles[name]
	if !ok {
		return model.Manifest{}, fmt.Errorf("未知 profile: %s", name)
	}
	allowed := make(map[string]bool, len(ids))
	for _, id := range ids {
		allowed[id] = true
	}
	filtered := manifest
	filtered.Packages = make([]model.Package, 0, len(ids))
	for _, pkg := range manifest.Packages {
		if allowed[pkg.ID] {
			filtered.Packages = append(filtered.Packages, pkg)
		}
	}
	return filtered, nil
}
