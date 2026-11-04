# ACP 会话控制与投递证据的实际边界

Status: resolved
Type: research
Labels: wayfinder:research
Parent: ../map.md
Assignee: chldu2000 (Codex, background research)
Blocked by: 01

## Question

ACP 标准与已选 Codex/Claude 适配器分别能提供哪些证据，区分输入未投递、可能投递、已消费、回合完成和执行已停止？断连、取消、插话竞争及恢复时，哪些结果不能从请求 ID、方法返回或历史重放推断？

核查 prompt/cancel/permission/配置确认及 steering 的具体语义，load/resume 是否可能推进自主工作，以及宿主文件/终端能力与 Agent 自执行的责任。区分标准保证、固定版本实现、未核实项；不测试真实账户或读取实际会话库。

## Comments

2026-11-04：从 [会话控制、授权与消息投递的责任](07-session-control-and-delivery.md)的具体接口问题形成研究票。之前的能力覆盖研究确认存在功能路径，但不足以证明投递去重、恢复与停止的语义；本票补充这些事实，不替用户决定交互偏好。

研究使用独立 worktree 与 `research/acp-session-control` 分支；结果保存为 `docs/research/acp-session-control.md`，完成后在此记录分支、提交与证据范围。

## Answer

2026-11-04：研究完成，结果见 [ACP 会话控制与投递证据的实际边界](../../../docs/research/acp-session-control.md)。研究分支 `research/acp-session-control`，提交 `0e70f2b2aa722b7ab21e47c0770e9cd46e8e2058`；独立 worktree 位于 `/private/tmp/aibo-rev-acp-session-control`。主工作区已按该提交复制研究文档，未合并或提交主分支。

关键结论：

- 请求 ID 只关联调用与响应，不提供跨连接重试幂等保证；v1 prompt 最终响应与 v2 草案的接受响应不能混用。
- prompt 结算、可接收下一输入、全部后台/原生活动停止是不同判断。一般错误、超时、断连、空历史不能证明输入未消费。
- 固定 Codex steering 实现的错误后回退分支存在未消费证明缺口，不能无条件以 `promptRequired` 自动重发；固定 Claude 可以强制结算 `cancelled`，底层 query 仍可能卡住。
- 配置请求、已确认配置及当前原生回合的实际参数不能合并；所查实现未提供配置与投递的统一原子合同。
- Codex resume 有目标自动启动的源码证据；Claude load/resume 可启动或复用原生 CLI。原生恢复不能默认作为纯历史读取。
- v1 的 always 选项由 Agent 记忆并决定范围，客户端回传本次 optionId；ACP 不要求跨 Agent 重放审批。ACP 也不提供进程沙箱，宿主只能直接控制自己执行的操作与自己授予的能力。
- MCP 配置接线不自动证明调用者身份，引用工具需要宿主自己的会话凭证与授权绑定。

研究区分当日官方规范、固定源码观察和待验证推论。仅进行了静态核查；未执行真实 Agent、访问账户或读取实际会话库。所列歧义和恢复条件进入会话控制决议与后续组合验收，不以研究完成冒充运行验证通过。
