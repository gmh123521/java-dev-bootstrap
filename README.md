# Java Dev Bootstrap

Java Dev Bootstrap 是一个面向 Windows 和 macOS 的 Java 开发环境初始化 CLI。它可以检查当前环境、预览安装计划，并在用户确认后调用系统包管理器安装常用工具。

## 功能

- Eclipse Temurin JDK 21
- Git、Maven、Gradle
- IntelliJ IDEA Community、Visual Studio Code
- Docker Desktop
- Windows 使用 winget，macOS 使用 Homebrew
- 支持 `list`、`plan`、`install`、`doctor`
- 支持 `java-basic`、`spring-backend`、`java-fullstack` 三种开发环境预设
- 支持 `--dry-run` 预览执行、安装超时和中文日志
- 默认内置清单，也可以通过 `--manifest` 使用自定义清单

## 使用

```text
jdb list
jdb version
jdb profiles
jdb list --profile java-basic
jdb plan
jdb plan --profile spring-backend
jdb install
jdb install --yes
jdb install --dry-run
jdb install --log 安装日志.txt
jdb doctor
```

`install` 默认会先检测已安装软件，再显示计划并要求输入 `yes`。`--yes` 只适合用户已经审阅清单后的自动化场景。

`--dry-run` 只生成安装计划，不检查或安装软件；`--log` 可指定安装日志文件，默认写入当前目录的 `jdb.log`。安装流程默认最多运行 30 分钟，超时后会终止后续命令并输出失败结果。

profile 说明：`java-basic` 适合基础 Java 开发；`spring-backend` 在基础环境上增加 Docker；`java-fullstack` 再增加 Visual Studio Code。不指定 profile 时使用清单中的全部软件。

## 发布

推送形如 `v0.1.0` 的 Git 标签后，GitHub Actions 会自动构建 Windows amd64/arm64、macOS amd64/arm64 二进制文件，并在 GitHub Release 中附带 SHA256 校验文件。

## 前置条件

- Windows 10/11：需要可用的 Windows Package Manager（winget）。
- macOS：需要可用的 Homebrew。
- 安装软件时可能需要系统管理员权限、网络连接和包管理器自身的协议确认。
- `doctor` 会实际执行包管理器版本检查；如果提示 winget 或 Homebrew 不可用，应先修复包管理器再运行安装。

## 安全边界

程序只执行清单中声明的程序和参数，不执行任意 shell 字符串；默认不安装；软件来源和版本由包管理器负责，使用前应审阅清单和包管理器输出。

## 开发

```text
go test ./...
go build -o jdb ./cmd/jdb
```

项目说明、注释和提交信息统一使用中文。
