package detection

import (
	"context"
	"errors"
	"fmt"
	"os/exec"

	"github.com/gmh123521/java-dev-bootstrap/internal/model"
	"github.com/gmh123521/java-dev-bootstrap/internal/ports"
)

type ToolDetector struct {
	Runner   ports.Runner
	LookPath func(string) (string, error)
}

func (d ToolDetector) Detect(ctx context.Context, pkg model.Package) Result {
	if pkg.CheckProgram == "" {
		return Result{Status: StatusMissing, Detail: "未配置命令检测规则"}
	}
	if d.Runner == nil {
		return Result{Status: StatusError, Source: pkg.CheckProgram, Err: fmt.Errorf("未配置检测命令执行器")}
	}

	command := ports.Command{Program: pkg.CheckProgram, Args: pkg.CheckArgs}
	runResult := d.Runner.Run(ctx, command)
	if runResult.Err != nil {
		if errors.Is(runResult.Err, exec.ErrNotFound) {
			return Result{Status: StatusMissing, Source: pkg.CheckProgram, Detail: "命令不存在"}
		}
		return Result{Status: StatusError, Source: pkg.CheckProgram, Detail: runResult.Output, Err: runResult.Err}
	}

	version, err := VersionText(runResult.Output)
	if err != nil {
		return Result{Status: StatusError, Source: pkg.CheckProgram, Detail: runResult.Output, Err: err}
	}
	major, err := MajorVersion(version)
	if err != nil {
		return Result{Status: StatusError, Source: pkg.CheckProgram, Version: version, Err: err}
	}

	path := ""
	lookPath := d.LookPath
	if lookPath == nil {
		lookPath = exec.LookPath
	}
	if detectedPath, lookErr := lookPath(pkg.CheckProgram); lookErr == nil {
		path = detectedPath
	}

	status := StatusInstalled
	if !MeetsMinimum(major, pkg.MinVersion) {
		status = StatusOutdated
	}
	return Result{Status: status, Version: version, Source: pkg.CheckProgram, Path: path}
}
