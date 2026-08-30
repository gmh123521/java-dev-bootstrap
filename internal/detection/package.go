package detection

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/gmh123521/java-dev-bootstrap/internal/model"
	"github.com/gmh123521/java-dev-bootstrap/internal/platform"
	"github.com/gmh123521/java-dev-bootstrap/internal/ports"
)

type PackageDetector struct {
	Runner   ports.Runner
	Platform model.Platform
	LookPath func(string) (string, error)
}

func (d PackageDetector) Detect(ctx context.Context, pkg model.Package) Result {
	var commandError Result
	var commandMissing Result
	var commandOutdated Result
	if pkg.CheckProgram != "" {
		commandResult := (ToolDetector{Runner: d.Runner, LookPath: d.LookPath}).Detect(ctx, pkg)
		switch commandResult.Status {
		case StatusInstalled:
			return commandResult
		case StatusOutdated:
			commandOutdated = commandResult
		case StatusError:
			commandError = commandResult
		case StatusMissing:
			commandMissing = commandResult
		}
	}

	for _, configuredPattern := range pkg.CheckPaths {
		for _, pattern := range candidatePatterns(configuredPattern, d.Platform) {
			matches, err := filepath.Glob(expandEnvironment(pattern))
			if err != nil {
				return Result{Status: StatusError, Source: "path", Detail: pattern, Err: err}
			}
			if len(matches) > 0 {
				return Result{Status: StatusInstalled, Source: "path", Path: matches[0]}
			}
		}
	}

	if d.Runner == nil {
		if commandError.Status == StatusError {
			return commandError
		}
		return Result{Status: StatusError, Err: fmt.Errorf("未配置软件检测命令执行器")}
	}
	check, err := platform.CheckCommand(d.Platform, pkg)
	if err != nil {
		return Result{Status: StatusError, Err: err}
	}
	manager, err := platform.ManagerFor(d.Platform)
	if err != nil {
		return Result{Status: StatusError, Err: err}
	}
	managerResult := d.Runner.Run(ctx, check)
	if managerResult.Err == nil {
		return Result{Status: StatusInstalled, Source: manager, Detail: managerResult.Output}
	}
	if managerResult.Output != "" {
		if commandError.Status == StatusError {
			return commandError
		}
		if !isPackageMissingOutput(manager, managerResult.Output) {
			return Result{Status: StatusError, Source: manager, Detail: managerResult.Output, Err: managerResult.Err}
		}
		if commandOutdated.Status == StatusOutdated {
			return commandOutdated
		}
		return Result{Status: StatusMissing, Source: manager, Detail: managerResult.Output}
	}
	if commandOutdated.Status == StatusOutdated {
		return commandOutdated
	}
	if commandMissing.Status == StatusMissing {
		commandMissing.Detail = "命令不存在；包管理器状态查询不可用"
		return commandMissing
	}
	return Result{Status: StatusError, Source: manager, Err: managerResult.Err}
}

func candidatePatterns(pattern string, current model.Platform) []string {
	if current != model.PlatformWindows || !strings.Contains(pattern, "{drive}") {
		return []string{pattern}
	}
	patterns := make([]string, 0, 24)
	for drive := 'C'; drive <= 'Z'; drive++ {
		patterns = append(patterns, strings.ReplaceAll(pattern, "{drive}", string(drive)))
	}
	return patterns
}

func isPackageMissingOutput(manager, output string) bool {
	text := strings.ToLower(output)
	patterns := []string{
		"no installed package found",
		"no package found matching",
		"没有找到符合条件的已安装程序包",
		"没有找到与输入条件匹配的已安装程序包",
	}
	if manager == "brew" {
		patterns = append(patterns, "no such keg", "not installed")
	}
	for _, pattern := range patterns {
		if strings.Contains(text, pattern) {
			return true
		}
	}
	return false
}

var windowsEnvironmentPattern = regexp.MustCompile(`%([^%]+)%`)

func expandEnvironment(path string) string {
	expanded := os.ExpandEnv(path)
	return windowsEnvironmentPattern.ReplaceAllStringFunc(expanded, func(match string) string {
		parts := windowsEnvironmentPattern.FindStringSubmatch(match)
		return os.Getenv(parts[1])
	})
}
