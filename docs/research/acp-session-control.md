# ACP 会话控制与投递证据的实际边界

研究日期：2026-11-04。对应架构研究票“ACP 会话控制与投递证据的实际边界”，补充 [ACP 必需能力覆盖研究](acp-required-capabilities.md)。本文提供控制证据，不决定队列、取消、信任或恢复的产品交互。

## 范围与证据等级

- **规范合同**：当日 ACP 官方文档；网站仍标示 v1 Latest、v2 Draft。v2 的 prompt 接受与完成语义不同，不可用于解释 v1 响应。[版本迁移][P-migration]
- **固定文档／源码观察**：Codex ACP 固定为 `927c4802f83c0859b439e7d641b5eb692560e5df`，Claude Agent ACP 固定为 `eb78afa8556b8fc4082495e48e3fd1762fb6a50f`；所有代码链接均指向 `agentclientprotocol` 下的这两个提交，不采用旧 `zed-industries` 仓库或 HEAD。[Codex 提交][C-commit]、[Claude 提交][L-commit]
- **推论**：从已引用合同／分支得到的设计约束，不冒充协议保证。**缺口**：本轮没有足够证据证明；不等于已经证实上游故障。
- 仅阅读公开规范与固定源码，没有安装或执行 Agent，没有登录厂商账户，没有读取用户会话库，没有执行运行时测试。既有研究中的“存在能力路径”不自动提升为已验证组合。

## 1. prompt 响应、接受、消费与重试

**规范合同：JSON-RPC 请求 ID 负责关联请求和响应。** 规范要求响应回显 ID；notification 没有响应。它没有定义同 ID 重试去重、跨连接结果查询或 exactly-once 执行。因此“已分配 ID”“写入传输成功”“重新连接后仍有 sessionId”均不能单独证明输入被接受或可以安全重发。这是对合同范围的判断，不是网络实测。[JSON-RPC 请求与响应][P-jsonrpc]

| 所见证据 | 可以据此说什么 | 不能据此说什么 |
| --- | --- | --- |
| v1 `session/prompt` 返回 `stopReason` | 对应 prompt 的回合按协议已结束；这是最终回合响应，不是独立的早期接受回执。[v1 prompt][P-prompt-v1] | 不能统一称为“任务成功”，也不能反推所有输入都曾到达模型；取消可能发生在启动前。[C prompt][C-prompt]、[L 启动前取消][L-turn-queue] |
| v1 JSON-RPC error | 这次方法调用失败。[JSON-RPC][P-jsonrpc] | 不保证失败发生在输入消费／工具副作用之前；Codex 会在原生 turn 完成并检查错误后抛错，Claude 会在已排队回合的消费循环失败时拒绝请求。[C 回合后错误][C-prompt-result]、[L 失败结算][L-failure] |
| v1 `user_message_chunk`、工具或文本更新 | Agent 公开了对应内容／活动；v1 消息 ID 可选。[v1 prompt][P-prompt-v1] | 内容相同不证明它属于某次重试；消息 ID 与 RPC 请求 ID 不是同一合同。[JSON-RPC][P-jsonrpc] |
| v2 成功响应 `{messageId}` | 当日草案定义为已插入 ACP 对话；不等于持久保存或模型消费。回合完成转到 `state_update`。[v2 接受边界][P-prompt-v2] | 响应丢失仍可能不确定，历史可缺失临时／已丢弃消息，重试可能新增一次提交；草案明确指出这一点。[v2 接受边界][P-prompt-v2] |

**固定源码观察：没有发现可供宿主使用的 prompt 幂等键合同。** Claude 每次 `startTurn` 生成新 `promptUuid`，随后把新消息推入 SDK；Codex 每次 steering 生成新 `clientUserMessageId`。这些内部 ID 有路由用途，不能把它们解释成“宿主重发相同 RPC ID 会复用同次输入”。[L UUID 与投递][L-prompt]、[L 入队][L-turn-queue]、[C steering][C-steer-code]

**stopReason 与错误的含义。** v1 的 `end_turn`、`max_tokens`、`max_turn_requests`、`refusal`、`cancelled` 描述结束原因；`end_turn` 不是业务结果验收。尤其在协商 AIR typed failure 的路径中，两家固定源码均可能返回 `end_turn` 并把终止失败放在 `_meta`：Codex 的 `terminalFailurePromptResponse`，以及 Claude 的 `failActiveWithSessionFailure`；Claude 未协商时保留 JSON-RPC error 路径。因此只判断“有 result”或“stopReason 是 end_turn”会漏掉实现公开的失败。[v1 原因表][P-prompt-v1]、[C 终止失败][C-terminal-failure]、[L 终止失败][L-failure]

**推论：投递证据必须保留歧义。** “尚未交给传输”“可能已送达”“明确被拒绝且未插入／未消费”“已获得特定阶段响应”“对应回合结束”是不同事实。一般 error、超时、断连、空历史都不等于未消费证明；不能把取消或恢复当作已执行副作用的回滚。v2 的接受合同更明确，但仍不提供丢失响应后的自动重发保证。[JSON-RPC][P-jsonrpc]、[v2 接受边界][P-prompt-v2]、[C 回合后错误][C-prompt-result]

## 2. cancel、权限与“已经停止”的范围

**规范合同：`session/cancel` 是 notification，方法本身没有成功回执。** v1 要求客户端对待决 `session/request_permission` 回答 `cancelled`；Agent 中止该 prompt 的操作、发出待发送更新后，以原 `session/prompt` 的 `cancelled` 响应确认取消。客户端应继续接收取消之后、最终响应之前的更新。仅发送 cancel 或把工具卡先标为 cancelled，都不是 Agent 已停妥的证据。[v1 取消][P-prompt-v1]

**请求取消与会话取消不同。** `$/cancel_request` 针对单个 RPC 请求，支持是可选的，原请求仍需得到相应响应。它不自动等价于取消整段会话工作。[v1 请求取消][P-cancel-v1] 当日 v2 草案把前台取消完成放在 `idle + cancelled`，并要求取消相关待决权限；取消单个权限／追问不自动决定整个工作流的状态。[v2 取消][P-cancel-v2]

**追问不能只按 sessionId 整批推定已经撤销。** 当日 v2 草案明确：session cancel 不自动取消该会话的每个 elicitation，Agent 应分别取消因前台取消而失效的请求；已接受的外部 URL 工作流可能继续，取消不保证关掉浏览器或撤销外部副作用。[v2 elicitation 取消][P-elicitation-v2] 固定 Claude 的一般权限处理把 cancelled 作为工具使用中止，但保留具体例外：对仍运行的后台子 Agent，以及 v2 中没有 session cancel 的交互撤回，可以映射为单次工具拒绝并继续。因此不能把每次权限弹窗取消都解释成整个 prompt 已停止。[L 权限取消文档][L-permission-doc]、[L 后台权限例外][L-permission-cancel]

**allow_once／allow_always 的责任已有明确标准答案。** v1 客户端回传 `selected + optionId`；`kind` 是呈现提示。官方文档明确由 Agent 记忆 `allow_always`／`reject_always`，决定作用范围与存续时间，并通过不再询问应用该决定；选项名称说明范围。标准另行允许客户端依据用户设置自动允许／拒绝，但没有要求客户端把一次 always 选择存成自己的跨 Agent 自动审批规则。[v1 权限选项][P-permission]

**固定映射说明了为何不能只看 kind。** Codex 可把 `acceptForSession`、命令前缀策略修订都映射为 `allow_always`，再按返回的 optionId 找回具体原生决定。Claude 验证选项确实由本次请求提供，解释为 SDK PermissionResult／updatedPermissions；其规则可以写向 localSettings，也可以只改变 session mode。长期效果由 Agent／适配器和原生权限机制实现，没有统一的跨 Agent scope、持久化寿命或撤销合同。[C 选项映射][C-permission-options]、[C 选择解析][C-permission-handler]、[L 选择校验][L-permission-response]、[L 权限效果][L-permission-effects]

| 固定实现 | 观察到的取消路径 | 证据限制 |
| --- | --- | --- |
| Codex | 先标记 active prompt 并 abort 权限／elicitation，取消尚未插入的 v2 排队 prompt，再调用原生 `turn/interrupt`；普通 cancel 等待后续 `turn/completed` 使 prompt 完成。原生中断有注册竞争重试。[C cancel][C-cancel]、[C interrupt][C-interrupt] | `requestTurnInterrupt` 对最终失败记录日志并返回；“已调用该函数”不等于收到原生停妥事件。cancel 的 Promise 也不是 ACP notification 的 wire 回执。[C interrupt][C-interrupt]、[JSON-RPC][P-jsonrpc] |
| Claude | 标记取消边界与开放交互，取消已收到的队列项，尝试从原生队列删除消息，再 `query.interrupt()`；正常活动回合由消费循环结算。[L cancel][L-cancel] | 已送入 SDK 的排队消息可能已被其取走；代码专门保留 orphan 追踪，避免其稍后结果误归下一回合。这不是“这些输入肯定未运行”的证明。[L 队列取消][L-cancel-queue] |
| Claude 的兜底 | 默认 30 秒后 SDK 仍不产出时，唤醒消费循环并强制结算 active prompt 为 `cancelled`；消费循环可以继续等待底层迟到结果。[L 计时器][L-cancel-backstop]、[L 常量][L-cancel-constant]、[L 强制结算][L-cancel-consumer] | 代码日志明确底层 query 可能仍卡住；故固定实现的 `cancelled` 只能视为适配器回合结算，不能无条件升级成“所有原生工作已停止”。这是源码暴露的保证缺口，未做故障复现。[L 计时器][L-cancel-backstop] |

**前台 prompt 结束与后台／原生任务停止必须分开。** 两家固定 goal 文档都把目标生命期与单次 prompt 分开；Codex 有独立后台终端 terminate 路径。Claude 创建选项明确后台命令可跨 interrupt 存续，采用 subagent RFD 时还会保留后台子 Agent；后台 workflow 则有另外的 stopTask 补偿。因此不能把根会话 cancel 声称为通用“杀掉全部任务”。[C goal 生命周期][C-goal-doc]、[L goal 生命周期][L-goal-doc]、[C 后台任务][C-async-doc]、[L 后台存续选项][L-background-options]、[L workflow 停止][L-cancel]

当日 v2 草案进一步规定 cancel 不覆盖最近 `background` 列表中的工作，前台 idle 也可以伴随后台工作；这能作为未来版本的明确范围，但不能据此声称固定适配器的所有任务都已满足该草案。[v2 前后台与取消][P-prompt-v2]

这三个判断不能合并：

| 判断 | 所需证据及限制 |
| --- | --- |
| 当前 prompt 已结算 | v1 对应请求的最终响应；Claude 强制 cancelled 也属于这个层级。[P-prompt-v1]、[L-cancel-backstop] |
| 可以接收下一输入 | v1 允许完成回合后再发 prompt，但这不是进程健康证明；Claude 对 queryClosed 会拒绝新 prompt，Codex 可能需要先恢复原生会话。宿主队列是否继续仍是独立产品控制。[L-prompt]、[C-recovery] |
| 全部后台／原生活动已停止 | 需要对应任务／进程的独立停止证据；当前 prompt 结算与下一输入被允许都不能推出这一点。[C-async-doc]、[L-background-options] |

## 3. steering 与一次普通 prompt 回退的证据

**扩展文档合同。** Codex 的 `_session/steering` 是 pre-standard 扩展；`_meta.steering.supported` 不区分可选行为，文档要求结合适配器版本判断。`idleBehavior: "promptRequired"` 是请求级 opt-in。[C steering 文档][C-steer-doc]

| 结果 | 固定文档含义／源码实际确认点 | 对重发的含义 |
| --- | --- | --- |
| `injected` | C：原生 `steerTurn` 成功返回；L：给当前未结算 turn 记 UUID 并 `session.input.push`，之后返回。L 不等待该条消息的模型消费确认。[C 注入][C-steer-code]、[L 注入][L-steer] | 不应另发同内容的普通 prompt；也不能称为该输入已得到最终回答。 |
| `startedNewTurn` | C 文档说已启动新回合，代码一般等到启动回调，但命令型即时完成也可走成功分支。L 调用 detached `prompt()` 后直接返回，之后的失败只记日志。[C 新回合][C-steer-code]、[L 空闲 steering][L-steer] | 不能统一视为“已有可等待的普通 prompt 完成响应”，更不是失败后可安全重发的凭据。 |
| `promptRequired` | C 文档承诺未消费、未启动新回合；L 空闲分支在 `prompt()`／SDK push 前返回，检查与活动分支 push 位于同步段。[C 合同][C-steer-doc]、[L 空闲分支][L-steer] | 有可靠未消费证据时，才具备把同一输入转为一次普通 prompt 的前提；不等于解除原 prompt 或其待决审批。 |
| `failed`／JSON-RPC error | C 将非 `RequestError` 异常折为 `failed`；L 的 `SteerResponse` 只有前三种结果，handler 直接调用 `steer`，错误走异常路径。[C 异常][C-steer-code]、[L 结果类型][L-steer-type]、[L 路由][L-steer-route] | 一般失败不承诺输入未消费。错误码本身也不补充这一证明，除非能定位为投递前拒绝。 |

**Codex 的文档合同与实现缺口必须并列保留。** `injectSteerIntoActiveTurn` 捕获原生 steer 错误后先等待事件；仅在“原回合仍活动，而且错误不是 no active turn”时重新抛错。若原回合已结束，它对其他错误也返回 false，调用方随后可给出 `promptRequired`。这段代码没有要求返回 false 前得到明确的“原生拒绝且未接受”证据。[C 注入 catch 与回退][C-steer-code]

**推论／待验证反例：** 若原生已接受 steer，但回复在适配器与原生服务之间丢失，同时回合结束，则上述分支是否仍可能报 `promptRequired`，本轮静态阅读无法排除。不能把文档中的未消费承诺直接当作该固定组合对所有传输故障均已证明的性质。它与“在任何投递动作前即因无活动 turn 返回 promptRequired”不同，也与“明确收到 no active turn 拒绝”不同。[C 注入 catch][C-steer-code]

**Claude 也存在错误阶段差异。** 活动路径先 push，再 await `publishGoalFromPrompt`；因此后续发布异常可以发生在输入已经排入 SDK 之后。`injected` 只是插入适配器／SDK 输入路径的确认，不是模型已读回执；丢失该响应不能安全回退。[L 注入顺序][L-steer]

**回退的证据条件，而非产品决定：** 已协商且固定版本支持该 opt-in；对应尝试得到可靠未消费证据；此后没有其它路径投递同一输入；原 prompt／审批已按当前控制状态处理；再由一次普通 prompt 拥有最终结果。Codex 文档要求在原 prompt 仍待 plan approval 时，先处理其权限并等待原 prompt 响应。响应丢失、普通 `failed` 或上述有歧义的 `promptRequired` 都不能自动补成这些条件。[C host-owned fallback][C-steer-doc]、[C 注入 catch][C-steer-code]

## 4. 配置、模型与模式的确认

**规范合同。** v1 config options 可以在空闲或生成中修改；`session/set_config_option` 响应和 `config_option_update` 携带完整配置及当前值，模型变化可能带来其它选项变化。客户端不能只把请求中的值当成确认值，也没有通用的“所有选项在本次正在进行的模型调用中立即生效”保证。[v1 配置][P-config-v1]

| 固定实现 | 确认与生效观察 | 运行中限制／缺口 |
| --- | --- | --- |
| Codex model、reasoning、权限 mode、fast | 校验后修改 SessionState 并返回配置；不是逐项等待原生配置 RPC。正常 prompt 在准备启动时从 SessionState 取 model／mode／serviceTier，再传给原生 `runTurn`。[C config][C-config]、[C 回合配置读取][C-dispatch-config]、[C sendPrompt][C-send-prompt] | 检查的配置入口没有统一 busy 拒绝。已取得的回合参数不会因 SessionState 后续改值而自动重建；不能声称新模型／权限模式已改写当前原生回合。 |
| Codex collaboration mode | 活动 app-server 路径 await `threadSettingsUpdate` 后更新；失去原生进程的会话先保存值，之后 resume 应用。[C config][C-config]、[C 原生设置][C-native-setting] | 与 model／权限 mode 不是同一确认层级；具体原生运行中拒绝条件本轮未证明。 |
| Claude model／permission mode | 模型先 await `query.setModel`；模式先 await `query.setPermissionMode`，可能返回实际 fallback mode；再同步 config。[L config][L-config]、[L mode][L-mode] | 配置变更自身按会话串行；所查入口未统一拒绝 active turn，也未把配置队列和普通 prompt 变成同一个原子事务。[L 配置串行][L-config-queue]、[L prompt][L-prompt] |
| Claude effort／fast／agent selector | 先 await `applyFlagSettings` 再改已展示值；agent selector 的注释明确是下一回合生效。Fast 还会随 SDK 状态／禁用原因再次同步。[L flag 配置][L-flags] | 不能把所有 selector 的成功响应概括为当前模型请求已切换；SDK 在每种运行状态的完整接受／拒绝矩阵未实测。 |

**推论：** 宿主可以分别记录请求值、最后确认配置、配置失败和实际投递时的确认快照；“排队输入采用哪次设置”仍需宿主自己的时序合同。以上标准与实现没有替宿主提供跨配置 RPC、prompt RPC 和原生队列的全局原子性。[P-config-v1]、[C-dispatch-config]、[L-config-queue]、[L-turn-queue]

## 5. load／resume／重连能否只算无害重建

**规范合同。** v1 load 包含历史重放，resume 不重放；load 响应标记该重放过程完成，随后会话可继续接收 prompt。该合同没有把原生 load/resume 定义成“仅查看历史、不得启动其它活动”。[v1 setup][P-setup-v1]

- **Codex 固定实现有直接证据表明不能假定恢复静默。** ACP load 与 resume 都进入原生 `thread/resume` 路径；`startCodexTurnTracker` 特意在请求之前订阅，注释说明要捕获恢复后数毫秒内由 goal 自动启动的回合；load 只延后向客户端放行这些 live 事件，直到历史重放完成。门控的是事件呈现次序，不是禁止原生执行。[C 原生恢复][C-native-resume]、[C 基线 tracker][C-resume-tracker]、[C load][C-load]
- **Codex 的内部重连也不是纯网络重连。** 固定 `AppServerRecovery` 会在后续需要时重启 app-server，并在使用失效会话时重新 resume。它描述的是适配器内部恢复机制，不是 Aibo 应用重启后的用户授权策略。[C recovery][C-recovery]
- **Claude 固定实现不等同于读文件。** 同设置且仍在内存的会话可直接复用；设置指纹变化会 teardown 再创建；不存在的本地实例通过 `query({prompt: input, options: {resume: ...}})` 创建，源码明确 query 立即启动 CLI。load 可以与 replay 并行创建该会话，返回后还安排命令发现、子会话状态及 MCP 认证工作。[L getOrCreate][L-get-or-create]、[L 创建与 CLI][L-create]、[L load][L-load]、[L setup 后工作][L-after-setup]
- **Claude 的自主活动有文档，恢复即自主推进的普遍保证仍未核实。** 文档确认后台任务完成后可以无前台 prompt 地开启自主回合；代码的长寿命 consumer 处理这些活动，但一般由首次 prompt 延迟启动。不能从这一点断言“每次 load 都会继续模型工作”，也不能反推“load 不可能触发任何原生执行”。[L 自主回合][L-async-doc]、[L consumer][L-consumer]

**推论：** 重新读取宿主自己保存的 ACP 可观察历史／共通呈现状态，与启动 Agent、建立原生 session、调用 load/resume 是不同动作。若产品要求重启后不自动推进，单靠“没有再发送普通 prompt”不足以证明满足；需要对会话打开、原生恢复和持续目标分别设定允许条件。本文只确认这一事实依赖，不代替产品决策。[C-resume-tracker]、[L-create]

**缺口：** 本轮没有证明跨进程崩溃后全部原生目标、后台任务、未完成权限与事件都可恢复；也没有证明原生恢复具有“无副作用预览”模式。静态代码与历史回放不能替代这些性质的组合验收。[C-recovery]、[C-load]、[L-load]

## 6. 客户端 fs／terminal 与 Agent 自执行

**规范合同：ACP 不是进程沙箱。** 官方安全说明明确 Agent 以其进程权限运行，自行执行的操作由 Agent 负责遵守 permission 决定；客户端负责自己实际执行的 fs／terminal 请求。Agent 可以不提请权限便行动，session／message／tool ID 也不是凭证。[ACP 信任模型][P-security]

**v1 fs 与 terminal 是可选客户端设施。** 未声明对应能力，Agent 不得调用那些客户端方法；这不禁止 Agent 用自己的本地工具执行。fs 特别用于客户端环境、未保存编辑器内容等；terminal 则是在客户端环境创建与管理命令。v2 草案移除这些客户端执行 API，客户端工具可经 MCP 提供；v2 display-only terminal 是输出呈现，不是宿主代执行。[v1 fs][P-fs]、[v1 terminal][P-terminal]、[v2 迁移][P-migration]

**固定实现与此一致。** Codex 给原生 runTurn 传 `approvalPolicy`、`approvalsReviewer`、`sandboxPolicy`；这些来自原生模式预设，full access 会使用 `never` 与 `dangerFullAccess`。Claude 默认暴露原生 `claude_code` tools，给 SDK `canUseTool` 回调；源码强调 bypass 已由 Claude Code 在回调前处理，仍抵达回调的请求不能因 UI mode 再被宿主自动放行。[C runTurn][C-send-prompt]、[C 模式预设][C-mode]、[L 原生工具][L-native-tools]、[L SDK 配置][L-sdk-options]、[L 权限回调][L-permission]

**推论：** 宿主可以可靠控制自己执行的工具和自己授予的引用读取授权，但不能因接入 ACP 或显示了审批 UI，就承诺“所有文件、网络和命令操作均经过宿主”。未受信工作区启动原生 Agent 前是否取得信任，是执行边界；延迟第一条用户 prompt 本身不构成沙箱。原生启动、恢复、MCP 启动和运行模式都需要纳入已验证组合的范围。[P-security]、[C-send-prompt]、[L-create]、[L-sdk-options]

### 会话注入 MCP 的身份边界

**已核实范围：** ACP session setup 可传 MCP 的 command／args／env 或 HTTP url／headers；固定 C/L 映射把这些配置交给各自原生 MCP 客户端。所查映射没有为每次 tool call 自动添加一个已认证的 Aibo Session 身份，也不能把 ACP `sessionId` 当凭证。[ACP MCP 接线][P-setup-v1]、[C MCP 映射][C-mcp]、[L MCP 映射][L-mcp]、[ACP 凭证边界][P-security]

**推论／缺口：** 宿主可考虑在自己控制的 MCP 连接配置中绑定独立的会话凭证，再由服务端解析其引用读取授权；这属于宿主自己的合同，不是 MCP 已替宿主验证调用者的保证。相同服务器配置复用、子 Agent 继承和重连后的实际凭证隔离，本轮未证明。不能仅接收模型给出的 sessionId 参数就授权；本票不扩展到通用 MCP 身份认证设计。[C-mcp]、[L-mcp]、[P-security]

## 需要保留的证据缺口

1. Codex steering 原生响应丢失并与回合结束竞争时，`promptRequired` 的未消费承诺是否成立；当前固定代码不足以排除歧义。[C-steer-code]
2. Claude 强制 cancelled 路径以及两家后台任务取消，是否能给出独立的原生任务终止证据；prompt 结算不能代替这一证明。[L-cancel-backstop]、[L-background-options]、[C-async-doc]
3. 各配置在运行中、待权限、SDK 队列已有输入时的实际应用边界；目前只有上述具体代码路径，没有统一原子确认合同。[C-config]、[C-dispatch-config]、[L-config]、[L-config-queue]
4. 原生 load/resume 的自主工作、后台存续及会话 MCP 凭证隔离的可验证条件；不能用纯历史重建或一次正常 prompt 验收替代。[C-resume-tracker]、[L-create]、[C-mcp]、[L-mcp]

这些缺口可由后续协议夹具、可控传输故障与已验证组合验收补证；本轮没有运行上述验证，也没有关闭人工决策票。

[P-jsonrpc]: https://www.jsonrpc.org/specification#request_object
[P-migration]: https://agentclientprotocol.com/protocol/v2/migration
[P-prompt-v1]: https://agentclientprotocol.com/protocol/v1/prompt-turn
[P-prompt-v2]: https://agentclientprotocol.com/protocol/v2/prompt-lifecycle
[P-cancel-v1]: https://agentclientprotocol.com/protocol/v1/cancellation
[P-cancel-v2]: https://agentclientprotocol.com/protocol/v2/cancellation
[P-elicitation-v2]: https://agentclientprotocol.com/protocol/v2/elicitation#session-cancellation
[P-config-v1]: https://agentclientprotocol.com/protocol/v1/session-config-options
[P-setup-v1]: https://agentclientprotocol.com/protocol/v1/session-setup
[P-security]: https://agentclientprotocol.com/protocol/v1/security
[P-fs]: https://agentclientprotocol.com/protocol/v1/file-system
[P-terminal]: https://agentclientprotocol.com/protocol/v1/terminals
[P-permission]: https://agentclientprotocol.com/protocol/v1/tool-calls#permission-options
[C-commit]: https://github.com/agentclientprotocol/codex-acp/commit/927c4802f83c0859b439e7d641b5eb692560e5df
[C-prompt]: https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/src/CodexAcpServer.ts#L4759-L4840
[C-prompt-result]: https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/src/CodexAcpServer.ts#L5140-L5190
[C-terminal-failure]: https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/src/CodexAcpServer.ts#L5558-L5577
[C-cancel]: https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/src/CodexAcpServer.ts#L5648-L5680
[C-interrupt]: https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/src/CodexAcpServer.ts#L4398-L4530
[C-steer-doc]: https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/docs/steering.md
[C-steer-code]: https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/src/CodexAcpServer.ts#L2953-L3185
[C-config]: https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/src/CodexAcpServer.ts#L2735-L2900
[C-dispatch-config]: https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/src/CodexAcpServer.ts#L5084-L5120
[C-send-prompt]: https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/src/CodexAcpClient.ts#L1360-L1391
[C-native-setting]: https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/src/CodexAcpClient.ts#L1393-L1398
[C-load]: https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/src/CodexAcpServer.ts#L1844-L1955
[C-native-resume]: https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/src/CodexAcpClient.ts#L642-L790
[C-resume-tracker]: https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/src/CodexAcpServer.ts#L1614-L1668
[C-recovery]: https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/src/app-server-recovery/AppServerRecovery.ts#L100-L206
[C-goal-doc]: https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/docs/air-extensions.md#lifecycle
[C-async-doc]: https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/docs/air-extensions.md#async-tasks
[C-mode]: https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/src/AgentMode.ts#L39-L106
[C-mcp]: https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/src/McpServerConfig.ts
[C-permission-options]: https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/src/permissions/options.ts#L20-L77
[C-permission-handler]: https://github.com/agentclientprotocol/codex-acp/blob/927c4802f83c0859b439e7d641b5eb692560e5df/src/permissions/CodexApprovalHandler.ts#L135-L158
[L-commit]: https://github.com/agentclientprotocol/claude-agent-acp/commit/eb78afa8556b8fc4082495e48e3fd1762fb6a50f
[L-prompt]: https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/acp-agent.ts#L3897-L4040
[L-turn-queue]: https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/acp-agent.ts#L4070-L4123
[L-failure]: https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/acp-agent.ts#L5805-L5900
[L-cancel]: https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/acp-agent.ts#L8831-L8955
[L-cancel-queue]: https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/acp-agent.ts#L8955-L9065
[L-cancel-backstop]: https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/acp-agent.ts#L9150-L9235
[L-cancel-constant]: https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/acp-agent.ts#L451-L460
[L-cancel-consumer]: https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/acp-agent.ts#L5930-L6006
[L-background-options]: https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/acp-agent.ts#L12083-L12095
[L-permission-doc]: https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/docs/air-extensions.md#cancellation-and-failures
[L-permission-cancel]: https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/acp-agent.ts#L10887-L10918
[L-permission-response]: https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/permissions/response.ts
[L-permission-effects]: https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/permissions/effects.ts#L34-L221
[L-steer]: https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/acp-agent.ts#L4503-L4608
[L-steer-type]: https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/acp-agent.ts#L729-L778
[L-steer-route]: https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/acp-agent.ts#L14702-L14708
[L-config]: https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/acp-agent.ts#L9390-L9520
[L-mode]: https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/session-mode.ts#L246-L280
[L-config-queue]: https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/acp-agent.ts#L12811-L12823
[L-flags]: https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/acp-agent.ts#L11510-L11655
[L-get-or-create]: https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/acp-agent.ts#L11701-L11755
[L-create]: https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/acp-agent.ts#L12276-L12328
[L-load]: https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/acp-agent.ts#L3165-L3288
[L-after-setup]: https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/acp-agent.ts#L11191-L11214
[L-consumer]: https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/acp-agent.ts#L4636-L4660
[L-async-doc]: https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/docs/air-extensions.md#autonomous-turn
[L-goal-doc]: https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/docs/air-extensions.md#lifecycle
[L-native-tools]: https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/acp-agent.ts#L11975-L12004
[L-sdk-options]: https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/acp-agent.ts#L12110-L12147
[L-permission]: https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/acp-agent.ts#L10705-L10828
[L-mcp]: https://github.com/agentclientprotocol/claude-agent-acp/blob/eb78afa8556b8fc4082495e48e3fd1762fb6a50f/src/acp-agent.ts#L11888-L11914
