# aibo-rev

现有 Aibo 的彻底重构项目。以已经可用于日常 vibe-coding 的桌面工作台为功能基线，重新梳理功能范围、模块职责与运行数据生命周期。

架构与界面已经决定，正在进入产品实现（[ADR 0010](docs/adr/0010-implementation-start.md)）：MyGO 原生界面 + Go 宿主，全部 Agent 经 ACP 接入，正式版本先支持 Codex 与 Claude Code，平台为 macOS arm64 与 Windows 11 x64。终端优先、CLI 配合 GUI 阅读、VS Code 扩展以及跨框架呈现包都是历史备选，不再作为主线。实现 agent 先读 [AGENTS.md](AGENTS.md)。

## 阅读入口

**范围与决定**

- [功能 Checklist](docs/refactor-feature-inventory.md)：范围以已勾选的 81 项为准。
- [方向](docs/adr/0002-aibo-refactor-direction.md)：降低维护成本，数据治理是主线。
- [原生界面](docs/adr/0006-native-ui.md)：模块划分、依赖方向、线程与 panic 边界。
- [进程、存储、ACP 接入、包与工具面板](docs/adr/0005-runtime-storage-and-packages.md)（经 ADR 0006、0008、0009 修订）。
- [Liquid Glass 外观](docs/adr/0007-liquid-glass-default-appearance.md)、[Windows 工具面板恢复](docs/adr/0008-windows-tool-panel-recovery.md)、[官方 Agent 包与 Node 运行时](docs/adr/0009-official-agent-packages-and-node-runtime.md)、[开始实现的约定](docs/adr/0010-implementation-start.md)。
- [共享交互状态](docs/adr/0004-presentation-protocol.md)：只有状态归属、草稿同步与工具面板协议仍有效；[技术栈](docs/adr/0003-desktop-and-ui-stack.md)：只有 MyGO 与 Go 宿主仍有效。
- [领域词汇](CONTEXT.md)。

**界面**

- [界面硬性要求](docs/design/ui-requirements.md)（合入门槛，配套 [ui-checks](tools/ui-checks/README.md) 与[设计令牌](docs/design/tokens/tokens.json)）。
- [界面与交互方案](docs/design/ui-interaction.md)。
- 原型（只作示意）：[`workbench-shell`](prototypes/workbench-shell/README.md)、[`native-ui`](prototypes/native-ui/README.md)。

**决策与任务**

- [架构图](.scratch/aibo-rev-architecture/map.md)：每张已解决票的答案是行为合同；未关的是 12 号（验收）与 20 号（Claude Code 并发）。
- [实施地图](.scratch/aibo-rev-implementation/map.md)：产品实现的任务。

**研究与盘点**

- [ACP 研究](docs/research/)：必需能力、会话控制、目标与子任务、模型与推理强度；[包生命周期证据](docs/research/package-lifecycle-evidence.md)；[MyGO 调研](docs/research/mygo-desktop-evaluation.md)。
- [数据生命周期盘点](docs/data-lifecycle-inventory.md)、[Zed / ACP 实现研究](docs/zed-acp-research.md)：旧版证据与研究，以架构图决议为准。

## 历史备选与证据

- [讨论记录](docs/discussion.md)：2026-11-01 的方向演进。
- [跨框架呈现研究](docs/research/framework-neutral-presentation.md)、[`stream-composer`](prototypes/stream-composer/README.md)：WebView 时期的研究与原型。
- [终端优先提案](docs/design.md)：此前候选方案。
- [CLI 执行与 GUI 历史阅读](docs/cli-gui-reading-proposal.md)：此前候选方案。
- [终端路线引用决定](docs/adr/0001-terminal-transcript-references.md)：此前语境下确认的取舍，不自动成为当前重构合同。
- [原仓库未提交文档补丁](docs/archive/aibo-design-changes.patch)：迁入前的原始文档证据，不含代码或个人配置改动。

## 与原仓库的关系

此前“保留完整插件平台与管理中心、简化内部机制”的提案已提交在 `aibo` 的 `docs/lightweight-architecture-proposal` 分支，提交为 `de70665`。它是前一条候选路线，不能把其中所有产品范围要求自动继承为本仓库的目标。

本仓库最初保存随后尚未提交的终端路线，现根据用户明确指示回到 Aibo 本体重构。原仓库工作区保持原样；归档补丁是证据副本。当前工作依据方向决定和功能盘点，历史方案不得替代当前范围确认。
