# 同一 ACP 连接中并发多个会话的可靠性（Claude Code）

Status: open
Type: prototype
Labels: wayfinder:prototype
Parent: ../map.md
Assignee: unassigned
Blocked by: none

## Question

用户 2026-11-12 决定正式版本先支持 Codex 与 Claude Code（F02.1、F02.3）。[ADR 0005](../../../docs/adr/0005-runtime-storage-and-packages.md#尚未验证) 要求每个正式支持的 Agent 验证同一 ACP 连接中并发多个会话的可靠性；Codex 已由 [19 号票](19-acp-same-connection-concurrency.md)验证。本票对 Claude Code 做同样的验证：不可靠时，该 Agent 包改为声明每会话独立进程。

复用 [`prototypes/acp-concurrency`](../../../prototypes/acp-concurrency/README.md) 的客户端与场景（S1 并发回合无串话、S2 各自 `cwd`、S3 定向取消、S4 适配器被强制结束后子进程退出），记录 Claude Code 与 ACP 适配器的实际版本、认证方式与协议版本。以 macOS 为主要验证平台（用户 2026-11-12 决定）。

## Comments

2026-11-12 新建。本机有 Claude Code 2.1.318；npm 上可见 `@agentclientprotocol/claude-agent-acp` 0.106.0 与 `@zed-industries/claude-code-acp` 0.16.2，实际采用哪个适配器在本票中确定并记录理由。

2026-11-15：适配器先用 `@agentclientprotocol/claude-agent-acp`（与 codex-acp 同属 `agentclientprotocol` 组织，07、14 号研究已固定其源码）。按 [ADR 0009](../../../docs/adr/0009-official-agent-packages-and-node-runtime.md) 一并验证：把适配器打包为单个 JS 文件、不含 Claude Agent SDK 的平台 CLI，经 `CLAUDE_CODE_EXECUTABLE` 指向本机 `claude` 后，S1–S4 是否全部成立；记录所用 Node 版本（适配器要求 ≥ 22）。
