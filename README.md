# aibo-rev

现有 Aibo 的彻底重构项目。以已经可用于日常 vibe-coding 的桌面工作台为功能基线，重新梳理功能范围、模块职责与运行数据生命周期。

当前阶段是功能与数据盘点，尚未开始产品实现。“彻底重构”不预先等于重写全部代码、照搬全部旧合同或删除旧数据。终端优先、CLI 配合 GUI 阅读及 VS Code 扩展保留为历史备选，不再作为当前主线。

## 阅读入口

- [当前方向](docs/adr/0002-aibo-refactor-direction.md)：以现有 Aibo 为基线推进彻底重构。
- [完整版本功能 Checklist](docs/refactor-feature-inventory.md)：手动勾选完整版本必需的特性，包含已有功能、数据治理和新增候选。
- [数据生命周期盘点](docs/data-lifecycle-inventory.md)：数据放大来源、已有治理和仍需完成的能力。
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
