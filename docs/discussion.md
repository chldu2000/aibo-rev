# 讨论记录：从平台简化到终端优先

日期：2026-11-01。本文是讨论摘要，不是逐字聊天 transcript；只把用户已明确接受的内容标记为已确认。

## 前一条候选路线

用户最初希望降低 Aibo 的维护成本，同时保留统一工作台、跨 Agent 完整历史引用、完整插件平台和 Agent 管理中心。允许一次破坏性合同升级与历史迁移，管理功能可使用配置编辑器或嵌入终端，外部界面可运行隔离前端代码，缺失功能可由宿主默认面板提供。

这一轮已提交到原 `aibo` 仓库的 `docs/lightweight-architecture-proposal` 分支，提交 `de70665`。本仓库没有把旧路线的全部范围当成终端路线的既定要求。

## 用户提出新的方向

用户询问：如果只保留“可跨 Agent 引用会话”的核心特性，是否能通过 Ghostty 等终端直接运行和渲染各 Agent 的 CLI，并尽量保留会话列表、Git 和文件浏览等扩展性；还是仍需通过 ACP 自行渲染。

讨论建议把终端交互与会话共享分开：原生 CLI 通过 PTY 运行，会话桥从可靠的原生记录获取历史，工作台管理目录与引用。ACP 可以用于结构化接入，但不是保留原生 CLI 界面的必要条件。Ghostty 尚未选定，原项目现成的 xterm.js 路径也可用于验证。

## 已确认的两轮回答

第一轮，用户接受：

1. 点击引用时冻结历史，而非严格拦截原生 CLI 提交时点。
2. 区分“可运行 Agent”和“支持完整会话引用的 Agent”；无可靠历史入口时允许手动导入。

用户补充了关键意图：传递会话时直接传递原始 transcript；默认可能只传一部分，完整内容可通过 MCP 工具等方式提供。

第二轮，用户接受：

1. 默认最近几轮原文，保留关联工具记录，设总大小上限，允许选择全文。
2. 引用插入当前 CLI 输入框，由用户检查和发送；未验证粘贴兼容性的 Agent 提供复制入口，不自动模拟 Enter。

这些决定已整理为 [ADR 0001](adr/0001-terminal-transcript-references.md)。原仓库尚未提交的 ADR 0012 早于第二轮回答，因此仍把摘录及交付方式列为待定；本仓库已更新。

## 尚未授权或确定的事项

尚未选择终端引擎、ACP 适配范围、桌面技术栈或首批支持平台，未批准开始产品实现。建议的下一步是用两个 Agent 验证原生 CLI 与固定 transcript 引用闭环，不代表该验证已经执行。

用户随后要求在 `aibo-dev` 下创建独立 `aibo-rev` 仓库保存这轮未提交设计与讨论。本仓库据此建立，原仓库工作区改动未删除。

## 调查依据

以下是本轮讨论查阅的资料，并非运行验收证据：

- 原仓库终端实现：`aibo-plugins/plugins/terminal/README.md`，说明 xterm.js、portable-pty、内存滚动记录和进程生命周期。
- 原仓库宿主合同：`aibo/docs/tool-view-contract.md`。
- [ACP Overview](https://agentclientprotocol.com/protocol/v1/overview)：客户端、会话更新和权限交互。
- [Ghostty 架构](https://ghostty.org/docs/about)：终端核心与原生 GUI 的关系。
- [xterm.js PTY 接线](https://xtermjs.org/docs/guides/encoding/)：终端输入输出连接。
- [Claude Code Hooks](https://code.claude.com/docs/en/hooks)：会话身份、transcript 路径及写入时序限制。

上述外部资料的版本会变化；实施前须对选定版本重新核实。这里没有承诺各 Agent 都能提供完整或实时落盘的 transcript。
