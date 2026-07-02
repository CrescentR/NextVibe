# NextVibe 开发路径与流程

## 1. 项目定位

**NextVibe** 是一个面向 AI 编程用户、vibe coder、独立开发者的 **agent-native project navigation tool**。

它的目标不是替代 Codex、Claude Code、Cursor 等 AI 编程工具，而是成为这些工具在开发过程中可以直接调用的本地项目导航工具。

NextVibe 应该像 `git`、`npm`、`python`、`go test`、`chrome-devtools` 一样自然地融入 agent 工作流。

## 2. 核心问题

很多人在使用 AI 编程工具时，通常可以很快完成：

* 项目初始化
* 框架搭建
* 前端页面生成
* 基础组件生成

但项目做到一半后，往往会卡住：

* 不知道下一步该做什么
* 不知道该先补 API、数据库、业务闭环、测试还是部署
* 不知道当前项目缺什么
* 不知道该如何约束 agent 的开发范围
* 不知道如何让项目从 demo 推进到可交付版本

NextVibe 要解决的就是这个问题。

## 3. 一句话定位

> NextVibe tells your coding agent what to build next.

## 4. Slogan

> From vibe coding to vibe shipping.

中文理解：

> 不只是把项目搭起来，而是把项目推到能上线。

## 5. 产品原则

NextVibe 必须严格遵守以下原则：

1. 不接入任何大模型 API。
2. 不要求用户配置 OpenAI、Anthropic 或其他模型 Key。
3. 不做另一个 AI Chat 工具。
4. 不做 AI IDE。
5. 不做项目管理平台。
6. 不要求用户复制 Prompt 再粘贴给 agent。
7. 不直接生成大量业务代码。
8. 不替代 Codex、Claude Code、Cursor。
9. 主要输出应适合 agent 读取，而不是只适合人阅读。
10. 核心能力应该通过 CLI、JSON、配置文件、Skill、Rules、MCP 暴露给编程 agent。

## 6. 正确工作流

NextVibe 的理想工作流不是：

```bash
human runs nextvibe suggest
human copies prompt
human pastes prompt into Codex
```

而是：

```bash
agent reads project instructions
agent calls nextvibe scan --json
agent calls nextvibe suggest --json
agent calls nextvibe task --json
agent edits code
agent calls nextvibe check --json
agent summarizes result
```

也就是说：

```text
Codex / Claude Code / Cursor = reasoning + editing + command execution
NextVibe = project state + next-step protocol + task navigation
```

## 7. 核心命令

推荐 CLI 命令如下：

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

其中：

```bash
nextvibe suggest
```

比：

```bash
nextvibe next
```

更合适，因为 `nextvibe next` 读起来重复。

## 8. `.nextvibe/` 工作区设计

NextVibe 应该在项目根目录生成 `.nextvibe/` 工作区。

```text
.nextvibe/
├── project.md
├── stage.md
├── roadmap.md
├── current-task.md
├── decisions.md
├── agent-context.md
├── tasks/
└── state.json
```

各文件职责：

| 文件                 | 作用                |
| ------------------ | ----------------- |
| `project.md`       | 记录项目是什么、目标用户、核心功能 |
| `stage.md`         | 记录当前项目阶段          |
| `roadmap.md`       | 记录项目推进路线          |
| `current-task.md`  | 记录当前任务            |
| `decisions.md`     | 记录重要架构和产品决策       |
| `agent-context.md` | 给 agent 读取的项目上下文  |
| `tasks/`           | 保存历史任务            |
| `state.json`       | 给 agent 读取的结构化状态  |

`.nextvibe/` 目录应该可以被 Git 追踪。

原因：

* 项目状态应该和代码一起演进
* 不同 agent 可以共享上下文
* 不同开发者可以理解当前项目阶段
* 项目推进过程可以被复盘

## 9. 推荐 Go 项目结构

```text
nextvibe/
├── cmd/
│   └── nextvibe/
│       └── main.go
├── internal/
│   ├── cli/
│   ├── scanner/
│   ├── detector/
│   ├── planner/
│   ├── taskgen/
│   ├── checker/
│   ├── installer/
│   ├── workspace/
│   ├── protocol/
│   └── output/
├── docs/
│   ├── vision.md
│   ├── roadmap.md
│   ├── protocol.md
│   └── agent-native-design.md
├── examples/
│   └── sample-project/
├── go.mod
├── README.md
└── LICENSE
```

目录职责：

| 目录                   | 职责                                   |
| -------------------- | ------------------------------------ |
| `cmd/nextvibe`       | CLI 入口                               |
| `internal/cli`       | 命令解析和命令分发                            |
| `internal/scanner`   | 扫描项目文件和目录                            |
| `internal/detector`  | 判断技术栈、项目信号和项目阶段                      |
| `internal/planner`   | 根据阶段生成下一步建议                          |
| `internal/taskgen`   | 生成当前任务和任务文件                          |
| `internal/checker`   | 检查当前任务是否完成                           |
| `internal/installer` | 生成 Codex / Claude Code / Cursor 集成文件 |
| `internal/workspace` | 创建和维护 `.nextvibe/` 工作区               |
| `internal/protocol`  | 定义 JSON 输出结构                         |
| `internal/output`    | 负责 text / json 输出                    |

## 10. 开发阶段规划

整体开发路线分为 7 个阶段：

```text
Phase 0：项目定位与协议设计
Phase 1：CLI MVP
Phase 2：项目扫描与阶段判断
Phase 3：任务生成与任务检查
Phase 4：Agent 集成
Phase 5：规则系统增强
Phase 6：MCP Server
Phase 7：开源发布与生态扩展
```

## Phase 0：项目定位与协议设计

### 目标

确定 NextVibe 的核心定位、工作区结构、JSON 协议和 agent 调用方式。

### 产物

```text
README.md
docs/vision.md
docs/agent-native-design.md
docs/protocol.md
.nextvibe/ 文件格式设计
JSON 输出协议设计
```

### 重点

这个阶段不急着写大量代码，先明确：

* NextVibe 是什么
* NextVibe 不是什么
* agent 如何调用 NextVibe
* `.nextvibe/` 目录如何保存状态
* JSON 输出如何稳定给 agent 使用

## Phase 1：CLI MVP

### 目标

完成最小可运行 CLI。

### 需要实现的命令

```bash
nextvibe init
nextvibe scan
nextvibe suggest
nextvibe task
nextvibe check
```

每个核心命令都应该支持：

```bash
--json
```

因为 NextVibe 的主要使用者不是人，而是 Codex、Claude Code、Cursor 这类编程 agent。

### 验收标准

执行：

```bash
go build ./cmd/nextvibe
```

然后可以运行：

```bash
./nextvibe init
./nextvibe scan
./nextvibe scan --json
./nextvibe suggest
./nextvibe suggest --json
./nextvibe task
./nextvibe task --json
./nextvibe check
./nextvibe check --json
```

Windows 下：

```bash
go build -o nextvibe.exe ./cmd/nextvibe
nextvibe.exe init
```

## Phase 2：项目扫描与阶段判断

### 扫描关键文件

第一版先扫描这些文件：

```text
package.json
go.mod
pom.xml
build.gradle
Cargo.toml
Dockerfile
docker-compose.yml
README.md
AGENTS.md
CLAUDE.md
.cursor/rules
```

### 扫描关键目录

```text
src/
app/
pages/
components/
api/
mock/
mocks/
server/
backend/
internal/
cmd/
docs/
test/
tests/
__tests__/
migrations/
sql/
```

### 输出项目信号

扫描结果不应该只告诉 agent “发现了什么文件”，而应该转换成结构化信号。

示例：

```json
{
  "signals": {
    "hasFrontend": true,
    "hasBackend": false,
    "hasMockData": true,
    "hasApiContract": false,
    "hasDatabaseSchema": false,
    "hasTests": false,
    "hasDockerfile": false,
    "hasAgentInstructions": false
  }
}
```

### 项目阶段定义

第一版可以定义这些阶段：

```text
empty-project
scaffold-created
frontend-prototype
frontend-prototype-api-missing
backend-created-contract-missing
contract-created-db-missing
core-flow-incomplete
core-flow-complete-tests-missing
tests-ready-deploy-missing
ready-to-ship
```

示例判断：

```text
有 package.json + pages/components + mock，但没有 api/server/backend
=> frontend-prototype-api-missing
```

## Phase 3：任务生成与任务检查

### `nextvibe suggest`

根据扫描结果推荐下一步。

示例：

```bash
nextvibe suggest --json
```

返回：

```json
{
  "currentStage": "frontend-prototype-api-missing",
  "recommendation": {
    "id": "design-api-contract",
    "title": "Design the first API contract",
    "priority": "high",
    "reason": "Frontend pages and mock data exist, but there is no API contract."
  },
  "artifacts": [
    "docs/api/openapi.yaml",
    "docs/api/api-design.md"
  ],
  "acceptanceCriteria": [
    "Each core page has a corresponding API endpoint",
    "Each endpoint has request and response examples",
    "No backend implementation is added in this task",
    "Mock data is used only as reference"
  ],
  "agentInstructions": {
    "allowedActions": [
      "read frontend pages",
      "read mock data",
      "create docs/api/openapi.yaml",
      "create docs/api/api-design.md"
    ],
    "forbiddenActions": [
      "do not implement backend logic",
      "do not modify UI components",
      "do not delete mock data"
    ]
  }
}
```

### `nextvibe task`

`nextvibe task` 应该基于 `suggest` 生成当前任务。

生成：

```text
.nextvibe/current-task.md
.nextvibe/tasks/001-design-api-contract.md
```

任务结构：

```markdown
# Task 001: Design the first API contract

## Background

## Goal

## Allowed Changes

## Forbidden Changes

## Artifacts

## Acceptance Criteria

## Agent Instructions

## Check Command
```

### `nextvibe check`

`nextvibe check` 用来检查当前任务是否完成。

第一版先支持基础检查：

* 要求文件是否存在
* 禁止修改路径是否被改动
* 当前任务状态是否可关闭
* 是否缺少关键产物

后续可以增强：

* build 是否通过
* 测试是否通过
* lint 是否通过
* OpenAPI 是否合法
* README 是否更新

## Phase 4：Agent 集成

这是 NextVibe 的差异化核心。

NextVibe 不能要求用户长期手动运行命令并复制粘贴给 agent。

正确方式是让 Codex / Claude Code / Cursor 自动知道应该调用它。

### 4.1 Codex 集成

命令：

```bash
nextvibe install codex
```

生成或更新：

```text
AGENTS.md
```

AGENTS.md 中加入 NextVibe section：

````markdown
<!-- NEXTVIBE:START -->

## NextVibe Workflow

When the user asks to continue the project, decide the next step, recover direction, or proceed with broad development, use NextVibe first.

Run:

```bash
nextvibe scan --json
nextvibe suggest --json
nextvibe task --json
````

Use the returned task boundaries.

Do not expand scope beyond the current task.

After editing, run:

```bash
nextvibe check --json
```

<!-- NEXTVIBE:END -->

````

如果已有 `AGENTS.md`，不要覆盖用户内容，只更新 `NEXTVIBE` 区域。

### 4.2 Claude Code 集成

命令：

```bash
nextvibe install claude
````

生成：

```text
CLAUDE.md
.claude/skills/nextvibe/SKILL.md
.claude/commands/nv-suggest.md
.claude/commands/nv-check.md
```

Claude Skill 核心说明：

```markdown
# NextVibe Skill

Use this skill when the user asks:

- continue this project
- what should I do next
- 下一步做什么
- 继续推进项目
- 项目卡住了
- 帮我规划接下来的开发

Workflow:

1. Run `nextvibe scan --json`
2. Run `nextvibe suggest --json`
3. Run `nextvibe task --json`
4. Respect task boundaries
5. Edit only allowed files
6. Run `nextvibe check --json`
7. Summarize what changed
```

### 4.3 Cursor 集成

命令：

```bash
nextvibe install cursor
```

生成：

```text
.cursor/rules/nextvibe.mdc
```

规则内容：

````markdown
---
description: Use NextVibe to determine what to build next in this project.
alwaysApply: true
---

Before continuing broad project development, inspect NextVibe state.

Run:

```bash
nextvibe scan --json
nextvibe suggest --json
nextvibe task --json
````

Respect:

* allowedActions
* forbiddenActions
* artifacts
* acceptanceCriteria

After implementation, run:

```bash
nextvibe check --json
```

````

## Phase 5：规则系统增强

MVP 阶段规则可以写死，后续应该支持配置化。

### 配置文件

支持：

```text
.nextvibe/config.yaml
````

示例：

```yaml
project:
  type: fullstack
  preferred_stack:
    frontend: react
    backend: go-gin
    database: postgres

workflow:
  require_api_contract_before_backend: true
  require_tests_before_deploy: true
  require_docker_before_ship: true

agents:
  default: codex
  strict_task_boundary: true
```

### 自定义规则目录

支持：

```text
.nextvibe/rules/
├── springboot.yaml
├── go-gin.yaml
├── nextjs.yaml
└── tauri.yaml
```

Go Gin 示例：

```yaml
stages:
  - router
  - handler
  - service
  - repository
  - model
  - migration
  - tests
  - docker
```

Spring Boot 示例：

```yaml
stages:
  - dto
  - entity
  - mapper
  - service
  - controller
  - validation
  - tests
  - docs
```

这样 NextVibe 可以适配不同技术栈，而不是固定死一套流程。

## Phase 6：MCP Server

MCP 不建议第一版做，但它是后续高级形态。

命令：

```bash
nextvibe mcp
```

暴露 tools：

```text
nextvibe.scan_project
nextvibe.get_stage
nextvibe.get_suggestion
nextvibe.get_current_task
nextvibe.check_task
nextvibe.update_task_status
```

暴露 resources：

```text
nextvibe://project
nextvibe://stage
nextvibe://roadmap
nextvibe://current-task
nextvibe://decisions
```

暴露 prompts：

```text
nextvibe-suggest
nextvibe-check
nextvibe-ship
```

需要注意：

```text
MCP 只是让 agent 更方便调用 NextVibe。
MCP 不是让 NextVibe 自己接模型。
```

## Phase 7：开源发布与生态扩展

### 发布方式

优先支持：

```bash
go install github.com/yourname/nextvibe/cmd/nextvibe@latest
```

后续支持：

```bash
brew install nextvibe
scoop install nextvibe
npm install -g nextvibe
```

Windows 用户可以通过 GitHub Releases 下载：

```text
nextvibe-windows-amd64.exe
```

### GitHub Release 文件

每个版本发布：

```text
nextvibe-darwin-arm64
nextvibe-darwin-amd64
nextvibe-linux-amd64
nextvibe-windows-amd64.exe
```

### 示例项目

建议提供：

```text
examples/
├── react-mock-only/
├── go-gin-empty/
├── springboot-no-tests/
├── tauri-prototype/
└── fullstack-ready/
```

每个示例都应该可以运行：

```bash
nextvibe scan --json
nextvibe suggest --json
```

## 11. 版本规划

### v0.1.0：CLI MVP

目标：项目能跑。

包含：

```text
init
scan
suggest
task
check
--json
.nextvibe 工作区
README
```

不包含：

```text
agent install
MCP
复杂规则
UI
```

### v0.2.0：Agent Integration

包含：

```text
install codex
install claude
install cursor
install all
AGENTS.md 生成
CLAUDE.md 生成
Claude Skill 生成
Cursor Rules 生成
```

这是非常关键的版本。

### v0.3.0：Better Checks

包含：

```text
任务完成检查
禁止路径检查
产物检查
Git diff 检查
build/test 命令检测
```

### v0.4.0：Configurable Rules

包含：

```text
.nextvibe/config.yaml
不同技术栈规则
自定义阶段
自定义验收标准
```

### v0.5.0：MCP Server

包含：

```text
nextvibe mcp
MCP tools
MCP resources
MCP prompts
```

### v1.0.0：Stable Protocol

包含：

```text
稳定 JSON schema
稳定 .nextvibe 文件格式
稳定 CLI 命令
完善文档
完善示例
多平台 release
```

## 12. 实际开发流程

### 第一步：初始化仓库

```bash
mkdir nextvibe
cd nextvibe
git init
go mod init github.com/yourname/nextvibe
```

提交：

```bash
git add .
git commit -m "chore: initialize nextvibe project"
```

### 第二步：创建 CLI 骨架

任务：

```text
创建 Go CLI 项目骨架，实现 main.go 和基础命令分发。
不要实现完整业务逻辑。
```

提交：

```bash
git commit -m "feat: add cli skeleton"
```

### 第三步：实现 workspace

任务：

```text
实现 nextvibe init，创建 .nextvibe 工作区。
```

提交：

```bash
git commit -m "feat: add workspace initialization"
```

### 第四步：实现 scanner

任务：

```text
实现项目文件和目录扫描，支持 text/json 输出。
```

提交：

```bash
git commit -m "feat: add project scanner"
```

### 第五步：实现 detector

任务：

```text
根据扫描结果判断技术栈、项目信号和项目阶段。
```

提交：

```bash
git commit -m "feat: detect stack and project stage"
```

### 第六步：实现 planner

任务：

```text
根据项目阶段生成下一步建议。
```

提交：

```bash
git commit -m "feat: suggest next development task"
```

### 第七步：实现 taskgen

任务：

```text
根据 suggest 结果生成 current-task 和 tasks/001-xxx.md。
```

提交：

```bash
git commit -m "feat: generate current task"
```

### 第八步：实现 checker

任务：

```text
检查当前任务要求的产物是否存在，输出检查结果。
```

提交：

```bash
git commit -m "feat: check current task completion"
```

### 第九步：实现 installer

任务：

```text
实现 install codex / claude / cursor / all。
```

提交：

```bash
git commit -m "feat: add agent integrations"
```

### 第十步：完善 README 和 docs

提交：

```bash
git commit -m "docs: add usage and agent-native design"
```

## 13. GitHub Issues 建议

可以一开始就创建这些 issues：

```text
#1 Initialize Go CLI project structure
#2 Implement .nextvibe workspace initialization
#3 Implement project scanner
#4 Add JSON output protocol
#5 Detect project stack and signals
#6 Detect project development stage
#7 Implement next task suggestion
#8 Generate current task files
#9 Implement task completion checker
#10 Generate AGENTS.md for Codex
#11 Generate CLAUDE.md and Claude Skill
#12 Generate Cursor Rules
#13 Add README and usage examples
#14 Add sample projects
#15 Add GitHub Actions build
#16 Add release workflow
#17 Design MCP server
#18 Implement MCP tools
```

## 14. 分支策略

一个人开发时，不需要复杂 Git Flow。

建议：

```text
main         稳定可运行
dev          日常开发
feat/xxx     单功能分支
```

实际流程：

```bash
git checkout -b feat/scanner

# 开发
go test ./...

git add .
git commit -m "feat: add project scanner"

git checkout dev
git merge feat/scanner
```

dev 稳定后：

```bash
git checkout main
git merge dev
git tag v0.1.0
```

## 15. 提交规范

建议使用 Conventional Commits。

示例：

```text
feat: add project scanner
fix: avoid overwriting existing AGENTS.md
docs: add agent-native workflow
refactor: split planner rules
test: add scanner tests
chore: add github actions
```

类型：

```text
feat      新功能
fix       修复
docs      文档
test      测试
refactor  重构
chore     工程配置
```

## 16. README 推荐结构

README 建议这样组织：

```markdown
# NextVibe

Agent-native project navigator for AI coding tools.

Know what to build next — no extra model API, no chat UI, no copy-paste workflow.

## Why

AI coding tools are great at generating code, but projects often get stuck after scaffolding and prototyping.

NextVibe gives coding agents a reliable way to understand project state and decide the next actionable task.

## What NextVibe is

- A local CLI tool
- A project state protocol
- An agent integration layer
- A next-step navigator for AI coding workflows

## What NextVibe is not

- Not an AI chat app
- Not another coding agent
- Not a model API wrapper
- Not a project management platform

## Quick Start

## Agent-native workflow

## Commands

## JSON Output

## Codex Integration

## Claude Code Integration

## Cursor Integration

## Roadmap
```

## 17. 当前最小闭环

当前阶段不要先纠结 MCP、官网、插件市场。

第一步只做：

```text
v0.1.0：本地 CLI MVP
```

最小闭环：

```bash
nextvibe init
nextvibe scan --json
nextvibe suggest --json
nextvibe task --json
nextvibe check --json
```

只要这个闭环跑通，项目就成立了。

因为这已经证明：

```text
agent 可以调用 NextVibe
NextVibe 可以返回当前项目状态
agent 可以根据返回结果继续开发
agent 可以检查任务是否完成
```

第二个版本再接：

```bash
nextvibe install codex
nextvibe install claude
nextvibe install cursor
```

这个时候 NextVibe 就不只是 CLI，而是一个真正的：

```text
agent-native project navigation layer
```

## 18. 最终路线总结

按这个顺序推进：

```text
1. 明确定位：agent-native，不接模型，不做聊天
2. 创建 Go CLI 项目骨架
3. 实现 .nextvibe 工作区
4. 实现项目扫描 scan
5. 实现 JSON 输出协议
6. 实现阶段判断 detector
7. 实现下一步建议 suggest
8. 实现当前任务 task
9. 实现任务检查 check
10. 实现 Codex / Claude / Cursor install
11. 增强规则系统
12. 增加示例项目
13. 增加 GitHub Actions 和 Release
14. 后续做 MCP Server
15. v1.0 固化协议和文档
```

## 19. 核心结论

> 先做 CLI 闭环，再做 agent 集成，最后做 MCP。

NextVibe 的核心价值不是生成代码，而是让 AI 编程工具在已有项目中知道：

```text
项目现在在哪
之前做了什么
当前缺什么
下一步该做什么
哪些文件可以改
哪些事情不能做
完成后如何检查
```

项目后续npm安装：

Go 写核心 CLI
npm 主包提供 nextvibe 命令
平台 npm 子包携带 Go 二进制
JS wrapper 自动选择当前平台二进制并转发参数

用户最终体验：

npm install -g nextvibe
nextvibe init
nextvibe install codex
nextvibe scan --json


最终目标是：

> 让 Codex / Claude Code / Cursor 像调用 `git`、`npm`、`python` 一样调用 NextVibe，然后自然推进项目开发。

## 20. 开发进度记录机制

从现在开始，开发过程应当同时维护两类记录：

1. `.nextvibe/tasks/`：记录当前聚焦任务的边界、验收标准和禁止事项。
2. `docs/development-log.md`：记录每次开发的实际过程、验证命令、结果和下一步。

每次推进项目时，建议遵循这个流程：

```text
1. 运行 nextvibe scan --json / suggest --json / task --json
2. 根据建议收束成一个聚焦任务
3. 在 .nextvibe/current-task.md 和 .nextvibe/tasks/ 中记录任务边界
4. 按任务边界开发，不扩展到无关功能
5. 运行必要的测试、构建或检查命令
6. 运行 nextvibe check --json
7. 把本次目标、改动、验证结果和遗留问题写入 docs/development-log.md
```

开发日志不替代 README 或正式文档。它的作用是让后续 agent 和开发者快速理解：

```text
这次为什么做
实际改了什么
哪些命令验证过
哪些事情故意没有做
下一次应该从哪里继续
```

记录应当简短、可复盘，并优先引用实际文件和命令。
