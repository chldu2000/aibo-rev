# ACP 对必需会话能力的实际覆盖

Status: resolved
Type: research
Labels: wayfinder:research
Parent: ../map.md
Assignee: chldu2000 (Codex research agent)
Blocked by: none

## Question

以当前 Codex 与 Claude Code 的 ACP 接入实现为样本，已勾选的会话能力分别由标准 ACP、草案、命名空间扩展、Agent 自身或宿主提供什么？哪些仍缺乏证据？

重点核查恢复、审批/追问、模式和模型参数、Fast/用量、命令/Skills、插话、持续目标、完整子 Agent 过程、后台任务读取/取消、MCP 历史读取，以及原始 transcript 是否可经 ACP 提供。区分协议能表达、适配器已实现、真实运行已验证三层证据；没有实测不得写成已验收。

所有 Agent 接入只能走 ACP；可以研究 ACP 适配器内部如何桥接原生能力，但不能建议宿主新增原生协议旁路。仅报告事实、缺口与扩展成本，不决定缺功能是否可接受。

范围依据：F02–F09、F11–F12、F22.3；资料入口：[Zed 研究](../../../docs/zed-acp-research.md)。

研究资产：`docs/research/acp-required-capabilities.md`；上下文：`research/acp-required-capabilities` 分支。

## Comments

2026-11-01：已由 chldu2000 的 Codex 研究会话认领。从 `86b8efd` 建立独立分支，工作树为 `/private/tmp/aibo-rev-acp-research.gGtKYN`。本票只核查事实；root 汇总研究资产并记录 Answer，支持范围决策留给后续讨论。

## Answer

2026-11-01：完成固定源码版本的能力核查，见 [ACP 必需能力覆盖研究](../../../docs/research/acp-required-capabilities.md)。研究分支 `research/acp-required-capabilities`，最终提交 `92b5ed6a84d0da41b174fb33e5f4fc811ba9996a`，已将同版文档收录到主工作区。

- 当前 Codex / Claude 适配器已迁至 `agentclientprotocol` 组织；基础会话、结构化交互和配置存在共同协议与实现基础，插话与子 Agent 也有具体实现路径。
- 持续目标和后台任务的重要路径依赖实验性 AIR 扩展；实现该命名空间的兼容不等于冒用客户端身份，但上游第三方稳定支持承诺未核实。扩展采用与维护责任需用户决策。
- 目标动作和预算语义不一致：所查 Codex 路径有 pause/resume，Claude 路径只有 set/clear；不能把清除目标或取消回合当作等价暂停。
- 未核实两家共有的原始 transcript 导出及后台任务全文分页补读合同。协议内容可观察性与原始记录完整性必须分开。
- 任务活动可发生在没有前台 prompt 的时候；宿主不能只在当前 prompt 存在时接收业务事件。

本票完成的是事实调查，没有运行真实 Agent 或改变功能承诺。上述问题已分别属于 [统一 ACP 接入与 Agent 能力承诺](04-agent-capability-contract.md)、[业务历史与跨 Agent 引用的权威来源](06-history-reference-authority.md)、[会话控制、授权与消息投递的责任](07-session-control-and-delivery.md)、[持续目标、子 Agent 与后台任务的职责](08-goals-and-background-work.md)，不重复建票。
