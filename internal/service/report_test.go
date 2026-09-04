package service

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/gmh123521/java-dev-bootstrap/internal/detection"
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

type sequenceDetector struct {
	results []detection.Result
	index   int
}

func (d *sequenceDetector) Detect(_ context.Context, _ model.Package) detection.Result {
	if d.index >= len(d.results) {
		return detection.Result{Status: detection.StatusError, Err: errors.New("没有更多检测结果")}
	}
	result := d.results[d.index]
	d.index++
	return result
}

func TestInstallReportVerifiesSuccessfulInstall(t *testing.T) {
	items := []PlanItem{{Package: model.Package{ID: "gradle", Name: "Gradle"}, Command: ports.Command{Program: "gradle"}}}
	detector := &sequenceDetector{results: []detection.Result{{Status: detection.StatusInstalled, Version: "8.14"}}}
	report := (Bootstrap{Runner: reportRunner{}, Detector: detector}).InstallReport(context.Background(), items)
	if report.Succeeded != 1 || report.Verified != 1 || report.VerificationFailed != 0 {
		t.Fatalf("安装成功后应复查通过: %#v", report)
	}
}

func TestInstallReportReportsVerificationFailure(t *testing.T) {
	items := []PlanItem{{Package: model.Package{ID: "gradle", Name: "Gradle"}, Command: ports.Command{Program: "gradle"}}}
	detector := &sequenceDetector{results: []detection.Result{{Status: detection.StatusMissing, Detail: "仍未检测到"}}}
	report := (Bootstrap{Runner: reportRunner{}, Detector: detector}).InstallReport(context.Background(), items)
	if report.Succeeded != 1 || report.Verified != 0 || report.VerificationFailed != 1 || len(report.VerificationErrors) != 1 {
		t.Fatalf("复查失败应被单独记录: %#v", report)
	}
}

func TestInstallReturnsVerificationFailure(t *testing.T) {
	items := []PlanItem{{Package: model.Package{ID: "gradle", Name: "Gradle"}, Command: ports.Command{Program: "gradle"}}}
	detector := &sequenceDetector{results: []detection.Result{{Status: detection.StatusMissing, Detail: "仍未检测到"}}}
	if err := (Bootstrap{Runner: reportRunner{}, Detector: detector}).Install(context.Background(), items); err == nil || !strings.Contains(err.Error(), "复查 Gradle 失败") {
		t.Fatalf("Install 应返回复查错误: %v", err)
	}
}
