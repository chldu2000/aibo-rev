# 同一 ACP 连接中并发多个会话（Codex）

一次性实验，不是产品实现。回答 [19 号票](../../.scratch/aibo-rev-architecture/issues/19-acp-same-connection-concurrency.md)：ADR 0005 中“正式支持的 Agent 在同一连接中并发多个会话的可靠性”。`main.go` 是一个最小的 Go ACP 客户端（JSON-RPC 按行分隔），只依赖标准库。

## 运行

```sh
(cd adapter && npm install --ignore-scripts)          # @agentclientprotocol/codex-acp 2.6.0
go run . -sessions 6 -rounds 3 -out reports/run.json  # 真实调用，约 1.5 分钟
```

适配器用 `CODEX_PATH` 指向本机 `~/.local/bin/codex`（0.167.0，ChatGPT 登录），`INITIAL_AGENT_MODE=read-only`，推理强度 low。客户端不声明 fs/terminal 能力，权限请求一律拒绝（本次实验中没有收到权限请求）。完整协议记录在 `reports/<时间>/wire.jsonl`（不提交）。

## 环境

macOS 27.0 arm64；Node（适配器）；codex-cli 0.167.0；`@agentclientprotocol/codex-acp` 2.6.0（ACP protocolVersion 1）；2026-11-12。

## 结果

一个适配器进程，下挂一个 `codex app-server`，同时承载 6 个会话，每个会话一个独立工作目录。

| 场景 | 结果 |
| --- | --- |
| 建 6 个会话 | 4.9 s（串行 `session/new`） |
| S1：6 个会话同时回合，3 轮（18 个回合） | 全部 `end_turn`，回复只含本会话标记，无串话；首轮 4.6–13.8 s（冷启动），后两轮 1.6–3.8 s |
| S2：6 个会话同时读取各自 `cwd/note.txt` | 全部读到本目录内容，没有读到其他目录（每个回合 1 个 `tool_call`） |
| S3：两个会话同时数数，取消其中一个 | 被取消的回合约 10 ms 后以 `cancelled` 结束；另一个照常 `end_turn` 并输出到 DONE；被取消的会话随后再次提问，回合正常 |
| S4：回合进行中强制结束适配器（`SIGKILL`） | 两个未完成的 `session/prompt` 都因连接 EOF 立即结束（没有 stopReason）；`codex app-server` 在 2 s 内随 stdin EOF 退出，无残留 |
| 无 `sessionId` 或会话未知的通知 | 0 |

注意：
- 适配器初始化时只公布 `api-key` 一种认证方式，但使用本机 Codex 已有的 ChatGPT 登录也能正常建会话和回合。宿主不能只凭 `authMethods` 判断是否需要登录。
- S2 有 3 个回合先输出“我会读取 note.txt”，再输出文件内容。验收比较回复时不能要求逐字相等。
- S4 中适配器被杀后，未完成回合的结果在客户端只表现为连接断开。按 07 号决议，这些投递应记为“结果未知”，不能自动重发。

## 结论

Codex（经 codex-acp 2.6.0）在同一连接中并发多个会话可靠，可以按 ADR 0005 的“包 + 工作区”共用一个进程，不需要声明为每会话独立进程。这个结论只适用于这一组合。Claude（`claude-agent-acp`）等其他 Agent 需要分别验证，结论记入各自的已验证组合。
