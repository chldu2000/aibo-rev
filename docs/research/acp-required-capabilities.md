# ACP 对 Aibo Rev 必需会话能力的覆盖

研究日期：2026-11-01。对应“ACP 对必需会话能力的实际覆盖”研究票。本文提供事实与缺口，不决定哪些 Agent 必须具备哪些功能，也不改变 [Checklist](../refactor-feature-inventory.md)。

## 证据范围

- ACP 官网当前区分 v1 Latest 与 v2 Draft；本文以 v1 的通用能力为基线，草案与专有扩展单列。v2 改变了 prompt 完成与消息更新语义，不能把两版合同混用。[v2 迁移规范](https://agentclientprotocol.com/protocol/v2/migration)
- 旧 `zed-industries/codex-acp` 已明确迁往 `agentclientprotocol/codex-acp`；新适配器在自己的进程内桥接 Codex App Server，客户端仍通过 ACP 接入。[迁移说明](https://github.com/zed-industries/codex-acp/blob/296069e841634cd4bb9bc4515602d836e49231ec/README.md)
- Codex 样本：[`927c4802f83c`](https://github.com/agentclientprotocol/codex-acp/commit/927c4802f83c0859b439e7d641b5eb692560e5df)，2026-10-30；源码包版本 `2.5.0`。[package.json](https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/package.json)
- Claude 样本：[`eb78afa8556b`](https://github.com/agentclientprotocol/claude-agent-acp/commit/eb78afa8556b8fc4082495e48e3fd1762fb6a50f)，2026-10-30；仓库已迁至 `agentclientprotocol/claude-agent-acp`，源码包版本 `0.99.1`，在适配器内使用 Claude Agent SDK。[package.json](https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/package.json)、[README](https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/README.md)
- 这是固定提交的源码与文档核查；没有安装、启动 Agent 或访问真实会话库。源码版本号不证明这些提交已进入稳定发布，更不证明在用户账户、模型与 macOS/Windows 上均已验收。

## 核心发现

基础会话、结构化内容、审批、模型配置与 MCP 接线已有共同协议和实现基础。高级能力也不是空白：两家当前适配器已有插话、子 Agent 历史，以及目标/后台任务的桥接代码。但其协议稳定性、协商入口和控制语义不一致。全 ACP 可成为单一接入路线；“任意 ACP Agent 均完整覆盖 Checklist”没有证据支撑。

以下“已实现”只表示所查版本存在代码路径；不表示实测通过，也不保证所有运行条件可用。“未核实”不等于不存在。

## 能力矩阵

| 必需能力 | 协议与 Codex 样本 | Claude 样本 | 对架构决策的事实输入 |
| --- | --- | --- | --- |
| 创建、列出、恢复会话（F03） | 初始化声明 `loadSession`；实现 load 并重放历史，也有 list/resume 路径。[C 会话](https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/src/CodexAcpServer.ts#L798) | 同样声明 load/list/resume；load 调用历史读取与重放。[L 会话](https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/acp-agent.ts#L3061) | ACP load 是可协商能力；需要宿主自己的稳定身份和恢复失败状态 |
| 正文、思考、图片、工具结果（F04/F06） | README 和初始化实现包含文本、图片、上下文及多种工具事件。[C 内容](https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/README.md#features) | README 与初始化包含图片、上下文和工具事件。[L 内容](https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/acp-agent.ts#L3087) | 保留结构化内容；文件/目录作为上下文怎样截取仍需宿主定义 |
| 多选项权限、追问（F05） | 有标准权限请求、用户问题到 form elicitation 的映射，另有增强展示元数据。[C 追问](https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/src/CodexElicitationHandler.ts#L158) | AskUserQuestion 和 SDK elicitation 都有桥接；含多选与自定义答案处理。[L 追问](https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/elicitation.ts#L165) | 通用表单可共用，但支持的字段、取消和模式转换不能由 UI 猜测 |
| 模型、推理、模式（F08） | 通过 session config options 设置；源代码维护模型与模式状态。[C 配置](https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/src/CodexAcpServer.ts#L2753) | 通过同类配置入口设置，SDK 确认后更新配置。[L 配置](https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/acp-agent.ts#L9396) | 可用通用配置接口；具体选项及可写状态随会话变化 |
| Fast 服务档位（F08.3） | `fast-mode` 配置，模型 `additionalSpeedTiers` 决定实际能否使用 fast；支持 boolean 或 select。[C Fast](https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/src/FastModeConfig.ts) | `fast` 配置，读取 SDK 的实际 fast 状态和禁用原因。[L Fast](https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/acp-agent.ts#L11625) | 同一用户能力的原生 ID 与条件不同；不能仅看到开关就宣称已生效 |
| 用量/上下文（F08.5） | 有 token usage 事件及专用额度元数据，不能将所有额度视为标准字段。[C 功能](https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/README.md#features) | 源码发送 `usage_update`，读取 SDK 统计。[L 用量](https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/acp-agent.ts#L7346) | 会话总量、上下文占用与目标预算是不同数据 |
| 命令、Skills（F04.9） | 已公开命令及已配置 Skills；通过 available commands 进入客户端。[C 命令](https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/README.md#features) | 读取自定义命令；有 available commands 更新及 Skill 元数据。[L 命令](https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/acp-agent.ts#L11118) | “可发现并调用”不等于 CLI 全命令对等；路径、类别等可能属扩展 |
| 执行中插话（F09.6） | `_session/steering` 是 pre-standard 扩展；advertise `_meta.steering.supported`。[C 插话](https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/docs/steering.md) | 有相同方法与协商标记。[L 握手](https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/acp-agent.ts#L3114) | 普通 prompt 队列与插话要分开；必须处理输入被注入、新建回合或未消费的不同结果 |
| 持久队列（F09.1–F09.5） | 插话支持并不提供 Aibo 所需的完整持久队列合同 | 声明原生 promptQueueing，但不证明 Aibo 的附件持久化、删除、暂停和重启语义 | 宿主投递合同仍待决，不能把原生队列标记当作全部需求已满足。[L 标记](https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/acp-agent.ts#L3077) |
| 持续目标（F11） | AIR goal 扩展实现 set/pause/resume/clear，映射原生 thread goal。[C Goal](https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/src/GoalExtension.ts) | AIR goal 扩展仅 set/clear；映射 SDK active_goal 与原生 /goal。[L Goal](https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/goal-extension.ts) | F11.3 的暂停/恢复存在明确差异，详见下节 |
| 子 Agent 过程与恢复（F12.1–F12.3） | 有标准草案子会话路由、历史重放及动态定向 cancel 能力；另有旧 AIR 路径。[C 子会话代码](https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/src/subagents/StandardSubagentEventRouter.ts) | 原生子会话、独立 transcript 重放、父子消息关系与定向取消都有实现说明；缺原生元数据时报告缺失。[L 子会话](https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/README.md#subagent-sessions) | 需要按协商后的具体合同接入，不能仅解析普通工具卡推断完整子历史 |
| 后台任务状态与取消（F12.4/F12.6） | AIR `asyncTasks` 提供后台终端任务事件和定向 stop。[C 后台任务](https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/docs/air-extensions.md#async-tasks) | AIR `asyncTasks` 提供任务状态、进度及 stop；子 Agent、Monitor 等有不同归属或排除规则。[L 后台任务](https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/async-tasks.ts) | “后台任务”不是单一通用进程类；控制能力需逐任务判断 |
| 后台任务完整输出读取（F12.5） | 后台命令沿原工具/终端通道发送输出 | 任务更新可提供摘要及 `outputFilePath` | 未核实存在两家通用、可分页的历史任务全文读取合同；路径提示不等于通过 ACP 得到全文。[L 输出合同](https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/docs/air-extensions.md#task-id-and-output-path) |
| MCP 按需读取引用全文（F07.6） | 支持客户端传入 stdio/HTTP MCP server；显式拒绝 SSE 与 ACP transport。[C MCP](https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/src/McpServerConfig.ts) | 接受客户端 MCP server 配置，有 HTTP/SSE 支持声明。[L MCP](https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/acp-agent.ts#L3091) | Aibo 可提供自己的只读历史工具；授权、引用版本、分页和完整性由其工具合同定义 |
| 原始 transcript 获取 | load 重放的是适配器转换后的 ACP 内容 | 适配器读取原生 transcript 后转换并重放 | 本轮未核实两家共有的逐字节原始 transcript 导出方法。不能将 ACP 重放标为原始 transcript。[C load](https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/src/CodexAcpServer.ts#L1872)、[L load](https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/acp-agent.ts#L3270) |

离线业务历史、搜索、引用版本、草稿、包升级、主题/呈现、窗口布局和日志是宿主产品能力，不能由以上 Agent 矩阵替代其设计。ACP 的 session load 也不会自动给 Aibo 建立独立持久历史。[Session Setup](https://agentclientprotocol.com/protocol/v1/session-setup)

## AIR 扩展不能当作普通 ACP 基线

两仓库公开文档把这套合同明确称为 **JetBrains AIR 专属、实验性扩展**，以 `_meta.jetbrains.air` 命名。代码按握手的元数据判断 AIR，而非仅按客户端显示名称判断；这是技术门控事实，不代表上游已承诺第三方客户端兼容。[C 扩展代码](https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/src/AirExtension.ts#L78)、[L 扩展代码](https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/air-extension.ts#L120)

具体入口是 `initialize.clientCapabilities._meta.jetbrains.air`，内含 `version: 1` 和 `capabilities` 字符串数组。单项能力要求有效整数版本至少为 1 且包含准确能力名；goal 则在判定为 AIR 后由 Agent 侧公布其动作列表。没有该标记的普通 ACP 客户端不会收到 goal/asyncTasks 的相同信息。[C 协商文档](https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/docs/air-extensions.md#negotiation)、[L 协商文档](https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/docs/air-extensions.md#negotiation)

协商该命名空间的兼容不等于冒用客户端身份；aibo 应保留真实 `clientInfo`，并仅在实际实现扩展合同后声明对应能力。上游是否为第三方提供稳定兼容承诺仍未核实。后续决策须评估兼容采用现有公开合同、推动中立扩展，或在独立 ACP 适配器中提供自身版本化扩展的维护成本；任何路径仍由 ACP 接入，不增加宿主原生协议旁路。

## 已核实的关键差异

**目标动作与预算。** Codex goal 代码提供四种控制动作，快照类型可包含 tokenBudget、tokensUsed、timeUsedSeconds；控制 set 请求只接受 objective，不能据此宣称已支持设置预算。Claude 当前只接受 set/clear，其 SDK 映射只填 objective、active、iterations、lastReason、createdAt，没有填充该类型中可选的预算/用量字段。不能用“字段在类型中存在”推导运行时会报告，也不能把取消回合或清除目标当作暂停。[C Goal 类型](https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/src/GoalExtension.ts)、[L Goal 映射](https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/goal-extension.ts)

**活动不一定属于正在等待响应的 prompt。** 两家 goal 文档都说明目标可跨 prompt 存续，完成一个 prompt 后仍可能出现后续自主活动。Claude 后台任务还可能触发无前台 prompt 的自主回合。只在当前 prompt 存在时接收并持久化事件，会遗漏合法活动。[C Goal 生命周期](https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/docs/air-extensions.md#lifecycle)、[L 自主回合](https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/docs/air-extensions.md#autonomous-turn)

**插话结果不能简单算作发送成功。** Codex 文档区分 injected、startedNewTurn、promptRequired、failed；opt-in 的 `idleBehavior: promptRequired` 可让无可接收回合时保留输入给宿主正常投递。需要把这一结果连接到队列和投递证据，不能把所有返回都当作一个新完成回合。[插话合同](https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/docs/steering.md)

**子 Agent 文档与实现存在漂移。** Codex 的 `docs/subagent-sessions.md` 仍写早期事件名和不支持定向 cancel，但当前 `StandardSubagentEventRouter` 已发送 `subagent_update`，动态声明 cancel，并实现只中止目标子会话及其后代工作。不能将旧文档适用于所有路径；应对具体握手、版本和路由做协议夹具验证。[旧说明](https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/docs/subagent-sessions.md)、[当前路由](https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/src/subagents/StandardSubagentEventRouter.ts#L560)

**后台任务不保证同一种输出合同。** Codex 的任务输出由已有命令卡对应的流承载；Claude 任务可能仅有结构化摘要与输出文件路径。前者不能自动证明可补读丢失的输出，后者也不能自动授权宿主读取任意路径。如何通过 ACP 获得完整可保存输出仍是需要明确的能力。[C 任务](https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/docs/air-extensions.md#async-tasks)、[L 任务](https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/docs/air-extensions.md#task-id-and-output-path)

**v2 不能直接解决这些差异。** 两家存在 v2 相关实现，但规范仍为 Draft；Claude README 明确 steering 和 AIR 扩展当前为 v1 only，而 Codex 源码按版本生成扩展集合。上线组合需按固定版本验证，不能统一假设“升级 v2 后全部可用”。[Claude v2](https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/README.md#experimental-acp-v2)、[Codex v2 扩展选择](https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/src/CodexAcpServer.ts#L917)

## 后续决策需要回答的问题

1. F11.3 是否要求首发两个 Agent 都有等价的原生目标暂停/恢复？缺失时接受能力降级，还是投入 ACP 适配器扩展？不能由研究替用户选择。
2. 是否采用某些实验性/厂商扩展，如何协商、版本锁定与退出兼容？需要分别评估普通 ACP、标准草案子会话和 AIR 合同，不能共用一个“高级能力”开关。
3. F12.5 需要保留哪些任务输出，断连后怎样补齐？如需新的读取方法，应规定来源、范围、完整性和取消语义。
4. 原始 transcript 是必须取得的原生附件，还是另有明确范围的原文历史？若要求原始文件，经 ACP 暴露的导出合同尚需补证或设计。
5. 定版前应以真实 Codex/Claude 和两个目标 OS 验证创建/恢复、表单、配置确认、插话竞争、后台活动、子历史重放与引用 MCP；本轮静态结果不可替代这些验证。

这些问题输入后续“统一 ACP 接入与 Agent 能力承诺”“业务历史与跨 Agent 引用的权威来源”“会话控制、授权与消息投递的责任”“持续目标、子 Agent 与后台任务的职责”等决策票。未新增或关闭人工决策票。
