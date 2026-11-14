# Aibo Rev 产品实现

Status: open
Labels: wayfinder:map

## Destination

按已接受的 ADR 与[架构图](../aibo-rev-architecture/map.md)决议实现 Aibo Rev，覆盖[功能 Checklist](../../docs/refactor-feature-inventory.md)已勾选的 81 项，并通过 [12 号票](../aibo-rev-architecture/issues/12-architecture-acceptance-and-transition.md)定义的验收。本图拆解实施任务，不重新决定架构。

## Notes

- 2026-11-15 依 [ADR 0010](../../docs/adr/0010-implementation-start.md) 建立。开工前读 [AGENTS.md](../../AGENTS.md) 列出的材料。
- 采用[本地 Markdown tracker](../../docs/agents/issue-tracker.md)：先 claim 再开工；票的 `Type:` 一般为 `task`。
- 每张票的完成条件包含测试；界面改动须通过 `go -C tools/ui-checks run . all`，`manual` 项在变更说明中列出。
- 碰到文档冲突或需要推翻已接受的决定时，停下来报告（AGENTS.md“冲突与未决”）。
- 第一个里程碑：01–05 完成，第一条纵向链路在 macOS 上可用；之前关闭架构图 12 号票。

## Decisions so far

<!-- 一行对应一张已解决的子票。 -->

## Not yet specified

- 第一条链路之后的拆分（Claude Code、引用、目标与子任务、包安装与升级、工具面板、搜索、诊断等），在 01–05 期间按 12 号票的验收场景补票。
