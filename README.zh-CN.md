# NextVibe

[English](README.md)

NextVibe 是一个面向 AI 辅助开发者的 agent-native 项目导航工具。

它不替代 Codex、Claude Code 或 Cursor，而是给这些编程 agent 一个可靠的本地命令，用来理解项目状态、选择有边界的下一步任务，并检查任务是否完成。

不需要模型 API Key。不提供额外聊天界面。不要求复制粘贴 prompt。

安装一次之后，让你的编程 agent 像调用 `git`、`go test`、`npm` 一样调用 `nextvibe`。

## 为什么需要它

AI 编程工具可以很快完成原型、脚手架或前端页面，但项目做到一半后，常常会卡在“下一步该做什么”：

- 先补 API 合约，还是先写后端边界
- 先做数据模型，还是先补测试
- 项目是否已经具备部署路径
- agent 这次到底能改哪些文件
- 如何避免一次任务扩散成无关重构

NextVibe 只回答一个窄问题：

> 编程 agent 下一步应该构建什么？

它通过本地项目扫描、规则化阶段判断、稳定 JSON 输出，以及 `.nextvibe/` 下的任务边界文件来回答这个问题。

## 产品原则

- NextVibe 不调用 OpenAI、Anthropic 或任何其他模型 API。
- NextVibe 不要求用户配置模型 Key。
- NextVibe 不是另一个 AI Chat 应用。
- NextVibe 不是让人复制 prompt 的提示词生成器。
- NextVibe 是一个 CLI 加 agent 集成层，主要输出机器可读结果。
- Markdown 文件用于保存项目状态和兜底文档，不是主要交互方式。

理想工作流是：

```text
agent 读取项目指令
agent 调用 nextvibe 命令
agent 解析 JSON
agent 修改代码
agent 调用 nextvibe check
agent 更新项目状态
```

而不是：

```text
人手动运行 nextvibe
人复制 prompt
人粘贴给 agent
```

## 构建

需要 Go 1.22 或更高版本。

```bash
go build ./cmd/nextvibe
```

在 macOS 或 Linux 上会生成：

```bash
./nextvibe
```

在 Windows 上会生成：

```powershell
.\nextvibe.exe
```

示例：

```powershell
.\nextvibe.exe scan --json
```

## 平台支持

NextVibe 目标是支持 Windows、macOS 和 Linux。

CI 会在三个系统上运行测试，并交叉编译这些发布目标：

- `linux/amd64`
- `linux/arm64`
- `darwin/amd64`
- `darwin/arm64`
- `windows/amd64`
- `windows/arm64`

带版本号的 GitHub Release 会为 macOS 和 Linux 发布 `.tar.gz`，为 Windows 发布 `.zip`。

维护者发布流程见 [docs/release.md](docs/release.md)。

## 从 Release 安装

从 GitHub Release 页面下载与你系统匹配的压缩包，解压后把 `nextvibe` 放到 `PATH` 中。

Windows 二进制名称是：

```text
nextvibe.exe
```

macOS 和 Linux 二进制名称是：

```text
nextvibe
```

## 核心命令

```bash
nextvibe init
nextvibe scan
nextvibe suggest
nextvibe task
nextvibe check
nextvibe install codex
nextvibe install claude
nextvibe install cursor
nextvibe install all
```

面向 agent 的核心命令支持稳定 JSON 输出：

```bash
nextvibe scan --json
nextvibe suggest --json
nextvibe task --json
nextvibe check --json
```

`init` 和 `install` 也支持 `--json`，便于自动化。

## 第一次接入项目

在目标项目根目录运行：

```bash
nextvibe init
nextvibe install all
```

这会创建或更新：

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

已有文件会被保留。NextVibe 会追加或更新带标记的区域，而不是覆盖用户自己写的内容。

## Agent 工作流

当用户说“继续这个项目”“下一步做什么”“项目卡住了”时，编程 agent 应先运行：

```bash
nextvibe scan --json
nextvibe suggest --json
nextvibe task --json
```

然后严格遵守返回的任务边界：

- 允许修改的文件或动作
- 禁止事项
- 验收标准
- 预期产物

完成修改后运行：

```bash
nextvibe check --json
```

## JSON 输出示例

`nextvibe scan --json` 会返回项目状态：

```json
{
  "projectName": "example-project",
  "detectedStacks": ["go", "react"],
  "keyFiles": ["README.md", "go.mod", "package.json"],
  "keyDirectories": ["components", "mock", "src"],
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

`nextvibe suggest --json` 返回推荐任务。`nextvibe task --json` 创建或读取当前任务。`nextvibe check --json` 检查当前任务的基础完成信号。

## 规则系统

第一版规则是本地、确定性、可解释的。NextVibe 会读取仓库中可见的信号，例如：

- `package.json`、`go.mod`、`pom.xml`、`build.gradle`
- `README.md`、`AGENTS.md`、`CLAUDE.md`
- `src/`、`app/`、`pages/`、`components/`
- `api/`、`server/`、`backend/`、`internal/`、`cmd/`
- `mock/`、`mocks/`
- `docs/api/`、`openapi.yaml`
- `migrations/`、`sql/`
- `test/`、`tests/`、`__tests__` 和常见测试文件后缀
- `Dockerfile`、`docker-compose.yml`

优先级规则从简单开始：

1. 缺少项目目标或实现结构：先写项目目标。
2. 前端存在但缺 API 合约：设计 API 合约。
3. API 合约存在但缺数据模型：设计数据模型。
4. 后端或业务代码存在但缺测试：补聚焦测试。
5. 缺部署配置：添加最小部署路径。
6. 缺 agent 集成文件：运行 `nextvibe install all`。

## 命名与重命名

项目名、CLI 命令名和工作区目录集中在：

```text
internal/brand/brand.go
```

如果未来重命名项目，先更新这些常量，再重新生成文档和集成文件。

## 当前范围

第一版范围：

- 跨平台 Go CLI
- `.nextvibe/` 项目状态目录
- 项目扫描
- 规则化技术栈和阶段判断
- 下一步建议
- 当前任务生成
- 基础任务检查
- Codex、Claude Code、Cursor 集成文件
- 面向 agent 的稳定 JSON 输出
- 面向 Windows、macOS、Linux 的 CI 和 release 产物

暂不包含：

- Web UI
- 云服务
- 账号系统
- 模型 API 集成
- MCP Server
- Prompt 市场

MCP 以后可以作为另一种暴露同一本地协议的方式加入，但不应该改变核心原则：NextVibe 帮助 agent 导航本地项目状态，它本身不变成模型。

## 相关文档

- [使用说明](docs/usage.md)
- [发布流程](docs/release.md)
- [部署说明](docs/deployment.md)
- [开发路线与流程](docs/development-roadmap.md)
- [开发日志](docs/development-log.md)
