package detection

import (
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	"github.com/gmh123521/java-dev-bootstrap/internal/model"
	"github.com/gmh123521/java-dev-bootstrap/internal/ports"
)

type Level string

const (
	LevelOK      Level = "ok"
	LevelWarning Level = "warning"
)

type Diagnostic struct {
	Level      Level
	Name       string
	Current    string
	Suggestion string
}

type EnvironmentDetector struct {
	Runner   ports.Runner
	Getenv   func(string) string
	LookPath func(string) (string, error)
	Stat     func(string) (os.FileInfo, error)
}

func (d EnvironmentDetector) Diagnose(ctx context.Context, current model.Platform) []Diagnostic {
	getenv := d.Getenv
	if getenv == nil {
		getenv = os.Getenv
	}
	lookPath := d.LookPath
	if lookPath == nil {
		lookPath = exec.LookPath
	}
	stat := d.Stat
	if stat == nil {
		stat = os.Stat
	}

	javaName := "java"
	javacName := "javac"
	if current == model.PlatformWindows {
		javaName = "java.exe"
		javacName = "javac.exe"
	}
	items := make([]Diagnostic, 0, 5)
	home := getenv("JAVA_HOME")
	homeJava := ""
	homeVersion := ""
	if home == "" {
		items = append(items, Diagnostic{
			Level:      LevelWarning,
			Name:       "JAVA_HOME",
			Current:    "未设置",
			Suggestion: "将 JAVA_HOME 指向当前使用的 JDK 根目录",
		})
	} else if info, err := stat(home); err != nil || !info.IsDir() {
		items = append(items, Diagnostic{
			Level:      LevelWarning,
			Name:       "JAVA_HOME",
			Current:    home,
			Suggestion: "JAVA_HOME 指向的目录不存在，请改为有效 JDK 根目录",
		})
	} else {
		homeJava = filepath.Join(home, "bin", javaName)
		homeJavac := filepath.Join(home, "bin", javacName)
		if _, err := stat(homeJava); err != nil {
			items = append(items, Diagnostic{
				Level:      LevelWarning,
				Name:       "JAVA_HOME",
				Current:    home,
				Suggestion: fmt.Sprintf("没有找到 %s，请确认 JAVA_HOME 指向有效 Java 安装", homeJava),
			})
			homeJava = ""
		} else {
			items = append(items, Diagnostic{Level: LevelOK, Name: "JAVA_HOME", Current: homeJava})
			if _, err := stat(homeJavac); err != nil {
				items = append(items, Diagnostic{
					Level:      LevelWarning,
					Name:       "JDK 完整性",
					Current:    "缺少 " + homeJavac,
					Suggestion: "JAVA_HOME 可能指向 JRE，请改为包含 javac 的完整 JDK",
				})
			} else {
				items = append(items, Diagnostic{Level: LevelOK, Name: "JDK 完整性", Current: homeJavac})
			}
			if d.Runner != nil {
				homeVersion, _ = d.javaVersion(ctx, homeJava)
			}
		}
	}

	pathJava, err := lookPath("java")
	if err != nil {
		items = append(items, Diagnostic{
			Level:      LevelWarning,
			Name:       "Java PATH",
			Current:    "未找到 java 命令",
			Suggestion: "将 JAVA_HOME 的 bin 目录加入 PATH",
		})
		return items
	}
	pathVersion := ""
	if d.Runner == nil {
		items = append(items, Diagnostic{Level: LevelWarning, Name: "Java PATH", Current: pathJava, Suggestion: "未配置诊断命令执行器"})
		return items
	}
	pathVersion, pathErr := d.javaVersion(ctx, pathJava)
	if pathErr != nil {
		items = append(items, Diagnostic{Level: LevelWarning, Name: "Java PATH", Current: pathJava, Suggestion: "java 命令无法正常执行"})
		return items
	}
	items = append(items, Diagnostic{Level: LevelOK, Name: "Java PATH", Current: fmt.Sprintf("%s（%s）", pathVersion, pathJava)})
	if homeVersion == "" {
		return items
	}
	if homeVersion != pathVersion {
		items = append(items, Diagnostic{
			Level:      LevelWarning,
			Name:       "Java 版本",
			Current:    fmt.Sprintf("JAVA_HOME=%s（%s），PATH=%s（%s）", homeVersion, homeJava, pathVersion, pathJava),
			Suggestion: "调整 PATH 顺序，使 java 与 JAVA_HOME 使用同一 JDK",
		})
		return items
	}
	items = append(items, Diagnostic{Level: LevelOK, Name: "Java 版本", Current: fmt.Sprintf("%s（JAVA_HOME 与 PATH 一致）", homeVersion)})
	return items
}

func (d EnvironmentDetector) DiagnoseTools(ctx context.Context, packages []model.Package) []Diagnostic {
	items := make([]Diagnostic, 0, len(packages))
	toolDetector := ToolDetector{Runner: d.Runner, LookPath: d.LookPath}
	for _, pkg := range packages {
		if pkg.ID == "jdk" || pkg.CheckProgram == "" {
			continue
		}
		result := toolDetector.Detect(ctx, pkg)
		switch result.Status {
		case StatusInstalled:
			current := result.Version
			if result.Path != "" {
				current += "（" + result.Path + "）"
			}
			items = append(items, Diagnostic{Level: LevelOK, Name: pkg.Name, Current: current})
		case StatusOutdated:
			items = append(items, Diagnostic{
				Level:      LevelWarning,
				Name:       pkg.Name,
				Current:    "版本 " + result.Version + " 低于要求",
				Suggestion: "升级到清单要求的最低版本",
			})
		case StatusMissing:
			items = append(items, Diagnostic{
				Level:      LevelWarning,
				Name:       pkg.Name,
				Current:    "未找到 " + pkg.CheckProgram + " 命令",
				Suggestion: "确认软件已安装并将命令目录加入 PATH",
			})
		case StatusError:
			items = append(items, Diagnostic{
				Level:      LevelWarning,
				Name:       pkg.Name,
				Current:    "检测失败",
				Suggestion: result.Detail,
			})
		}
	}
	return items
}

func (d EnvironmentDetector) DiagnoseApplications(ctx context.Context, packages []model.Package, detector Detector) []Diagnostic {
	items := make([]Diagnostic, 0, len(packages))
	for _, pkg := range packages {
		if pkg.CheckProgram != "" || detector == nil {
			continue
		}
		result := detector.Detect(ctx, pkg)
		switch result.Status {
		case StatusInstalled:
			current := "已安装"
			if result.Path != "" {
				current += "（" + result.Path + "）"
			} else if result.Source != "" {
				current += "，来源 " + result.Source
			}
			items = append(items, Diagnostic{Level: LevelOK, Name: pkg.Name, Current: current})
		case StatusMissing:
			items = append(items, Diagnostic{Level: LevelWarning, Name: pkg.Name, Current: "未检测到应用", Suggestion: "确认应用已安装，或通过安装计划补齐"})
		case StatusOutdated:
			items = append(items, Diagnostic{Level: LevelWarning, Name: pkg.Name, Current: "应用版本低于要求", Suggestion: "升级应用"})
		case StatusError:
			items = append(items, Diagnostic{Level: LevelWarning, Name: pkg.Name, Current: "检测失败", Suggestion: result.Detail})
		}
	}
	return items
}

func (d EnvironmentDetector) javaVersion(ctx context.Context, program string) (string, error) {
	result := d.Runner.Run(ctx, ports.Command{Program: program, Args: []string{"-version"}})
	if result.Err != nil {
		return "", result.Err
	}
	return VersionText(result.Output)
}
