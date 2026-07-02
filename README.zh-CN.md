<p align="center">
  <img src="docs/assets/nextvibe-hero.svg" alt="NextVibe - 面向 AI 编程工具的项目导航" width="760">
</p>

<p align="center">
  <strong>面向 AI 辅助开发的 agent-native 项目导航工具。</strong>
</p>

<p align="center">
  <a href="README.md">English</a>
  ·
  <a href="docs/usage.md">使用说明</a>
  ·
  <a href="docs/release.md">发布流程</a>
  ·
  <a href="docs/roadmap.md">路线图</a>
</p>

<p align="center">
  <img alt="Go 1.22+" src="https://img.shields.io/badge/Go-1.22%2B-00ADD8?logo=go&logoColor=white">
  <img alt="Platforms" src="https://img.shields.io/badge/platform-Windows%20%7C%20macOS%20%7C%20Linux-2E7D32">
  <img alt="No model API key" src="https://img.shields.io/badge/model%20API%20key-not%20required-6A1B9A">
  <img alt="License" src="https://img.shields.io/badge/license-MIT-blue">
</p>

NextVibe 是一个本地 CLI，用来告诉编程 agent 下一步应该构建什么。它会扫描项目、识别可见的开发信号、推荐一个有边界的下一步任务，并检查任务是否具备预期产物。

它不替代 Codex、Claude Code 或 Cursor，而是给这些工具一个可靠的本地命令，让它们可以像调用 `git`、`go test`、`npm` 一样调用 `nextvibe`。

## 为什么需要 NextVibe

AI 编程工具很擅长搭脚手架，但项目完成第一波生成后，经常卡在下一步：API 合约、后端边界、数据模型、测试、部署，还是 agent 指令？

NextVibe 只回答一个窄问题：

> 编程 agent 下一步应该构建什么？

| 信号 | NextVibe 做什么 |
| --- | --- |
| 🧭 项目状态 | 扫描文件、目录、技术栈、测试和风险 |
| 🎯 下一任务 | 推荐一个带验收标准的有边界任务 |
| 🧱 任务边界 | 写入 `.nextvibe/current-task.md` 和任务文件 |
| ✅ 完成检查 | 检查必要产物和变更路径 |
| 🤖 Agent 适配 | 安装 Codex、Claude Code、Cursor 指令文件 |

## 快速开始

需要 Go 1.22 或更高版本。

```bash
go build ./cmd/nextvibe
```

Windows 会生成 `nextvibe.exe`。macOS 或 Linux 会生成 `./nextvibe`。

在项目根目录运行核心闭环：

```bash
nextvibe init
nextvibe install all
nextvibe scan --json
nextvibe suggest --json
nextvibe task --json
nextvibe check --json
```

文本输出支持英文和中文：

```bash
nextvibe scan --lang en
nextvibe scan --lang zh
nextvibe suggest --language zh
```

JSON 输出保持稳定字段名和结构，方便 agent 和自动化脚本解析。

## 完整本地流程

用下面的流程验证一个新项目是否已经可以被 agent 接管：

```bash
nextvibe init
nextvibe install all
nextvibe scan --json
nextvibe suggest --json
nextvibe task --json
nextvibe check --json
```

预期结果：

- `.nextvibe/` 存在并记录项目状态
- `AGENTS.md`、`CLAUDE.md`、`.cursor/rules/nextvibe.mdc` 和 Claude 命令文件存在
- `scan --json` 返回技术栈、风险、信号和可能的测试命令
- `suggest --json` 返回一个推荐任务
- `task --json` 创建或读取当前任务
- `check --json` 判断任务产物是否存在

## 核心命令

| 命令 | 作用 |
| --- | --- |
| `nextvibe init` | 创建 `.nextvibe/` 工作区 |
| `nextvibe scan` | 检查项目信号和当前阶段 |
| `nextvibe suggest` | 推荐下一个有边界的任务 |
| `nextvibe task` | 创建或读取当前任务文件 |
| `nextvibe check` | 检查基础完成信号 |
| `nextvibe install all` | 安装 Codex、Claude Code、Cursor 指令 |

面向 agent 的命令支持 `--json`。文本命令支持 `--lang en|zh` 或 `--language en|zh`。

## 会创建哪些文件

```text
.nextvibe/
  project.md
  stage.md
  roadmap.md
  current-task.md
  decisions.md
  agent-context.md
  tasks/

AGENTS.md
CLAUDE.md
.cursor/rules/nextvibe.mdc
.claude/skills/nextvibe/SKILL.md
.claude/commands/nv-suggest.md
.claude/commands/nv-check.md
```

已有文件会被保留。NextVibe 会追加或更新托管区域，而不是覆盖用户自己写的内容。

## JSON 输出示例

`nextvibe scan --json` 会返回项目状态：

```json
{
  "projectName": "example-project",
  "detectedStacks": ["go", "react"],
  "keyFiles": ["README.md", "go.mod", "package.json"],
  "keyDirectories": ["components", "mock", "src"],
  "testCommands": ["go test ./...", "npm test"],
  "signals": {
    "hasFrontend": true,
    "hasBackend": true,
    "hasMockData": true,
    "hasApiContract": false,
    "hasDatabaseSchema": false,
    "hasTests": false,
    "hasDockerfile": false,
    "hasAgentInstructions": false
  },
  "stage": {
    "id": "frontend-prototype-backend-incomplete",
    "label": "Frontend prototype exists, backend contract missing"
  },
  "risks": ["Mock data detected", "No API contract found"]
}
```

`suggest`、`task` 和 `check` 也提供稳定 JSON，便于 agent 和脚本使用。

## 规则系统

第一版规则是本地、确定性、可解释的。NextVibe 会读取仓库中的可见信号：

- `package.json`、`go.mod`、`pom.xml`、`build.gradle`
- `README.md`、`AGENTS.md`、`CLAUDE.md`
- `src/`、`app/`、`pages/`、`components/`
- `api/`、`server/`、`backend/`、`internal/`、`cmd/`
- `mock/`、`mocks/`
- `docs/api/`、`openapi.yaml`
- `migrations/`、`sql/`
- `test/`、`tests/`、`__tests__` 和常见测试文件后缀
- `Dockerfile`、`docker-compose.yml`

它也会检测简单的本地测试命令：

- Go modules：`go test ./...`
- 带 `test` script 的 Node 项目：`npm test`
- Maven 项目：`mvn test`
- Gradle 项目：`gradle test`

优先级规则从简单开始：

1. 缺少项目目标或实现结构：先写项目目标。
2. 前端存在但缺 API 合约：设计 API 合约。
3. API 合约存在但缺数据模型：设计数据模型。
4. 后端或业务代码存在但缺测试：补聚焦测试。
5. 缺部署配置：添加最小部署路径。
6. 缺 agent 集成文件：运行 `nextvibe install all`。

## 平台支持

NextVibe 目标支持 Windows、macOS 和 Linux。

CI 会在三个操作系统上运行测试，并交叉编译这些发布目标：

- `linux/amd64`
- `linux/arm64`
- `darwin/amd64`
- `darwin/arm64`
- `windows/amd64`
- `windows/arm64`

带版本号的 GitHub Release 会为 macOS 和 Linux 发布 `.tar.gz`，为 Windows 发布 `.zip`。详见 [docs/release.md](docs/release.md)。

## 产品边界

当前范围：

- 跨平台 Go CLI
- `.nextvibe/` 项目状态
- 项目扫描和阶段判断
- 下一任务推荐
- 当前任务生成
- 基础任务检查
- Codex、Claude Code、Cursor 集成文件
- 面向 agent 的稳定 JSON 输出
- 中英文文本输出

暂不包含：

- Web UI
- 云服务
- 账号系统
- 模型 API 集成
- MCP Server
- Prompt 市场
- 包管理器分发

MCP 以后可以作为另一种暴露本地协议的方式加入，但不应该改变核心原则：NextVibe 帮助 agent 导航本地项目状态，它本身不变成模型。

## 项目结构

| 路径 | 作用 |
| --- | --- |
| `cmd/nextvibe` | CLI 入口 |
| `internal/scanner` | 文件和目录扫描 |
| `internal/detector` | 技术栈、信号、阶段和测试命令检测 |
| `internal/planner` | 下一任务推荐规则 |
| `internal/taskgen` | 当前任务和任务文件生成 |
| `internal/checker` | 基础任务完成检查 |
| `internal/installer` | Agent 集成文件生成 |
| `docs/` | 使用、发布、路线、部署和产品说明 |

## 更多文档

- [使用说明](docs/usage.md)
- [发布流程](docs/release.md)
- [部署说明](docs/deployment.md)
- [路线图](docs/roadmap.md)
- [产品愿景](docs/vision.md)
- [Agent-native 设计](docs/agent-native-design.md)
