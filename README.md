# Java Dev Bootstrap

Java Dev Bootstrap 是一个面向 Windows 和 macOS 的 Java 开发环境初始化 CLI。它可以检查当前环境、预览安装计划，并在用户确认后调用系统包管理器安装常用工具。

## 功能

- Eclipse Temurin JDK 21
- Git、Maven、Gradle
- IntelliJ IDEA Community、Visual Studio Code
- Docker Desktop
- Windows 使用 winget，macOS 使用 Homebrew
- 支持 `list`、`plan`、`install`、`doctor`
- 默认内置清单，也可以通过 `--manifest` 使用自定义清单

## 使用

```text
jdb list
jdb plan
jdb install
jdb install --yes
jdb doctor
```

`install` 默认会先检测已安装软件，再显示计划并要求输入 `yes`。`--yes` 只适合用户已经审阅清单后的自动化场景。

## 前置条件

- Windows 10/11：需要可用的 Windows Package Manager（winget）。
- macOS：需要可用的 Homebrew。
- 安装软件时可能需要系统管理员权限、网络连接和包管理器自身的协议确认。

## 安全边界

程序只执行清单中声明的程序和参数，不执行任意 shell 字符串；默认不安装；软件来源和版本由包管理器负责，使用前应审阅清单和包管理器输出。

## 开发

```text
go test ./...
go build -o jdb ./cmd/jdb
```

项目说明、注释和提交信息统一使用中文。
