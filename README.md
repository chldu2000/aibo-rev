# aibo-rev

现有 Aibo 的彻底重构项目。以已经可用于日常 vibe-coding 的桌面工作台为功能基线，重新梳理功能范围、模块职责与运行数据生命周期。

当前进入架构决策阶段，依据已勾选的核心功能，通过决策地图明确实现路线，尚未开始产品实现。“彻底重构”不预先等于重写全部代码、照搬全部旧合同或删除旧数据。终端优先、CLI 配合 GUI 阅读及 VS Code 扩展保留为历史备选，不再作为当前主线。

2026-11-04 阶段记录：全 ACP、框架中立、会话控制、历史引用、扩展生命周期和数据所有权等架构基线已形成；用户认可跨框架呈现原型基本通过。正式技术栈、用户界面与交互方案后续单独设计，暂不指定设计系统或呈现风格（2026-11-10 已确定技术栈与默认界面方案，见下方阅读入口）。原型代码与证据保留在 `prototype/presentation-conformance` 分支（`83ac7e8`），不作为正式实现选型。

## 阅读入口

- [当前方向](docs/adr/0002-aibo-refactor-direction.md)：以现有 Aibo 为基线推进彻底重构。
- [技术栈](docs/adr/0003-desktop-and-ui-stack.md)：MyGO 桌面容器、Go 宿主（2026-11-10；默认界面部分已由 ADR 0006 取代）。
- [原生界面](docs/adr/0006-native-ui.md)：默认界面用 MyGO 原生 UI，取消外部完整呈现包，工具面板为嵌入的独立 WebView（2026-11-11）。
- [共享交互状态与公开呈现协议](docs/adr/0004-presentation-protocol.md)：状态在 Go 宿主；版本化 JSON 协议现只用于工具面板（2026-11-11，经 ADR 0006 修订）。
- [默认外观](docs/adr/0007-liquid-glass-default-appearance.md)：Liquid Glass，玻璃只用于外框层（2026-11-12）。
- [进程、存储、ACP 接入、包与工具面板](docs/adr/0005-runtime-storage-and-packages.md)：单一应用进程、SQLite、自写 ACP 连接层、zip 包与 MCP 能力包、工具面板（2026-11-11；面板装载方式经 ADR 0006 修订）。
- [全 ACP 与框架中立架构地图](.scratch/aibo-rev-architecture/map.md)：本轮约束、已完成研究与逐项架构决策的入口。
- [完整版本功能 Checklist](docs/refactor-feature-inventory.md)：手动勾选完整版本必需的特性，包含已有功能、数据治理和新增候选。
- [数据生命周期盘点](docs/data-lifecycle-inventory.md)：数据放大来源、已有治理和仍需完成的能力。
- [Zed / ACP 实现研究](docs/zed-acp-research.md)：统一会话呈现的源码路径、适配边界及对重构的启发。
- [默认工作台界面与交互方案](docs/design/ui-interaction.md)：MyGO + WebView 下的布局、关键流程、运行控制与恢复交互，附可点击原型（草案）。
- [默认工作台界面硬性要求](docs/design/ui-requirements.md)：编号条款与验收条件，配套 [本地验证脚本](tools/ui-checks/README.md) 与 [设计令牌](docs/design/tokens/tokens.json)。
- [MyGo 桌面方案调研](docs/research/mygo-desktop-evaluation.md)：构建链、WebView 与原生 UI 路线、发布成熟度及 Aibo 的验证边界。
- [领域词汇](CONTEXT.md)：业务历史、恢复、审计、过程明细和派生数据的区别。
- [讨论记录](docs/discussion.md)：方向演进及尚未定案的取舍。

## 历史备选与证据

- [终端优先提案](docs/design.md)：此前候选方案。
- [CLI 执行与 GUI 历史阅读](docs/cli-gui-reading-proposal.md)：此前候选方案。
- [终端路线引用决定](docs/adr/0001-terminal-transcript-references.md)：此前语境下确认的取舍，不自动成为当前重构合同。
- [原仓库未提交文档补丁](docs/archive/aibo-design-changes.patch)：迁入前的原始文档证据，不含代码或个人配置改动。

## 与原仓库的关系

此前“保留完整插件平台与管理中心、简化内部机制”的提案已提交在 `aibo` 的 `docs/lightweight-architecture-proposal` 分支，提交为 `de70665`。它是前一条候选路线，不能把其中所有产品范围要求自动继承为本仓库的目标。

本仓库最初保存随后尚未提交的终端路线，现根据用户明确指示回到 Aibo 本体重构。原仓库工作区保持原样；归档补丁是证据副本。当前工作依据方向决定和功能盘点，历史方案不得替代当前范围确认。
