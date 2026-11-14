---
status: accepted
date: 2026-11-15
---

# 开始产品实现的约定

用户于 2026-11-15 同意以下约定（Agent 包与 Node 运行时见 [ADR 0009](0009-official-agent-packages-and-node-runtime.md)）。

1. **不等 12 号票关闭再开工。** 按已有决议先做第一条纵向链路；[12 号票](../../.scratch/aibo-rev-architecture/issues/12-architecture-acceptance-and-transition.md)的端到端验收规格与实现并行编写，在第一个里程碑之前关闭。
2. **ACP 只实现 v1。** 官方 schema 的 v1 为 Latest、v2 为 Draft；codex-acp 2.6.0 使用 v1。从 v1 schema 生成 Go 类型，生成代码按协议版本放在独立的包中，以后加 v2 不改动 v1。
3. **代码位置。** 产品代码放在本仓库，单个 Go module（`github.com/chldu2000/aibo-rev`）。ADR 0006 的四个模块（运行核心、交互层、界面包、桌面宿主）各为一个顶层包目录；依赖方向（ADR 0006 约束 1）由检查强制，不靠约定。原型保留在 `prototypes/`，产品代码不得 import。
4. **实施地图。** 实施任务在新的地图 [`.scratch/aibo-rev-implementation/`](../../.scratch/aibo-rev-implementation/map.md) 中拆解；架构图不再增加实施票。
5. **第一条纵向链路。** 添加工作区 → 信任 → 用 Codex 新建会话 → 发送 → 流式回复 → 审批 → 取消 → 重启后历史仍在。覆盖 ACP 连接层、SQLite 先落库再发送、会话控制 goroutine 与原生界面。
6. **CI。** GitHub Actions：macOS arm64 上运行 `go vet`、`go test` 与 `go -C tools/ui-checks run . all`；Windows x64 上编译并运行 `go test`。Windows 的实机复核仍在验收节点进行（12 号票）。
