package service

import (
	"context"
	"errors"
	"testing"

	"github.com/gmh123521/java-dev-bootstrap/internal/model"
	"github.com/gmh123521/java-dev-bootstrap/internal/ports"
)

type reportRunner struct{ failures map[string]error }

func (r reportRunner) Run(_ context.Context, command ports.Command) ports.Result {
	if err := r.failures[command.Program]; err != nil {
		return ports.Result{Command: command, Err: err}
	}
	return ports.Result{Command: command}
}

func TestInstallReportContinuesAfterFailure(t *testing.T) {
	items := []PlanItem{
		{Package: model.Package{ID: "one", Name: "第一个"}, Command: ports.Command{Program: "one"}},
		{Package: model.Package{ID: "two", Name: "第二个"}, Command: ports.Command{Program: "two"}},
	}
	report := (Bootstrap{Runner: reportRunner{failures: map[string]error{"one": errors.New("失败")}}}).InstallReport(context.Background(), items)
	if report.Succeeded != 1 || report.Failed != 1 || len(report.Errors) != 1 {
		t.Fatalf("安装汇总错误: %#v", report)
	}
}
