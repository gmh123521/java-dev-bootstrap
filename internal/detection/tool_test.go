package detection

import (
	"context"
	"errors"
	"os/exec"
	"testing"

	"github.com/gmh123521/java-dev-bootstrap/internal/model"
	"github.com/gmh123521/java-dev-bootstrap/internal/ports"
)

type toolRunner struct {
	result ports.Result
}

func (r toolRunner) Run(_ context.Context, command ports.Command) ports.Result {
	result := r.result
	result.Command = command
	return result
}

func TestToolDetectorAcceptsJava23ForMinimum21(t *testing.T) {
	pkg := model.Package{ID: "jdk", CheckProgram: "java", CheckArgs: []string{"-version"}, MinVersion: 21}
	detector := ToolDetector{Runner: toolRunner{result: ports.Result{Output: `java version "23.0.2"`}}}

	result := detector.Detect(context.Background(), pkg)

	if result.Status != StatusInstalled || result.Version != "23.0.2" || result.Source != "java" {
		t.Fatalf("Java 23 应满足最低 Java 21: %#v", result)
	}
}

func TestToolDetectorMarksOlderVersionAsOutdated(t *testing.T) {
	pkg := model.Package{ID: "jdk", CheckProgram: "java", CheckArgs: []string{"-version"}, MinVersion: 21}
	detector := ToolDetector{Runner: toolRunner{result: ports.Result{Output: `openjdk version "17.0.12"`}}}

	result := detector.Detect(context.Background(), pkg)

	if result.Status != StatusOutdated || result.Version != "17.0.12" {
		t.Fatalf("Java 17 应标记为版本过低: %#v", result)
	}
}

func TestToolDetectorTreatsMissingCommandAsMissing(t *testing.T) {
	pkg := model.Package{ID: "gradle", CheckProgram: "gradle", CheckArgs: []string{"--version"}}
	detector := ToolDetector{Runner: toolRunner{result: ports.Result{Err: errors.Join(exec.ErrNotFound, errors.New("找不到命令"))}}}

	result := detector.Detect(context.Background(), pkg)

	if result.Status != StatusMissing {
		t.Fatalf("命令不存在应标记为未安装: %#v", result)
	}
}

func TestToolDetectorTreatsMacJavaShimWithoutRuntimeAsMissing(t *testing.T) {
	pkg := model.Package{ID: "jdk", CheckProgram: "javac", CheckArgs: []string{"-version"}, MinVersion: 21}
	detector := ToolDetector{Runner: toolRunner{result: ports.Result{
		Output: "The operation couldn't be completed. Unable to locate a Java Runtime.",
		Err:    errors.New("exit status 1"),
	}}}

	result := detector.Detect(context.Background(), pkg)

	if result.Status != StatusMissing {
		t.Fatalf("macOS Java shim 未找到运行时时应标记 JDK 缺失: %#v", result)
	}
}
