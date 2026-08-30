package detection

import (
	"context"
	"fmt"
	"os"
	"path/filepath"
	"regexp"

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
	if pkg.CheckProgram != "" {
		commandResult := (ToolDetector{Runner: d.Runner, LookPath: d.LookPath}).Detect(ctx, pkg)
		switch commandResult.Status {
		case StatusInstalled, StatusOutdated:
			return commandResult
		case StatusError:
			commandError = commandResult
		}
	}

	for _, pattern := range pkg.CheckPaths {
		matches, err := filepath.Glob(expandEnvironment(pattern))
		if err != nil {
			return Result{Status: StatusError, Source: "path", Detail: pattern, Err: err}
		}
		if len(matches) > 0 {
			return Result{Status: StatusInstalled, Source: "path", Path: matches[0]}
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
		return Result{Status: StatusMissing, Source: manager, Detail: managerResult.Output}
	}
	return Result{Status: StatusError, Source: manager, Err: managerResult.Err}
}

var windowsEnvironmentPattern = regexp.MustCompile(`%([^%]+)%`)

func expandEnvironment(path string) string {
	expanded := os.ExpandEnv(path)
	return windowsEnvironmentPattern.ReplaceAllStringFunc(expanded, func(match string) string {
		parts := windowsEnvironmentPattern.FindStringSubmatch(match)
		return os.Getenv(parts[1])
	})
}
