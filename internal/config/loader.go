package config

import (
	"bufio"
	_ "embed"
	"fmt"
	"os"
	"strings"

	"github.com/gmh123521/java-dev-bootstrap/internal/model"
)

//go:embed default.yaml
var defaultManifest []byte

func LoadDefaultManifest() (model.Manifest, error) {
	return ParseManifest(defaultManifest)
}

func LoadManifest(path string) (model.Manifest, error) {
	data, err := os.ReadFile(path)
	if err != nil {
		return model.Manifest{}, fmt.Errorf("读取清单失败: %w", err)
	}
	manifest, err := ParseManifest(data)
	if err != nil {
		return model.Manifest{}, fmt.Errorf("解析清单失败: %w", err)
	}
	if err := manifest.Validate(); err != nil {
		return model.Manifest{}, fmt.Errorf("清单校验失败: %w", err)
	}
	return manifest, nil
}

func ParseManifest(data []byte) (model.Manifest, error) {
	manifest, err := parseManifest(string(data))
	if err != nil {
		return model.Manifest{}, err
	}
	if err := manifest.Validate(); err != nil {
		return model.Manifest{}, err
	}
	return manifest, nil
}

// parseManifest 只解析本项目公开清单所需的 YAML 子集，避免启动器依赖外部运行环境。
func parseManifest(content string) (model.Manifest, error) {
	var manifest model.Manifest
	var current *model.Package
	scanner := bufio.NewScanner(strings.NewReader(content))
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" || strings.HasPrefix(line, "#") || line == "packages:" {
			continue
		}
		if strings.HasPrefix(line, "version:") {
			if strings.TrimSpace(strings.TrimPrefix(line, "version:")) != "1" {
				return manifest, fmt.Errorf("目前仅支持清单版本 1")
			}
			manifest.Version = 1
			continue
		}
		if strings.HasPrefix(line, "- id:") {
			manifest.Packages = append(manifest.Packages, model.Package{ID: valueAfterColon(line)})
			current = &manifest.Packages[len(manifest.Packages)-1]
			continue
		}
		if current == nil {
			return manifest, fmt.Errorf("软件包字段必须位于 id 之后: %s", line)
		}
		switch {
		case strings.HasPrefix(line, "name:"):
			current.Name = valueAfterColon(line)
		case strings.HasPrefix(line, "description:"):
			current.Description = valueAfterColon(line)
		case strings.HasPrefix(line, "platforms:"):
			value := strings.Trim(valueAfterColon(line), "[]")
			for _, item := range strings.Split(value, ",") {
				current.Platforms = append(current.Platforms, model.Platform(strings.TrimSpace(item)))
			}
		case strings.HasPrefix(line, "manager:"):
			current.Manager = valueAfterColon(line)
		case strings.HasPrefix(line, "manager_id:"):
			current.ManagerID = valueAfterColon(line)
		case strings.HasPrefix(line, "kind:"):
			current.Kind = valueAfterColon(line)
		case strings.HasPrefix(line, "darwin_id:"):
			current.DarwinID = valueAfterColon(line)
		default:
			return manifest, fmt.Errorf("无法识别的字段: %s", line)
		}
	}
	if err := scanner.Err(); err != nil {
		return manifest, err
	}
	return manifest, nil
}

func valueAfterColon(line string) string {
	parts := strings.SplitN(line, ":", 2)
	if len(parts) != 2 {
		return ""
	}
	return strings.Trim(strings.TrimSpace(parts[1]), "\"")
}
