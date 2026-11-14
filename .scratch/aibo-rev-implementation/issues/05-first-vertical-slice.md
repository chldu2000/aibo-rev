# 第一条纵向链路：Codex 会话的发送、审批、取消与重启

Status: open
Type: task
Blocked by: 02, 03, 04

## Question

在原生界面中跑通：添加工作区 → 信任 → 用 Codex 新建会话 → 发送 → 流式回复 → 审批 → 取消 → 重启后历史仍在（ADR 0010 第 5 条）。会话控制按 ADR 0005 §3（每会话一个控制 goroutine）与 [07 号决议](../../aibo-rev-architecture/issues/07-session-control-and-delivery.md#answer)；界面以 `prototypes/workbench-shell` 为外观参照、按 Clean Code 规范重写。为涉及的 UI 要求补写 `uitest` 验收用例。
