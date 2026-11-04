# ACP 持续目标控制与子任务补读的实际边界

研究日期：2026-11-04。补充 [必需能力覆盖](acp-required-capabilities.md)与[会话控制证据](acp-session-control.md)，只核查目标动作、子 Agent 过程和独立后台任务输出，不重述基础 ACP 合同。

## 范围与证据等级

- **固定源码观察**：Codex ACP `927c4802f83c0859b439e7d641b5eb692560e5df`；Claude Agent ACP `eb78afa8556b8fc4082495e48e3fd1762fb6a50f`。goal／asyncTasks 仍属实验性 AIR 合同；子会话标准草案路由与 AIR 旧路由须分别协商，不可把不同握手的效果拼成一个保证。[C goal 文档][C-goal-doc]、[L goal 文档][L-goal-doc]、[C 子路由][C-sub-protocol]、[L 子路由][L-sub-protocol]
- **对应原生源码**：Codex 锁文件解析为 `@openai/codex 0.165.0`，官方 `rust-v0.165.0` 标签指向 `360319d8dee62d87e2d59fc0c21805d79519cfed`。下文原生结论仅适用于这一对应依赖样本，**不证明用户实际安装或适配器实际启动的二进制就是该版本**。[锁文件][C-lock]、[标签对象][N-tag]
- Claude 依赖固定为 SDK `0.3.312`，该 SDK 标签的 changelog 写明与 Claude Code `2.1.312` 对齐；公开标签树没有原生 `/goal`／Stop hook 实现。本轮未取得可逐行核查的对应原生实现，不能用当前其他版本文档补成保证。[依赖][L-package]、[SDK changelog][L-sdk-change]、[SDK 标签树][L-sdk-tree]
- **推论**表示从上述调用链得到的限制；**缺口**表示尚未证明。全程仅静态阅读，没有启动 Agent、安装依赖、访问真实会话或运行验收测试。官方 OpenAI goal 指南只作背景，不取代固定版本证据。[官方指南][N-guide]

## 1. goal 动作实际改变什么

| 动作 | 固定适配器调用 | 对应原生效果与不能承诺的范围 |
| --- | --- | --- |
| Codex set | `_session/goal {action:"set",objective}` → `runGoalSet` → `thread/goal/set {objective,status:"active"}`。[C handler][C-goal-handler]、[C client][C-goal-client] | 写目标；原生将目标纳入当前回合记账，已有目标的 objective 改变时可注入当前回合；调用 `continue_if_idle` 尝试启动自主回合。受功能、原生运行门控及提交结果限制，active 不证明已经开始执行。[N set][N-set]、[N runtime][N-runtime]、[N continuation][N-continue] |
| Codex resume | 同方法 `action:"resume"` → `thread/goal/set {status:"active"}`，不重发 objective。[C-goal-client] | 同一 active 路径可自行启动后续工作；无需宿主补发“继续”。没有现存 goal 时原生拒绝状态更新。[N mutation][N-mutation]、[N-runtime] |
| Codex pause | `action:"pause"` → `thread/goal/set {status:"paused"}`。[C-goal-handler]、[C-goal-client] | 原生更新持久状态并清除 active-goal 记账；调用链没有中断当前回合。已经开始的前台或自主回合可继续；后续 goal continuation 读取到非 active 状态便返回。没有额外停止后台任务。[N-mutation]、[N-runtime]、[N accounting][N-accounting]、[N-continue] |
| Codex clear | `action:"clear"` → `thread/goal/clear`。[C-goal-client] | 删除 goal，并清理待用回合选项和 active-goal 记账；没有中断当前回合或终止后台任务。之后无 goal 时不再启动目标续跑；这不撤销已发生的工具副作用。[N clear][N-clear]、[N-runtime]、[N-continue] |
| Claude set／clear | 生成 `/goal <objective>` 或 `/goal clear`，先 steer（`idleBehavior:"promptRequired"`）；无未结算回合时再 `await prompt()`。[L handler][L-goal-handler] | 文档说明 set 安装会话 Stop hook，clear 移除目标。两者都是输入动作，活动时通常以 `priority:"now"` 投递；待权限／追问时改为 `later`。适配器注释说明 now 可中断当前生成 cycle，或把前台工具转入后台；**不能承诺只改未来状态，也不是 session/cancel 的停妥保证**。[L mapping][L-goal-doc]、[L steering][L-steer] |
| Claude pause／resume | 未声明、解析器拒绝；不是上述 set／clear 的别名。[L actions][L-goal-type] | 原生暂停能力在此 ACP 样本中不可用；不能以清除目标、重发 objective 或暂停宿主队列冒充原生暂停／恢复。 |

**Codex 的成功和快照。** 原生 set 先返回目标并发出有序 `thread/goal/updated`，之后才应用运行效果；运行效果失败可只记日志。适配器的非 active 路径等待匹配的目标更新；clear 若原生报告确曾删除，则等 cleared 通知，若原本无目标则直接返回。最终 ACP 返回 `{}`，pause／clear 另发目标快照或 null；这是目标状态证据，不是当前回合或后台停止证据。[N-set]、[N runtime 错误][N-effects]、[C 等待逻辑][C-goal-wait]、[C-goal-handler]

set／resume 的 `runGoalSet` 还可能等到观察到的首个目标回合结束，或在未观察到新回合时返回；handler 丢弃该完成结果。因此 `{}` 既不是统一的“仅接受”时点，也不是“整个持续目标成功完成”。恢复失效原生会话发生在 goal handler 执行动作之前，仍须保留既有研究中的恢复可启动目标回合这一边界。[C-goal-wait]、[C-goal-handler]、[C 恢复 tracker][C-resume-tracker]

**Claude 的成功和快照更弱。** 活动路径在输入 push 后就可返回 `injected`；空闲路径虽等待 prompt，却没有检查返回的 `PromptResponse` 就返回 `{}`。提交 goal 命令时适配器先发布乐观 active／null，匹配的 `active_goal` 才提供运行时快照；命令失败可回滚到旧目标。只见 clear 的 null，不能断言原生已处理清除，更不能断言所有工作停止。[L-goal-handler]、[L-steer]、[L 乐观状态][L-goal-state]、[L 回滚][L-goal-rollback]

**Claude 的原生缺口与后台边界。** 对应版本的 Stop hook 清理与当前生成之间的精确边界尚未证明。goal handler 没有另外调用 stopTask／interrupt；steer 自身的影响已在上表单列。后台任务完成也能独立触发无前台 prompt 的自主回合，因此清除目标不能概括为“全部自主活动停止”。[L-goal-handler]、[L-goal-doc]、[L 后台自主回合][L-autonomous]

**对独立目标控制的推论。** 仅调用目标入口时，Codex 当前工作可继续，Claude 当前工作可能因插话而变化，均需按实际证据显示。回合 cancel 与目标 pause 是可区分的操作及结果；Claude 的 pause 不支持，clear 不等价于暂停。若组合调用，也不能声称原子成功或全部后台停止。[C-goal-handler]、[L-goal-type]、[会话取消边界](acp-session-control.md#2-cancel权限与已经停止的范围)

**预算／用量展示。** Codex 快照实际映射 `tokenBudget/tokensUsed/timeUsedSeconds`，Claude 非空快照只填 objective、active、iterations、lastReason、createdAt；其可选预算字段没有被填充。可展示已取得值，不把字段存在当成持续刷新或两家均有数据的承诺；本票不增加预算设置或宿主硬限额要求。[C 快照][C-goal-snapshot]、[L-goal-type]

## 2. 已推送内容与主动补读

| 对象 | 已协商路径主动推送什么 | 身份关联与实际读取路径 |
| --- | --- | --- |
| Codex 子 Agent | 原生 spawn／activity 发现子 thread 后注册通知；标准路由先在父会话发 `subagent_update`，再把子消息、思考、工具等路由至子 session；不以 UI 展开为条件。[C 订阅][C-sub-subscribe]、[C 路由][C-sub-router] | 标准路由保存 root／parent／native thread 的关联，ACP 子 ID 为编码后的 `root:subagent:thread`；关系来自原生 sender／receiver／agentPath，标题只是显示名。父 `session/load` 会主动递归读子历史并推送，不是独立的子历史 read RPC。[C 关联][C-sub-identity]、[C 子回放][C-sub-replay] |
| Claude 子 Agent | 协商子会话会启用子文本／思考转发；原生事件、`parent_tool_use_id` 与 launch tool-use ID 决定路由，RFD 还发父子 `session_message`，消息用 sender／recipient message ID 对应。长寿命 consumer 在 prompt 外继续接收。[L 转发][L-sub-forward]、[L 子会话说明][L-sub-readme]、[L consumer][L-consumer] | RFD 子 session ID 是 SDK agentId；AIR 旧路由可能带 generation 后缀，不能混用。父 load 根据 `agent-<id>.meta.json` 的 toolUseId 与包含 launch 的 transcript 确认谱系，再用 `getSubagentMessages(rootSessionId,agentId,{dir})` 读取独立子 transcript 并回放。[L 身份][L-sub-identity]、[L 元数据][L-sub-history]、[L reader][L-sub-reader]、[L 子回放][L-sub-replay] |
| Codex 后台命令 | `async_task_spawned/state_update` 是任务状态；输出仍由原 command 工具／终端流承载，任务面板不再复制全文。`showInTranscript:false` 是显示策略，不表示没有业务内容可保存。[C 任务文档][C-async-doc]、[C 输出][C-command] | 根任务 ID＝command item ID；子任务 ID 加原生 child thread 前缀，toolCallId 仍是 item ID，事件归子 session。内部 `backgroundTerminals/list` 只有 itemId/processId/command，补的是存活状态，不补输出字节。[C-async-doc]、[C 后台 API][C-background-api] |
| Claude 后台任务 | SDK 结构化事件产生 spawn／progress／state；可带 description、summary、usage、toolCallId、outputFilePath。原工具结果可能另带部分内容，路径本身不是文件全文。[L 任务文档][L-async-doc]、[L 更新类型][L-task-types] | taskId 来自结构化任务字段；有 tool-use ID 才关联工具，缺失时等待或降级发布，不从标题／命令文本猜身份。SDK 未给 output_file 时，只从该任务自己的工具结果提取匹配任务 ID 的绝对路径；仍未读取文件。[L 输出路径][L-output-path] |

**接收与补读是不同动作。** 上表已实际推送的消息、工具结果和任务更新可持续保存，即使详情面板未展开、前台 prompt 已结束。额外补读才涉及再次请求原生历史／资源；不应因为有路径、任务标题或 UI 展开状态，就声称已保存全文或获得读取权限。[C-goal-doc]、[L-consumer]、[L-output-path]

| 补读候选 | 固定样本可实行的路径 | 断线后能补什么，是否可能启动执行 |
| --- | --- | --- |
| 只看宿主已保存内容 | 读取 Aibo 的 ACP 可观察历史，不调用 Agent | 没有原生执行入口；无法凭空补回未捕获内容。这是宿主动作与原生 load 的职责区分。 |
| Codex 子历史 | 内部 `readSubagentHistory` → `thread/read`，legacy 再 includeTurns；分页存储用 `thread/items/list` 并取回合结尾。[C 子读取][C-sub-read]、[C 分页][C-history-pages] | 原生 reader 读取存储／已有 live snapshot，没有 start/resume 回合调用，但**尚未作为单独 ACP read 方法公开**。现有父 load 会 resume 根会话；子回放还对已加载、父子关系匹配的 child 调 `thread/resume` 恢复订阅，故整个 load 不是只读查看。[N reader][N-read]、[C 子恢复][C-sub-restore]、[C-sub-replay]、[C-resume-tracker] |
| Claude 子历史 | 内部 `subagentHistory` 与 `readSubagentMessages` 读元数据、transcript／SDK reader；父 load 自动调用。[L-sub-history]、[L-sub-reader]、[L-sub-replay] | 内部读取函数未调用 query／prompt；但对外 load 可同时 `query({options:{resume}})` 启动 CLI，不能把这个组合称为无执行补读。仍没有独立子历史只读 ACP 方法。[L load][L-load]、[L query][L-query]、[L ACP 入口][L-methods] |
| 后台输出全文 | 当前 asyncTasks 公开的是更新和 stop；未找到按任务读取输出全文／缺口区间的 ACP RPC。[C 扩展入口][C-methods]、[L-methods] | Codex load 可重放已保存的已完成 command `aggregatedOutput`；inProgress 历史仅重建 start，没有已漏流式字节的补齐保证。Claude load 可恢复已持久化任务通知中的状态／路径，不因此读取 output 文件。不能承诺断线补齐全文。[C command history][C-command-history]、[L 任务回放][L-task-replay] |

**历史完整性限制。** Codex 子回放会跳过没有安全 delta cursor 的未完成文本，以及不能当人类输入重放的子 userMessage，并公开 history gap；丢失通知不能都从快照复原。Claude reader 失败会返回空列表，缺元数据的 launch 保留普通工具卡；适配器注明 SDK reader 只返回最后压缩后的 active chain，还主动去掉 fork 继承前缀。这些是可观察历史，不是逐字节原始 transcript 或完整捕获证明。[C-sub-replay]、[L-sub-reader]、[L 谱系读取][L-sub-lineage]、[L-sub-readme]

**对按需补读的推论。** 可以按需请求已明确允许的读取能力，但固定样本的父 load 会自动回放可取得的子历史，并非“仅选中这一子任务”的分页读取合同。若打开详情必须不启动原生执行，现有 load 不能作为透明替代；需要另外核实或增加只读 ACP 扩展，否则明确缺失。内部 SDK／app-server reader 的存在不授权宿主绕过 ACP 直读。ACP v1 `fs/read_text_file` 是 Agent 请求读取客户端文件的反向接口，也不能拿来充当宿主向 Agent 拉取日志。[C-sub-replay]、[L-sub-replay]、[C-methods]、[L-methods]、[ACP fs][P-fs]

## 3. 定向停止与验收范围

| 可实施控制 | 固定调用与成功证据 | 不可据此承诺 |
| --- | --- | --- |
| 子 Agent cancel | 仅按已公布的子能力调用 `session/cancel(childSessionId)`。Codex 对目标子会话及有归属证据的后代调用原生 interrupt；Claude 用 `Query.stopTask`，由原生事件报告 idle／cancelled。[C 子取消][C-sub-cancel]、[L-sub-readme] | 不把子 ID 当成可任意 prompt／resume 的原生会话；不推断父会话停止、独立后台命令停止或副作用回滚。 |
| Codex 后台 stop | `_session/async_task/stop {sessionId,asyncTaskId}` → 原生 `backgroundTerminals/terminate(threadId,processId)`；terminated 为真后发 stopped 并返回 true。[C 后台 stop][C-task-stop] | true 是该任务终止请求获得原生确认的路径，不是全部进程树／所有任务停止，也没有全文读取含义。 |
| Claude 后台 stop | 普通任务 `query.stopTask(taskId)`；代表自主回合的特殊 task 调 `query.interrupt()`，有未结算 prompt 时拒绝。调用完成后适配器发 stopped；Monitor 不提供 stop，子 Agent 另走子会话。[L stop][L-task-stop]、[L 分类][L-task-class] | 不把 adapter stopped 扩大为所有原生活动停妥。后台列表修复丢失终止事件只证明列表层存活；迟到通知还能把 stopped 修正为 completed／failed。[L 存活修复][L-task-liveness] |

后续**实测范围**，本轮尚未执行：

1. 固定适配器与实际原生二进制版本、OS、握手组合；分别在空闲、前台生成、工具运行、待权限、目标自主回合中验证 set/pause/resume/clear 的事件顺序与最终原生状态。
2. 分别验证目标控制、回合 cancel，以及组合调用的成功／失败／不支持及响应丢失，不能用单一成功标记覆盖两项结果。
3. Claude 乐观 goal/null、命令失败回滚、now/later 插话和工具转后台；Codex mutation 与 continuation 竞争、set 响应后运行效果失败、clear 原本无目标。
4. 不展开子界面时持续保存 live 内容；模拟传输丢包／断线后比对 replay 的缺口、压缩、未完成 command、父子身份与任务归属；不得以回放完成冒充全文完整。
5. 对每条候选只读补读路径验证不会创建 query、resume thread 或开始模型／工具执行；分别验证路径提示缺全文、文件消失、读取失败，以及后台定向停止与后续迟到事件。

这些测试用于建立已验证组合，不能由静态源码或现有上游测试文件代替；本票不扩展数据格式或替代职责票中的产品决定。

[C-goal-doc]: https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/docs/air-extensions.md#L691-L758
[L-goal-doc]: https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/docs/air-extensions.md#L767-L844
[C-lock]: https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/package-lock.json#L933-L950
[N-tag]: https://api.github.com/repos/openai/codex/git/tags/e374836b2728d9b7aa981c2902649e640a2edaa8
[L-package]: https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/package.json#L67-L72
[L-sdk-change]: https://github.com/anthropics/claude-agent-sdk-typescript/blob/e2963809c50847f00143646418177031aa38dd9c/CHANGELOG.md#L3-L8
[L-sdk-tree]: https://github.com/anthropics/claude-agent-sdk-typescript/tree/e2963809c50847f00143646418177031aa38dd9c
[N-guide]: https://developers.openai.com/cookbook/examples/codex/using_goals_in_codex
[C-goal-handler]: https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/src/CodexAcpServer.ts#L1004-L1044
[C-goal-client]: https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/src/CodexAcpClient.ts#L1076-L1126
[C-goal-wait]: https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/src/CodexAppServerClient.ts#L424-L605
[N-set]: https://github.com/openai/codex/blob/360319d8dee62d87e2d59fc0c21805d79519cfed/codex-rs/app-server/src/request_processors/thread_goal_processor.rs#L133-L251
[N-mutation]: https://github.com/openai/codex/blob/360319d8dee62d87e2d59fc0c21805d79519cfed/codex-rs/ext/goal/src/api.rs#L147-L294
[N-runtime]: https://github.com/openai/codex/blob/360319d8dee62d87e2d59fc0c21805d79519cfed/codex-rs/ext/goal/src/runtime.rs#L191-L269
[N-accounting]: https://github.com/openai/codex/blob/360319d8dee62d87e2d59fc0c21805d79519cfed/codex-rs/ext/goal/src/accounting.rs#L389-L402
[N-clear]: https://github.com/openai/codex/blob/360319d8dee62d87e2d59fc0c21805d79519cfed/codex-rs/ext/goal/src/api.rs#L296-L344
[N-continue]: https://github.com/openai/codex/blob/360319d8dee62d87e2d59fc0c21805d79519cfed/codex-rs/ext/goal/src/runtime.rs#L436-L517
[N-effects]: https://github.com/openai/codex/blob/360319d8dee62d87e2d59fc0c21805d79519cfed/codex-rs/ext/goal/src/api.rs#L76-L84
[L-goal-handler]: https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/acp-agent.ts#L4403-L4415
[L-steer]: https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/acp-agent.ts#L4503-L4608
[L-goal-type]: https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/goal-extension.ts#L6-L83
[L-goal-state]: https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/acp-agent.ts#L4467-L4501
[L-goal-rollback]: https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/acp-agent.ts#L7192-L7208
[C-goal-snapshot]: https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/src/ThreadGoalSnapshot.ts#L8-L35
[C-resume-tracker]: https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/src/CodexAcpServer.ts#L1614-L1668
[C-sub-protocol]: https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/src/subagents/AcpSubagents.ts#L46-L59
[L-sub-protocol]: https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/acp-subagents.ts#L15-L32
[C-sub-subscribe]: https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/src/subagents/CodexSubagentSubscriptions.ts#L39-L54
[C-sub-router]: https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/src/subagents/StandardSubagentEventRouter.ts#L125-L181
[C-sub-identity]: https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/src/subagents/SubagentAssociations.ts#L3-L39
[C-sub-replay]: https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/src/CodexAcpServer.ts#L3801-L3893
[C-sub-read]: https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/src/CodexAcpClient.ts#L843-L856
[C-history-pages]: https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/src/CodexAcpClient.ts#L1847-L1870
[C-sub-restore]: https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/src/subagents/CodexSubagentSubscriptions.ts#L105-L121
[N-read]: https://github.com/openai/codex/blob/360319d8dee62d87e2d59fc0c21805d79519cfed/codex-rs/app-server/src/request_processors/thread_processor.rs#L3022-L3140
[L-sub-forward]: https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/acp-agent.ts#L11804-L11810
[L-sub-readme]: https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/README.md#L39-L101
[L-sub-identity]: https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/native-subagents.ts#L1630-L1662
[L-sub-history]: https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/subagent-history.ts#L45-L112
[L-sub-reader]: https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/subagent-history.ts#L285-L366
[L-sub-lineage]: https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/subagent-history.ts#L147-L184
[L-sub-replay]: https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/acp-agent.ts#L9969-L10042
[L-consumer]: https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/acp-agent.ts#L4636-L4660
[L-load]: https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/acp-agent.ts#L3160-L3213
[L-query]: https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/acp-agent.ts#L12276-L12328
[C-async-doc]: https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/docs/air-extensions.md#L778-L815
[C-command]: https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/src/tool-calls/reporters/CommandReporter.ts#L21-L102
[C-command-history]: https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/src/tool-calls/reporters/CommandReporter.ts#L152-L169
[C-background-api]: https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/src/async-tasks/BackgroundTerminalApi.ts#L1-L30
[L-async-doc]: https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/docs/air-extensions.md#L955-L978
[L-task-types]: https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/acp-subagents.ts#L73-L115
[L-output-path]: https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/docs/air-extensions.md#L980-L1010
[L-task-replay]: https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/acp-agent.ts#L9724-L9771
[C-methods]: https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/src/AcpExtensions.ts#L81-L99
[L-methods]: https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/acp-agent.ts#L14673-L14719
[P-fs]: https://agentclientprotocol.com/protocol/v1/file-system#reading-files
[C-sub-cancel]: https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/src/subagents/StandardSubagentEventRouter.ts#L560-L603
[C-task-stop]: https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/src/async-tasks/CodexBackgroundTerminalTasks.ts#L186-L211
[L-task-stop]: https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/acp-agent.ts#L4611-L4634
[L-task-class]: https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/async-tasks.ts#L411-L445
[L-task-liveness]: https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/async-tasks.ts#L341-L394
[L-autonomous]: https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/docs/air-extensions.md#L1012-L1024
