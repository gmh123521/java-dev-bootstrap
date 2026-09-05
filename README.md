# Java Dev Bootstrap

Java Dev Bootstrap 是一个面向 Windows 和 macOS 的 Java 开发环境初始化 CLI。它可以检查当前环境、预览安装计划，并在用户确认后调用系统包管理器安装常用工具。

## 功能

- Eclipse Temurin JDK 21
- Git、Maven、Gradle
- IntelliJ IDEA Community、Visual Studio Code
- Docker Desktop
- Windows 使用 winget，macOS 使用 Homebrew
- 支持 `list`、`plan`、`install`、`doctor`
- 支持 `prerequisites` 检查 Windows 的 winget 或 macOS 的 Homebrew
- 支持 `setup` 检查前置条件并输出处理建议
- 支持 `java-basic`、`spring-backend`、`java-fullstack` 三种开发环境预设
- `plan` 会真实检测现有软件，只安装缺失项或升级版本过低的软件
- 支持 `--dry-run` 纯预览、安装超时和中文日志
- 支持 `--json` 输出软件清单、安装计划、安装报告和环境诊断结果
- 支持 `--retry 次数` 对失败安装逐项重试
- `doctor` 会诊断 Java、常用工具命令和环境变量
- 默认内置清单，也可以通过 `--manifest` 使用自定义清单

## 使用

```text
jdb list
jdb version
jdb profiles
jdb prerequisites
jdb setup
jdb list --profile java-basic
jdb plan
jdb plan --profile spring-backend
jdb install
jdb install --yes
jdb install --dry-run
jdb list --json
jdb plan --profile java-basic --json
jdb install --profile java-basic --dry-run --json
jdb prerequisites --json
jdb setup --json
jdb doctor --json
jdb version --json
jdb profiles --json
jdb install --profile java-basic --retry 2
jdb install --log 安装日志.txt
jdb doctor
```

`plan` 和 `install` 会优先执行工具的版本命令，再检查实际应用路径，最后使用 winget 或 Homebrew 记录补充判断。任一种可靠方式检测成功就会跳过安装。JDK 通过 `javac -version` 判断完整性和最低主版本，默认要求 JDK 21 及以上，不限制 Oracle、Temurin 等发行厂商；例如现有完整 JDK 23 会直接复用，只有 `java` 而没有 `javac` 的运行时不会被误判为 JDK。

IntelliJ IDEA 和 Docker Desktop 使用应用程序路径与包管理器记录检测。JetBrains 用户配置残留不会被当成已安装，单独存在 `docker` CLI 也不会被当成 Docker Desktop。

`install` 默认会先检测已安装软件，再显示计划并要求输入 `yes`。`--yes` 只适合用户已经审阅清单后的自动化场景。

真实安装完成后，程序会对本次成功执行的项目进行安装后复查。复查通过才会计入“复查通过”；如果安装命令返回成功但工具仍未检测到，命令会报告“复查失败”，便于用户检查 PATH、安装器结果或重新执行安装。

`prerequisites` 只检查操作系统、处理器架构和当前平台的包管理器是否可用，不会安装软件，不会修改 PATH、注册表或其他系统配置。Windows 新电脑需要先具备 winget（由 Microsoft 应用安装程序提供）；macOS 新电脑需要先安装 Homebrew。前置条件检查通过后，再执行 `doctor`、`plan` 和 `install`。

`setup` 是面向新电脑的引导命令。它会检查当前平台和包管理器是否准备好，并在缺失时给出安装或修复建议；它不会执行 `sudo`，不会自动安装 winget 或 Homebrew，也不会修改 PATH、注册表及其他系统配置。处理建议完成后重新执行 `jdb prerequisites`，检查通过再执行 `jdb doctor`。

`plan` 会执行只读检测，但不会安装软件，整次检测最多运行 2 分钟；`install --dry-run` 会执行同样的只读软件检测并生成命令预览，不会调用包管理器安装或修改系统。`--log` 可指定安装日志文件，默认写入当前目录的 `jdb.log`。一次完整安装默认最多运行 30 分钟，超时后会终止当前命令和后续安装并输出失败结果。

`--json` 可用于 `version`、`profiles`、`list`、`plan`、`install --dry-run`、配合 `--yes` 的真实 `install`，以及 `prerequisites`、`setup` 和 `doctor`。JSON 模式只输出机器可解析的 JSON；真实安装使用 `--json` 时必须同时使用 `--yes`，避免交互确认提示混入 JSON 输出。

命令参数会按命令校验：`--yes`、`--dry-run`、`--retry` 和 `--log` 只用于 `install`；`--profile` 和 `--manifest` 可用于需要软件清单的命令，不适用于 `version`、`profiles`、`prerequisites` 和 `setup`。不支持的组合会立即报错，不会执行系统检测或安装。

真实安装可以使用 `--retry 次数` 设置每个失败项目的额外重试次数，默认不重试。重试只针对当前失败项目，已经成功或跳过的项目不会重复执行；最终仍失败时才计入失败汇总。

profile 说明：`java-basic` 适合基础 Java 开发；`spring-backend` 在基础环境上增加 Docker；`java-fullstack` 再增加 Visual Studio Code。不指定 profile 时使用清单中的全部软件。

## 发布

推送形如 `v0.1.0` 的 Git 标签后，GitHub Actions 会自动构建 Windows amd64/arm64、macOS amd64/arm64 二进制文件，并在 GitHub Release 中附带 SHA256 校验文件。

## 前置条件

- Windows 10/11：需要可用的 Windows Package Manager（winget）。
- macOS：需要可用的 Homebrew。
- 安装软件时可能需要系统管理员权限、网络连接和包管理器自身的协议确认。
- `doctor` 会实际执行包管理器版本检查；如果提示 winget 或 Homebrew 不可用，应先修复包管理器再运行安装。

## 安装位置与环境变量

Java Dev Bootstrap 不强制指定安装目录。Windows 上的实际位置由 winget 清单和软件官方安装器决定，macOS 上由 Homebrew 决定。常见位置如下：

| 软件 | Windows 常见位置 | macOS 常见位置 |
| --- | --- | --- |
| JDK | `C:\Program Files\Eclipse Adoptium\jdk-21*` | `/Library/Java/JavaVirtualMachines/temurin-21.jdk` |
| Git | `C:\Program Files\Git` | Homebrew 前缀下的 `bin/git` |
| Maven/Gradle | `%LOCALAPPDATA%\Microsoft\WinGet\Packages` | Homebrew 前缀下的 Cellar |
| IntelliJ IDEA | `C:\Program Files\JetBrains` 或用户选择的目录 | `/Applications` |
| VS Code | `%LOCALAPPDATA%\Programs\Microsoft VS Code` | `/Applications` |
| Docker Desktop | `C:\Program Files\Docker\Docker` | `/Applications` |

程序当前不会主动覆盖 `JAVA_HOME`、`MAVEN_HOME`、`GRADLE_HOME` 或 `PATH`。部分官方安装器会自行更新 PATH，安装结束后通常需要重新打开终端。使用下面的命令检查最终状态：

```text
jdb doctor
```

`doctor` 会比较 `JAVA_HOME/bin/java` 与 PATH 中 `java` 的实际版本，并检查 Git、Maven、Gradle、VS Code 和 Docker 命令。发现问题时只输出建议，不调用 `setx`、不修改注册表，也不覆盖用户已有配置。

常见用户配置目录：

- Git：`%USERPROFILE%\.gitconfig`
- Maven：`%USERPROFILE%\.m2\settings.xml` 和 `%USERPROFILE%\.m2\repository`
- Gradle：`%USERPROFILE%\.gradle`
- IntelliJ IDEA：`%APPDATA%\JetBrains`
- VS Code：`%APPDATA%\Code\User`
- Docker：`%USERPROFILE%\.docker`

## 安全边界

程序只执行清单中声明的程序和参数，不执行任意 shell 字符串；默认不安装；软件来源和版本由包管理器负责，使用前应审阅清单和包管理器输出。

## 开发

```text
go test ./...
go build -o jdb ./cmd/jdb
```

项目说明、注释和提交信息统一使用中文。
