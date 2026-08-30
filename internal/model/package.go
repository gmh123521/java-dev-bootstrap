package model

import (
	"fmt"
	"runtime"
)

type Platform string

const (
	PlatformWindows Platform = "windows"
	PlatformDarwin  Platform = "darwin"
)

func CurrentPlatform() (Platform, error) {
	switch runtime.GOOS {
	case "windows":
		return PlatformWindows, nil
	case "darwin":
		return PlatformDarwin, nil
	default:
		return "", fmt.Errorf("暂不支持操作系统: %s", runtime.GOOS)
	}
}

type Package struct {
	ID           string     `yaml:"id"`
	Name         string     `yaml:"name"`
	Description  string     `yaml:"description"`
	Platforms    []Platform `yaml:"platforms"`
	Manager      string     `yaml:"manager"`
	ManagerID    string     `yaml:"manager_id"`
	Kind         string     `yaml:"kind"`
	DarwinID     string     `yaml:"darwin_id"`
	CheckProgram string     `yaml:"check_program"`
	CheckArgs    []string   `yaml:"check_args"`
	CheckPaths   []string   `yaml:"check_paths"`
	MinVersion   int        `yaml:"min_version"`
}

func (p Package) ManagerIDFor(platform Platform) string {
	if platform == PlatformDarwin && p.DarwinID != "" {
		return p.DarwinID
	}
	return p.ManagerID
}

func (p Package) Supports(platform Platform) bool {
	for _, candidate := range p.Platforms {
		if candidate == platform {
			return true
		}
	}
	return false
}
