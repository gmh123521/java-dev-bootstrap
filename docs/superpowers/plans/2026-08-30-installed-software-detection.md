# 已安装软件检测与环境诊断实施计划

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 让 `plan` 和 `install` 准确识别已有开发工具、跳过满足要求的软件，并通过 `doctor` 报告环境变量问题。

**Architecture:** 新建独立的 `internal/detection` 包，将命令检测、版本解析和检测结果从 CLI 与安装服务中分离。`Bootstrap.Plan` 通过检测器生成统一计划，CLI 只负责展示；环境诊断保持只读，不直接修改用户的系统环境变量。

**Tech Stack:** Go 1.23、标准库 `os/exec`/`regexp`/`runtime`、现有 Runner 端口、Go 单元测试、winget、Homebrew。

---

## 文件结构

- 新建 `internal/detection/detection.go`：状态、结果、检测器接口和组合检测器。
- 新建 `internal/detection/version.go`：版本提取和最低主版本比较。
- 新建 `internal/detection/tool.go`：内置软件命令规则。
- 新建 `internal/detection/environment.go`：Java 环境变量诊断。
- 修改 `internal/service/bootstrap.go`：计划使用检测器，不再直接依赖包管理器检查。
- 修改 `internal/cli/cli.go`：`plan` 启用只读检测并展示详细结果，`doctor` 输出环境诊断。
- 修改 `README.md`：说明检测规则、安装位置和环境变量边界。

### Task 1: 通用版本解析与检测结果

**Files:**
- Create: `internal/detection/detection.go`
- Create: `internal/detection/version.go`
- Create: `internal/detection/version_test.go`

- [ ] **Step 1: 写版本解析失败测试**

```go
func TestMajorVersionParsesJavaAndToolOutputs(t *testing.T) {
    cases := map[string]int{
        `java version "23.0.2"`: 23,
        `openjdk version "21.0.5"`: 21,
        `git version 2.51.1.windows.1`: 2,
        `Apache Maven 3.9.9`: 3,
    }
    for output, expected := range cases {
        actual, err := MajorVersion(output)
        if err != nil || actual != expected {
            t.Fatalf("解析 %q：得到 %d, %v", output, actual, err)
        }
    }
}
```

- [ ] **Step 2: 运行测试并确认因函数不存在而失败**

Run: `go test ./internal/detection -run TestMajorVersionParsesJavaAndToolOutputs -v`

Expected: FAIL，提示 `undefined: MajorVersion`。

- [ ] **Step 3: 实现结果类型和最小版本解析**

```go
type Status string

const (
    StatusInstalled Status = "installed"
    StatusOutdated  Status = "outdated"
    StatusMissing   Status = "missing"
    StatusError     Status = "error"
)

type Result struct {
    Status  Status
    Version string
    Source  string
    Path    string
    Detail  string
}
```

`MajorVersion` 使用正则寻找第一个数字段并返回整数；空输出或无版本号返回中文错误。

- [ ] **Step 4: 补充版本下限测试并实现 `MeetsMinimum`**

覆盖 Java 23 满足最低 21、Java 17 不满足最低 21、未设置最低版本始终满足。

- [ ] **Step 5: 运行检测包测试**

Run: `go test ./internal/detection -v`

Expected: PASS。

- [ ] **Step 6: 提交第二批**

```bash
git add internal/detection
git commit -m "增加通用软件和版本检测器"
```

### Task 2: 内置软件命令检测

**Files:**
- Create: `internal/detection/tool.go`
- Create: `internal/detection/tool_test.go`
- Modify: `internal/model/package.go`
- Modify: `configs/default.yaml`
- Modify: `internal/config/default.yaml`

- [ ] **Step 1: 写 Java 23 满足 JDK 21 的失败测试**

测试使用假 Runner 返回 `java version "23.0.2"`，期望状态为 `StatusInstalled`、版本为 `23.0.2`、来源为 `java`。

- [ ] **Step 2: 写命令不存在时的失败测试**

假 Runner 返回 `exec.ErrNotFound`，期望 `StatusMissing`，而不是 `StatusError`。

- [ ] **Step 3: 运行测试确认失败**

Run: `go test ./internal/detection -run 'TestToolDetector' -v`

Expected: FAIL，提示 `ToolDetector` 或规则尚未定义。

- [ ] **Step 4: 扩展清单模型**

给 `model.Package` 增加：

```go
CheckProgram string   `yaml:"check_program"`
CheckArgs    []string `yaml:"check_args"`
MinVersion   int      `yaml:"min_version"`
```

内置清单配置：JDK 使用 `java -version`、最低 21；Git 使用 `git --version`；Maven 使用 `mvn -version`；Gradle 使用 `gradle --version`；VS Code 使用 `code --version`；Docker 使用 `docker --version`。

- [ ] **Step 5: 实现 `ToolDetector`**

检测器执行配置的命令，解析版本；命令不存在返回 Missing，版本过低返回 Outdated，成功返回 Installed。未配置命令时返回“不适用”，交给后续检测器。

- [ ] **Step 6: 运行配置和检测测试**

Run: `go test ./internal/detection ./internal/config ./internal/model -v`

Expected: PASS。

- [ ] **Step 7: 提交配置扩展**

```bash
git add internal/detection internal/model/package.go configs/default.yaml internal/config/default.yaml
git commit -m "扩展开发工具命令检测规则"
```

### Task 3: 安装计划识别并跳过已有软件

**Files:**
- Modify: `internal/service/bootstrap.go`
- Modify: `internal/service/bootstrap_test.go`
- Modify: `internal/cli/cli.go`
- Modify: `internal/cli/cli_test.go`

- [ ] **Step 1: 写计划使用检测器的失败测试**

构造一个返回 Installed 的假检测器，调用 `Bootstrap.Plan` 后断言 `Skipped == true` 且计划项保留检测结果。

- [ ] **Step 2: 写缺失软件仍待安装的失败测试**

假检测器返回 Missing，断言 `Skipped == false` 且安装命令仍正确生成。

- [ ] **Step 3: 运行服务测试确认失败**

Run: `go test ./internal/service -run TestPlan -v`

Expected: FAIL，提示 `Detector` 字段或检测结果不存在。

- [ ] **Step 4: 修改计划模型**

`PlanItem` 增加 `Detection detection.Result`，`Bootstrap` 增加 `Detector detection.Detector`。有检测器时统一调用检测器；Installed 跳过，Missing/Outdated 保留待安装，Error 返回错误阻止安装。

- [ ] **Step 5: 让 `plan` 和 `install` 都创建真实 Runner 与组合检测器**

`plan` 只运行版本/路径/包管理器查询，不创建日志、不执行安装。`install` 使用同一检测器并继续保留安装日志。

- [ ] **Step 6: 更新 CLI 输出测试**

断言输出包含 `已安装`、版本、来源和 `[跳过]`；缺失项显示 `[待安装]`。

- [ ] **Step 7: 运行服务和 CLI 测试**

Run: `go test ./internal/service ./internal/cli -v`

Expected: PASS。

- [ ] **Step 8: 提交第三批**

```bash
git add internal/service internal/cli
git commit -m "让安装计划识别并跳过已有软件"
```

### Task 4: 环境变量诊断

**Files:**
- Create: `internal/detection/environment.go`
- Create: `internal/detection/environment_test.go`
- Modify: `internal/cli/cli.go`
- Modify: `internal/cli/cli_test.go`

- [ ] **Step 1: 写 JAVA_HOME 缺失和不一致测试**

通过注入读取环境变量与查找命令路径的函数，分别断言缺失时给出警告、路径一致时通过、路径不一致时给出两条实际路径。

- [ ] **Step 2: 运行测试确认失败**

Run: `go test ./internal/detection -run TestEnvironment -v`

Expected: FAIL，提示环境诊断 API 不存在。

- [ ] **Step 3: 实现只读环境诊断**

结果包含级别、项目、当前值和建议；检查 `JAVA_HOME` 目录、`JAVA_HOME/bin/java`、PATH 中的 java，以及 Git/Maven/Gradle/Docker 命令可用性。不调用 `setx`，不修改注册表。

- [ ] **Step 4: 接入 `doctor` 并测试中文输出**

包管理器检查通过后输出环境诊断列表；警告不导致 `doctor` 失败，真正的检测执行错误才失败。

- [ ] **Step 5: 运行检测和 CLI 测试**

Run: `go test ./internal/detection ./internal/cli -v`

Expected: PASS。

- [ ] **Step 6: 提交第四批**

```bash
git add internal/detection/environment.go internal/detection/environment_test.go internal/cli
git commit -m "增强开发环境变量诊断"
```

### Task 5: 文档、忽略项和全量验证

**Files:**
- Modify: `README.md`
- Modify: `.gitignore`

- [ ] **Step 1: 更新使用说明**

说明 `plan` 会真实检测但不修改系统；JDK 21 及以上满足要求；安装位置由官方安装器决定；程序不自动覆盖环境变量；列出 Windows/macOS 常见配置目录。

- [ ] **Step 2: 忽略本地测试缓存和发布二进制**

加入 `.go-cache/`、`jdb-windows-*.exe`、`jdb-darwin-*`，防止本地验证产物误提交。

- [ ] **Step 3: 运行格式化和全量测试**

Run: `gofmt -w internal/detection internal/service internal/cli internal/model`

Run: `go test ./...`

Run: `go vet ./...`

Expected: 全部退出码为 0。

- [ ] **Step 4: 验证四个平台构建**

分别设置 `GOOS/GOARCH` 构建 `windows-amd64`、`windows-arm64`、`darwin-amd64`、`darwin-arm64`，使用 `-buildvcs=false`，Expected: 四次构建成功。

- [ ] **Step 5: 在当前 Windows 环境做只读验收**

Run: `jdb.exe plan --profile java-basic`

Expected: Java 23、Git、Maven、IntelliJ 被识别并跳过；Gradle 根据实际环境显示待安装；不弹出安装确认。

- [ ] **Step 6: 提交第五批**

```bash
git add README.md .gitignore
git commit -m "补充安装位置与环境配置说明"
```

- [ ] **Step 7: 检查提交边界**

Run: `git status --short && git log --oneline -8`

Expected: 工作区干净，每批提交内容单一且提交信息为中文。
