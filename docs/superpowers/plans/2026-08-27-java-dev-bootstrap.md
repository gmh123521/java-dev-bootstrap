# Java Dev Bootstrap Implementation Plan

> **For agentic workers:** REQUIRED SUB-SKILL: Use superpowers:subagent-driven-development (recommended) or superpowers:executing-plans to implement this plan task-by-task. Steps use checkbox (`- [ ]`) syntax for tracking.

**Goal:** 创建一个可在 Windows/macOS 上运行的 Java 开发环境初始化 CLI 第一版。

**Architecture:** 使用 Go 构建单文件 CLI。命令层调用应用服务生成计划，应用服务通过清单和平台端口抽象环境探测与安装执行；Windows 使用 winget，macOS 使用 brew。

**Tech Stack:** Go 1.22+、Cobra、gopkg.in/yaml.v3、Go Test、GitHub Actions。

---

### Task 1: 建立 Go 模块和领域模型

**Files:**
- Create: `go.mod`
- Create: `internal/model/package.go`
- Create: `internal/model/manifest.go`
- Test: `internal/model/manifest_test.go`

- [ ] 写测试：验证清单能筛选当前平台、拒绝空包 ID。
- [ ] 运行 `go test ./internal/model`，确认因实现缺失失败。
- [ ] 实现包模型、平台枚举和清单校验。
- [ ] 再次运行测试并确认通过。

### Task 2: 实现 YAML 清单加载

**Files:**
- Create: `internal/config/loader.go`
- Test: `internal/config/loader_test.go`
- Create: `configs/default.yaml`

- [ ] 先写从 YAML 加载默认清单的失败测试。
- [ ] 实现 YAML 解析、字段校验和文件读取。
- [ ] 验证清单包含 JDK、Git、Maven、Gradle 和 IDE。

### Task 3: 实现平台执行器和应用服务

**Files:**
- Create: `internal/ports/runner.go`
- Create: `internal/platform/platform.go`
- Create: `internal/service/bootstrap.go`
- Test: `internal/service/bootstrap_test.go`

- [ ] 测试安装计划只包含当前平台包，并在探测到已安装时跳过。
- [ ] 实现命令构造器，使用参数数组而非 shell 拼接。
- [ ] 实现 dry-run、确认前计划和失败结果收集。

### Task 4: 实现 Cobra CLI

**Files:**
- Create: `cmd/jdb/main.go`
- Create: `internal/cli/cli.go`
- Test: `internal/cli/cli_test.go`

- [ ] 测试 `list` 和 `plan` 输出包含中文名称与安装命令。
- [ ] 实现 `list`、`plan`、`install`、`doctor`。
- [ ] 为 `install` 增加 `--yes` 与 `--manifest` 参数。

### Task 5: 文档、CI 和发布基础

**Files:**
- Create: `README.md`
- Create: `.gitignore`
- Create: `.github/workflows/ci.yml`
- Create: `.goreleaser.yaml`

- [ ] 编写中文使用说明、安全边界和 Windows/macOS 前置条件。
- [ ] 配置 Windows/macOS 的 Go 测试矩阵。
- [ ] 本地运行格式化、测试和构建。
- [ ] 提交中文 Git 提交信息。
