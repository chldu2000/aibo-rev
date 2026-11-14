# AGENTS.md

aibo-rev 是现有 Aibo 的彻底重构。架构与界面已经决定，下一步是产品实现。实现以已有决定为准；任何设计都不要从头再做。

## 开始实现前

按顺序读完以下材料，再动手：

1. [README](README.md)：分清当前方向与历史备选。
2. [CONTEXT.md](CONTEXT.md)：领域词汇。代码标识符、票据、测试名都用这里的术语。
3. 当前 ADR（先看 frontmatter 的 `status` 与修订说明，被取代的段落只作背景）：
   - [0002](docs/adr/0002-aibo-refactor-direction.md) 方向：主目标是降低维护成本，数据治理是主线。
   - [0006](docs/adr/0006-native-ui.md) 原生界面：模块划分、依赖方向、线程、panic 边界。
   - [0005](docs/adr/0005-runtime-storage-and-packages.md) 进程、SQLite、ACP 连接层、包与 MCP、工具面板。
   - [0007](docs/adr/0007-liquid-glass-default-appearance.md) Liquid Glass 外观。
   - [0004](docs/adr/0004-presentation-protocol.md) 只读仍有效的部分：§1 状态归属、§5 草稿同步、工具面板协议。
   - [0008](docs/adr/0008-windows-tool-panel-recovery.md) Windows 工具面板恢复；[0009](docs/adr/0009-official-agent-packages-and-node-runtime.md) 官方 Agent 包与 Node 运行时；[0010](docs/adr/0010-implementation-start.md) 开始实现的约定（ACP v1、代码位置、CI、第一条纵向链路）。
   - [0003](docs/adr/0003-desktop-and-ui-stack.md) 只有 MyGO 与 Go 宿主仍有效。0001 已被取代。
4. [功能 Checklist](docs/refactor-feature-inventory.md)：范围以已勾选的 82 项为准，未勾选的不实现、也不留占位。
5. 与任务相关的[架构决议](.scratch/aibo-rev-architecture/map.md)：每张已解决票的 `## Answer` 是行为合同（例如 07 投递与控制、08 目标与后台任务、10 数据所有权）。
6. 涉及界面时，读下文“界面设计来源”。
7. 写代码前读 [Clean Code 规范](docs/agents/clean-code.md)。它约束所有产品代码与测试。

## 界面设计来源（按优先级）

冲突时以靠前者为准：

1. [界面硬性要求](docs/design/ui-requirements.md)：编号条款，合入门槛。
2. [界面与交互方案](docs/design/ui-interaction.md)：设计理由与细节；§2.5 组件方案 A“玻璃胶囊”，§16 列出用户已确认的取舍。
3. [设计令牌](docs/design/tokens/tokens.json)：所有颜色、字号、间距、圆角的唯一来源。
4. 原型，只作示意：
   - [`prototypes/workbench-shell`](prototypes/workbench-shell/README.md)：最新、用户已逐项确认的外观与交互，包括玻璃外观、组件方案 A（`kit.go`）、玻璃菜单、模型 × 推理强度矩阵（`picker.go`）、会话行操作（`sessions.go`）、工作区分组（`workspaces.go`）、区域级 panic 边界（`shell.go`）。看外观与交互先看这里及其 `reports/` 截图。
   - [`prototypes/native-ui`](prototypes/native-ui/README.md)：会话流与输入区，以及令牌如何映射到 MyGO `ui.Theme`；目前唯一的验收目标。
   - [`docs/design/prototype/index.html`](docs/design/prototype/index.html)：WebView 时期的静态原型，只用来看信息架构与流程。

原型是一次性实验，产品代码不得 import 原型。可以借鉴其中的做法，但要按 Clean Code 规范重写。原型里的示例数据不构成要求。

## 识别过时内容

2026-11-11 起，以下内容是历史，不要实现：外部呈现包及其换包恢复、“重新载入界面 / 切回默认界面”、默认界面的页面心跳（工具面板自己的心跳仍保留，见 ADR 0005 §5、ADR 0008）、Svelte / Bits UI / virtua / markdown-it / Shiki、跨框架公开呈现协议（工具面板协议除外）、iframe 工具面板。部分文档的正文仍然提到这些，以 ADR 0006 为准。

## 冲突与未决

- 文档之间、文档与原型之间、需求与可行性之间出现冲突时，停下来向用户报告，写清两边的出处。不要自行选一边继续。
- 要推翻或绕过某个已接受的决定，先提出新 ADR 或修改要求文档，经用户确认后再实现。
- 仍未决定或未验证的事项：[12 号票](.scratch/aibo-rev-architecture/issues/12-architecture-acceptance-and-transition.md)（验收与切换）、[17 号票](.scratch/aibo-rev-architecture/issues/17-windows-presentation-validation.md)（Windows），以及各 ADR 的“尚未验证”一节。碰到这些时按证据实现，结果写回对应票或 ADR。
- 旧 `aibo` 仓库的代码、合同与存储格式不是约束，只作行为与失败经验的参考。不迁移旧数据，不读取或清理真实运行库。

## 交付门槛

- Go 1.27.3、MyGO v0.9.0、`CGO_ENABLED=0`，版本锁定。
- 合入界面改动前 `go -C tools/ui-checks run . all` 全部通过；`manual` 项在变更说明中逐条列出。接入方法见 [ui-checks README](tools/ui-checks/README.md)。
- 改了 `tokens.json` 后运行 `go -C tools/ui-checks run . tokens` 重新生成令牌，不要手改生成文件。
- 提交信息遵循 [Conventional Commits](https://www.conventionalcommits.org/)，语言不限；末尾用 `Co-Authored-By:` 署名提交的 agent 自己。

## Agent skills

### Issue tracker

Issues live in `.scratch/<feature>/`. Before reading or writing tickets,
read [issue tracker conventions](docs/agents/issue-tracker.md).

### Triage labels

Use the five default triage labels. Before triaging,
read [label mappings](docs/agents/triage-labels.md).

### Domain docs

Single-context: root `CONTEXT.md` and `docs/adr/`. Before exploring,
read [domain documentation rules](docs/agents/domain.md).
