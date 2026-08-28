package model

import "fmt"

type Manifest struct {
	Version  int       `yaml:"version"`
	Packages []Package `yaml:"packages"`
}

func (m Manifest) Validate() error {
	if m.Version < 1 {
		return fmt.Errorf("清单版本必须大于等于 1")
	}
	seen := map[string]bool{}
	for _, pkg := range m.Packages {
		if pkg.ID == "" {
			return fmt.Errorf("软件包 ID 不能为空")
		}
		if pkg.Name == "" || pkg.Manager == "" || pkg.ManagerID == "" {
			return fmt.Errorf("软件包 %q 的名称、安装器和安装器 ID 不能为空", pkg.ID)
		}
		if seen[pkg.ID] {
			return fmt.Errorf("软件包 ID 重复: %s", pkg.ID)
		}
		seen[pkg.ID] = true
	}
	return nil
}

func (m Manifest) PackagesForPlatform(platform Platform) ([]Package, error) {
	if err := m.Validate(); err != nil {
		return nil, err
	}
	result := make([]Package, 0, len(m.Packages))
	for _, pkg := range m.Packages {
		if pkg.Supports(platform) {
			result = append(result, pkg)
		}
	}
	return result, nil
}
