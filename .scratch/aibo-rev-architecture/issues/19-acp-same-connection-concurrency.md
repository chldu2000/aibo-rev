# 同一 ACP 连接中并发多个会话的可靠性（Codex）

Status: resolved
Type: prototype
Labels: wayfinder:prototype
Parent: ../map.md
Assignee: claude（2026-11-12，用户同意使用本机安装的 codex）
Blocked by: none

## Question

[ADR 0005](../../../docs/adr/0005-runtime-storage-and-packages.md#尚未验证) 按“包 + 工作区”共用一个 Agent 进程，要求验证：正式支持的 Agent 在同一 ACP 连接中并发多个会话是否可靠；不可靠的 Agent 改为声明每会话独立进程。本票用本机 Codex 与 `@agentclientprotocol/codex-acp` 取证：

- 多个会话同时回合时，更新是否只送到所属会话，回合是否都正常结束。
- 同一进程内不同会话使用不同工作目录时，工具读取是否落在各自目录。
- 取消一个会话的回合，其他会话是否不受影响；被取消会话随后能否继续。
- 子进程退出时，未完成的请求如何结束。

不验证原生 transcript、子 Agent、目标等其他能力；这些已有研究票。

## Comments

2026-11-12 新建并认领。实验代码放 `prototypes/acp-concurrency`。

## Answer

Codex 在同一连接中并发多个会话可靠，可以按 ADR 0005 共用“包 + 工作区”进程，不需要声明每会话独立进程。证据见 [`prototypes/acp-concurrency`](../../../prototypes/acp-concurrency/README.md)（codex-cli 0.167.0、codex-acp 2.6.0、macOS）：

- 一个适配器加一个 `codex app-server` 承载 6 个会话；3 轮共 18 个并发回合全部正常结束，无串话，无未路由通知。
- 不同 `cwd` 的会话并发读文件，各自读到本目录内容。
- 取消一个会话约 10 ms 后以 `cancelled` 结束，另一会话不受影响，被取消的会话可继续提问。
- 强制结束适配器时，未完成回合只表现为连接断开（按 07 号决议记为结果未知）；app-server 随 stdin EOF 退出，无残留。

附带发现：适配器只公布 `api-key` 认证方式，但本机已有的 ChatGPT 登录可直接使用，宿主不能仅凭 `authMethods` 判断是否需要登录。结论只适用于这一组合，其他 Agent 待各自验证后计入已验证组合。
