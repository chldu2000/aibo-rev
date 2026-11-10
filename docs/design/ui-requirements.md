# Aibo Rev 默认工作台：界面硬性要求

Status: active（2026-11-10）

本文是默认呈现（F20.1）的**可验收约束**。设计背景、取舍理由与示意见 [界面与交互方案](ui-interaction.md)；原型只作示意。

## 0. 使用方式

**优先级。** 本文 > [界面与交互方案](ui-interaction.md) > [设计令牌](tokens/tokens.json)以外的任何数值 > [静态原型](prototype/index.html) 与 [`prototypes/stream-composer`](../../prototypes/stream-composer/README.md)。发现冲突时以本文为准，并把冲突报告给维护者，不得自行选择其中一方继续实现。原型中的示例数据（例如某个 Agent 不支持插话）不构成要求。

**等级。**

| 等级 | 含义 |
| --- | --- |
| 必须 | 不满足即不能合入。确有困难时先提出修改本文，不能在实现中绕过 |
| 不得 | 明确禁止的做法，同样不能合入 |
| 建议 | 默认应当这样做；偏离时须在变更说明中给出理由 |

**验证方式。** 每条要求列出一种或多种：

- `static:<检查名>` — `node tools/ui-checks/run.mjs static` 中的静态检查。
- `e2e` — `tools/ui-checks/e2e/acceptance.test.mjs` 中标题以该编号开头的验收测试。
- `perf` — 同上，但只在提供 `history-10k` 夹具的目标上运行。
- `manual` — 需要人工或在真实 MyGO WebView 中验证，不能由脚本替代。

`node tools/ui-checks/run.mjs requirements` 会检查本文：编号唯一、字段齐全、Checklist 编号存在且已勾选、声明的检查与测试确实存在，并列出尚无要求覆盖的已勾选 Checklist 项。运行方法、新增验收目标与例外标记见 [`tools/ui-checks/README.md`](../../tools/ui-checks/README.md)。

**交付门槛。** 合入默认呈现的改动前，`node tools/ui-checks/run.mjs all` 必须全部通过；因目标缺少夹具而跳过的测试不算通过，须在变更说明中列出。`manual` 项由评审者在变更说明中逐条确认。

**范围。** 只约束默认呈现。外部呈现包须满足 [05 号决议](../../.scratch/aibo-rev-architecture/issues/05-presentation-interface.md#answer)及其 2026-11-10 补充，可以使用不同的形式与位置，不受本文的视觉条款约束。

## 1. 总则

#### UI-GEN-01 · 必须 · 只通过公开呈现接口访问业务
- 条款：默认呈现读取数据、发起操作只经过与外部呈现包相同的公开呈现接口，不得调用外部包拿不到的私有入口。
- 依据：ui-interaction §0；决议 05
- Checklist：F20.1
- 验收：代码评审确认没有绕过公开接口的导入或调用。
- 验证：manual

#### UI-GEN-02 · 不得 · 在界面中推断执行结果
- 条款：界面不得自行解释厂商协议字段，也不得根据计时、按钮点击或文本内容推断“已投递”“已停止”“已暂停”等结果；这些状态只来自核心发布的状态。
- 依据：ui-interaction §1；决议 04、07、08
- Checklist：F04.7、F09.1
- 验收：代码评审确认状态文字与样式只由核心状态映射得到；e2e 中各“请求已发送”状态在核心确认前保持未确认样式。
- 验证：manual

#### UI-GEN-03 · 不得 · 为范围外功能放占位入口
- 条款：不得为未勾选的 Checklist 功能放置“即将推出”“敬请期待”等占位入口或文案。
- 依据：ui-interaction §0
- Checklist：F20.1
- 验收：源码中不出现禁用文案清单中的词。
- 验证：static:forbidden-copy

#### UI-GEN-04 · 必须 · 提供验收接口
- 条款：开发与测试构建必须提供 §11 的 `data-testid` 与夹具入口，使本文的 e2e 验收能够运行；正式发布构建可以移除夹具入口。
- 依据：本文 §11
- Checklist：F20.1
- 验收：e2e 套件在实现目标上全部可运行（无因缺少测试标识导致的失败）。
- 验证：e2e

## 2. 视觉

#### UI-VIS-01 · 必须 · 使用设计令牌
- 条款：颜色、字体、字号、间距、圆角、层级与动效使用 `docs/design/tokens/tokens.json`（经生成的 `tokens.css`）。组件样式中不得直接写颜色字面量；确需例外（例如 Agent 品牌色）时在同一行注明 `ui-check-allow: color <理由>`。
- 依据：ui-interaction §2
- Checklist：F20.1、F20.3
- 验收：静态检查无硬写颜色；样式中声明的令牌变量与 tokens.json 一致。
- 验证：static:no-hardcoded-color、static:token-sync

#### UI-VIS-02 · 必须 · 对比度
- 条款：tokens.json `contrast.pairs` 中的前景/背景组合在浅色与深色下都达到所列对比度；状态色会用作 12px 文字，按 4.5:1 要求。
- 依据：ui-interaction §2.1、§14
- Checklist：F20.3
- 验收：对比度检查全部通过。
- 验证：static:contrast

#### UI-VIS-03 · 必须 · 浅色与深色完整
- 条款：每个颜色令牌在浅色、深色下都有取值；深色不是浅色的自动反转。
- 依据：ui-interaction §2.1
- Checklist：F20.3
- 验收：令牌完整性检查通过。
- 验证：static:tokens-complete

#### UI-VIS-04 · 不得 · 小于 12px 的文字
- 条款：可读文字的字号不得小于 12px。纯图形徽记须标注 `role="img"` 与 `aria-label`，或 `aria-hidden="true"`，才可例外。
- 依据：ui-interaction §2.2
- Checklist：F20.1
- 验收：样式中没有小于 12px 的字号；运行时所有可见文字的计算字号不小于 12px。
- 验证：static:min-font-size、e2e

#### UI-VIS-05 · 不得 · 左侧信号条
- 条款：卡片、列表项、提示块、引用块不得用一侧加粗的彩色边条（`border-left`/`border-inline-start` 宽于 1px、`inset` 阴影竖条、绝对定位的窄竖条伪元素）表达状态或选中。所用 UI 库或设计系统明确规定时例外，须在同一行注明 `ui-check-allow: left-accent <出处>`。
- 依据：ui-interaction §2.4
- Checklist：F20.1
- 验收：静态与运行时检查都没有发现左侧信号条。
- 验证：static:no-left-accent、e2e

#### UI-VIS-06 · 必须 · 可见的焦点
- 条款：键盘焦点必须可见（2px `accent` 外描边或等效样式）。去掉 `outline` 时必须在同一规则或同一选择器的 `:focus-visible`/`:focus-within` 中提供替代样式，并注明 `ui-check-allow: outline <替代样式位置>`。
- 依据：ui-interaction §14
- Checklist：F20.1
- 验收：静态检查没有未说明的 `outline: none`。
- 验证：static:focus-outline

#### UI-VIS-07 · 必须 · 状态不只靠颜色
- 条款：状态标记（会话状态点、投递状态、请求状态）必须同时有文字或可读的 `aria-label`。
- 依据：ui-interaction §2.3、§14
- Checklist：F20.1
- 验收：会话列表的每个状态标记都有非空的可读名称。
- 验证：e2e

#### UI-VIS-08 · 必须 · 尊重减少动态效果
- 条款：使用动画的样式文件必须在 `prefers-reduced-motion: reduce` 下停用或弱化这些动画。
- 依据：ui-interaction §1、§14
- Checklist：F20.1
- 验收：含 `@keyframes` 的样式文件同时包含 `prefers-reduced-motion` 规则。
- 验证：static:reduced-motion

## 3. 导航与会话目录

#### UI-NAV-01 · 不得 · 独立的关闭会话入口
- 条款：默认界面用会话列表切换，不得提供“关闭会话”入口（包括禁用状态的菜单项）；离开会话不结束执行。
- 依据：ui-interaction §4.2；决议 03、05 Q23
- Checklist：F03.2、F03.3
- 验收：会话行菜单与会话头菜单中没有“关闭”项。
- 验证：e2e

#### UI-NAV-02 · 必须 · 排序与置顶
- 条款：会话默认按最近活动排序，置顶的会话排在分组顶部；仅切换查看不得改变顺序。
- 依据：ui-interaction §4.1、§4.2
- Checklist：F03.2
- 验收：切换到另一个会话后列表顺序不变；置顶某会话后它排在分组首位。
- 验证：e2e

#### UI-NAV-03 · 必须 · 有活动时归档先确认
- 条款：归档有活动（回合、目标、后台任务）的会话时，必须先列出这些活动并请用户选择停止后归档或取消；确认停止前不得归档。
- 依据：ui-interaction §4.2；决议 03
- Checklist：F03.5
- 验收：对执行中的会话选择归档 → 出现确认对话框，会话仍在列表中。
- 验证：e2e

#### UI-NAV-04 · 必须 · 会话详情默认隐藏
- 条款：引用授权、用量、适配器与原生会话信息默认不在工作台中展示，必须可从会话菜单打开。
- 依据：ui-interaction §8.3
- Checklist：F08.5
- 验收：初始界面不显示原生会话 ID；从会话头菜单打开“会话详情”后显示。
- 验证：e2e

#### UI-NAV-05 · 必须 · 工具面板区
- 条款：没有能力包提供面板时不显示工具面板区；有面板时默认收起，展开后标明提供它的插件与版本；面板故障不影响会话视图。
- 依据：ui-interaction §8.2；决议 09 修订
- Checklist：F17.1、F17.5
- 验收：初始无展开面板；展开后可见来源说明。
- 验证：e2e


#### UI-NAV-06 · 必须 · 新建会话与 Agent 检查
- 条款：新建会话时列出可用 Agent 及其适配器版本、验证状态、可执行文件与认证检查结果；检查未通过的 Agent 可以选中查看原因，但创建入口不可用。
- 依据：ui-interaction §4.3；决议 04
- Checklist：F03.1、F02.1、F02.3、F02.5、F02.6
- 验收：`Mod+N` 打开新建会话；选中检查未通过的 Agent 后“创建”不可用。
- 验证：e2e

#### UI-NAV-07 · 必须 · 未验证组合的标识
- 条款：当前会话的 Agent/适配器组合未经验证时，会话头显示“未验证”并说明含义；不得因此禁用全部功能。
- 依据：ui-interaction §5.1；决议 04 Q15
- Checklist：F02.5
- 验收：`untrusted`（其 Agent 为未验证组合）中会话头显示“未验证”。
- 验证：e2e

#### UI-NAV-08 · 必须 · 工作区的添加与移除
- 条款：添加工作区使用系统文件夹选择器并随后确认信任；移除工作区时说明保留业务历史、项目文件与暂停的队列，有活动时先确认停止。
- 依据：ui-interaction §4.2、§12
- Checklist：F01.1、F01.3
- 验收：在 MyGO 构建中验证系统文件夹选择器与移除确认。
- 验证：manual

#### UI-NAV-09 · 必须 · 已归档与离线历史可读
- 条款：已归档会话、Agent 不可用的会话都可以阅读与搜索；输入区替换为说明及恢复入口（取消归档、重新安装），不得显示为可发送。
- 依据：ui-interaction §4.2、§11；决议 06 Q34
- Checklist：F10.1、F12.3、F19.7
- 验收：从搜索打开已归档会话 → 显示历史消息与“取消归档后继续”。
- 验证：e2e
## 4. 会话流

#### UI-STREAM-01 · 必须 · 消息排布
- 条款：用户消息为右对齐的聊天气泡；Agent 回复不加气泡、占满正文宽度。
- 依据：ui-interaction §5.2
- Checklist：F04.1、F04.3
- 验收：用户消息右边缘贴近会话流正文区右边缘，左侧留白明显大于右侧。
- 验证：e2e

#### UI-STREAM-02 · 必须 · 投递状态
- 条款：用户消息必须区分并显示“已排队/投递中/已接收/结果未知”；“结果未知”使用 `unknown` 样式，只提供“核实”和“重新发送（可能重复执行）”。
- 依据：ui-interaction §5.2；决议 07
- Checklist：F09.2、F04.8
- 验收：`delivery-unknown` 中存在 `data-delivery="unknown"` 的用户消息，且提供上述两项操作。
- 验证：e2e

#### UI-STREAM-03 · 必须 · 跟随底部
- 条款：位于底部时新内容自动跟随；用户离开底部（滚轮、拖动滚动条、键盘或定位）后不得被自动拉回，并显示“有新内容”入口。
- 依据：ui-interaction §5.2；§15.1 发现 2
- Checklist：F04.2
- 验收：向上滚动后产生新内容 → 滚动位置不变，出现“有新内容”。
- 验证：e2e

#### UI-STREAM-04 · 必须 · 长会话
- 条款：1 万条记录的会话必须虚拟化渲染（挂载行数与可见区域相关，与总条数无关）；向前加载更早记录时可见内容不得跳动。
- 依据：ui-interaction §5.2、§15.1
- Checklist：F10.1
- 验收：`history-10k` 中挂载的行不超过 60；向前加载后首个可见条目位移不超过 2px。
- 验证：perf

#### UI-STREAM-05 · 不得 · 流式时阻塞主线程
- 条款：流式输出期间不得出现超过 50ms 的主线程阻塞；代码高亮等重计算必须移出主线程。
- 依据：ui-interaction §15.1 发现 1
- Checklist：F04.2、F04.3
- 验收：`history-10k` 中开始长回复，流式期间没有超过 50ms 的帧。
- 验证：perf

#### UI-STREAM-06 · 不得 · 恢复失效的请求
- 条款：失效的审批或追问（回合结束、连接更替、应用重启）只能显示为只读记录，不得出现可操作的选项。
- 依据：ui-interaction §7.1；决议 07
- Checklist：F05.1、F22.2
- 验收：`restart` 中原审批卡状态为已失效且没有选项按钮。
- 验证：e2e

#### UI-STREAM-07 · 必须 · 标示缺口与截断
- 条款：未能保存或被截断的内容必须以明确的分隔行标示，不得静默省略。
- 依据：ui-interaction §5.2；决议 06、10
- Checklist：F10.1、F22.3
- 验收：`turn-failed` 中可见缺口标示。
- 验证：e2e


#### UI-STREAM-08 · 必须 · 工具调用与思考摘要
- 条款：工具调用以单行卡片显示工具名、目标、状态与耗时，默认折叠、可展开查看参数与结果；提供者公开的思考摘要默认折叠为一行。展开状态属于共通呈现状态。
- 依据：ui-interaction §5.2
- Checklist：F04.4、F04.5
- 验收：工具卡片初始 `aria-expanded="false"`，点击后为 `true` 并显示结果。
- 验证：e2e

#### UI-STREAM-09 · 必须 · 恢复阅读位置
- 条款：切换会话后回来，恢复该会话的阅读位置；工作台布局与阅读状态在重启后恢复。
- 依据：ui-interaction §3、§5.2；决议 05 Q22
- Checklist：F21.4、F21.5
- 验收：在会话流中滚动到某位置 → 切换会话 → 切回，滚动位置误差不超过 5px。重启后的恢复人工验证。
- 验证：e2e、manual
## 5. 输入区

#### UI-COMP-01 · 不得 · 组字期间发送
- 条款：输入法组字期间（`isComposing` 为真，或 `keyCode` 为 229）按 Enter 不得发送或排队。
- 依据：ui-interaction §5.3
- Checklist：F04.1
- 验收：派发组字中的 Enter 后，输入内容保留，会话流与队列都没有新增。
- 验证：e2e、manual

#### UI-COMP-02 · 必须 · 执行中的发送进入队列
- 条款：会话执行中按 Enter 或主按钮，输入进入宿主队列，不得直接投递；此时主按钮文案为“排队”。
- 依据：ui-interaction §5.3、§6.2；决议 07 Q35
- Checklist：F09.1
- 验收：`running-approval` 中输入并按 Enter → 输入框清空，队列标签数量加 1，会话流没有新增用户消息。
- 验证：e2e

#### UI-COMP-03 · 必须 · 插话的独立入口与禁用原因
- 条款：执行中提供独立的插话入口。Agent 不支持时保持可聚焦并使用 `aria-disabled="true"`，提供可读的原因（`aria-describedby` 或等效），不得使用会移除焦点的 `disabled`。
- 依据：ui-interaction §1、§5.3；决议 04 Q17
- Checklist：F09.6
- 验收：`running-approval` 中插话控件 `aria-disabled="true"`、可聚焦、可读原因非空。
- 验证：e2e

#### UI-COMP-04 · 必须 · 触发符候选
- 条款：`@`（文件/目录）、`#`（会话）、`/`（命令与 Skills）打开候选；查询可以包含 `/`；`↑↓` 选择、Enter 确认、Esc 关闭；候选打开期间焦点保持在输入框。
- 依据：ui-interaction §5.3；§15.1 发现 4
- Checklist：F06.1、F06.2、F07.1、F04.9
- 验收：输入 `@src/co` 出现候选且焦点仍在输入框；按 Enter 后生成路径附件。
- 验证：e2e

#### UI-COMP-05 · 必须 · 引用固定版本
- 条款：会话引用在加入输入框时固定版本，附件上显示固定时间与范围；来源之后的新内容不得自动加入，只能由用户选择“更新到最新版本”。
- 依据：ui-interaction §5.3；决议 06 Q28
- Checklist：F07.1、F07.2、F07.5
- 验收：通过 `#` 加入引用后，引用附件文字包含“固定”与时间。
- 验证：e2e


#### UI-COMP-11 · 必须 · 全文引用的可用性
- 条款：引用范围可选近 N 轮或全文；接收 Agent 不支持全文读取工具时，“全文”禁用并说明“仅能附带摘录”。
- 依据：ui-interaction §5.3；决议 06
- Checklist：F07.6
- 验收：用不支持全文读取的 Agent 打开引用菜单，人工核对。
- 验证：manual
#### UI-COMP-06 · 必须 · 设置未确认时不发送
- 条款：模型、权限等必要设置在 Agent 确认前，发送入口必须不可用并说明原因；不得按旧值悄悄发送。
- 依据：ui-interaction §5.3；决议 07
- Checklist：F08.1、F08.6
- 验收：更改模型后、确认前主按钮 `aria-disabled="true"`；确认后恢复。
- 验证：e2e

#### UI-COMP-07 · 必须 · 草稿随会话保存
- 条款：草稿属于共通呈现状态，切换会话后回来必须恢复。
- 依据：ui-interaction §3；决议 05 Q22
- Checklist：F04.6
- 验收：输入草稿 → 切换到其他会话 → 切回，内容一致。
- 验证：e2e

#### UI-COMP-08 · 必须 · 停止的各项结果分别显示
- 条款：停止当前执行后，必须分别显示当前回合、未发送队列、持续目标、后台任务的结果；确认前显示“请求已发送”；不得提供“视为已停止”。
- 依据：ui-interaction §6.1；决议 08 Q48
- Checklist：F04.7、F09.4
- 验收：`running-approval` 中点击停止 → 出现停止摘要，立即显示“停止请求已发送”，并写明后台任务不受影响。
- 验证：e2e、static:forbidden-copy


#### UI-COMP-09 · 必须 · 图片附件
- 条款：粘贴或拖入图片生成图片附件；可以只发送图片；待发送附件可以逐个移除。
- 依据：ui-interaction §5.3
- Checklist：F06.3、F06.4、F06.5
- 验收：向输入框粘贴图片 → 出现图片附件；不输入文字直接发送 → 会话流新增用户消息；附件上的移除按钮可删除附件。
- 验证：e2e

#### UI-COMP-10 · 必须 · 不支持的执行设置
- 条款：推理强度、服务档位等 Agent 不支持的设置以禁用形式保留并说明原因（`aria-disabled` 加可读原因），不得隐藏。
- 依据：ui-interaction §5.3；决议 04 Q17
- Checklist：F08.2、F08.3
- 验收：`running-approval`（Agent 不提供服务档位）中 Fast 控件 `aria-disabled="true"` 且原因非空。
- 验证：e2e
## 6. 状态条与浮层

#### UI-STRIP-01 · 必须 · 状态条内容
- 条款：输入框上方的状态条显示待处理请求、持续目标、子 Agent、后台任务与队列；都没有时整行不出现。
- 依据：ui-interaction §8.1
- Checklist：F11.1、F12.1、F12.4、F09.1
- 验收：`delivery-unknown` 中依次出现目标、后台任务、队列标签；`idle-empty` 中没有状态条。
- 验证：e2e

#### UI-STRIP-02 · 必须 · 浮层行为
- 条款：状态条浮层在输入框上方展开，同一时间只开一个，Esc 收起；打开与收起不改变草稿。
- 依据：ui-interaction §8.1
- Checklist：F09.1
- 验收：打开队列浮层 → 按 Esc 关闭；草稿保持不变。
- 验证：e2e

#### UI-STRIP-03 · 必须 · 结果未知的重发
- 条款：结果未知的队列项只能核实或经二次确认后“重新发送（可能重复执行）”；原记录保留并标注已重发；其余队列保持暂停。
- 依据：ui-interaction §6.2；决议 07 Q42
- Checklist：F09.2、F09.5
- 验收：`delivery-unknown` 中点击重新发送 → 出现确认对话框；确认后原消息仍在并标注已重发，队列仍为暂停。
- 验证：e2e

#### UI-STRIP-04 · 必须 · 目标控制的范围与确认
- 条款：目标的暂停/恢复/清除只作用于目标，按钮附近说明影响范围；请求发出后在 Agent 确认前显示“请求已发送”；Agent 不支持时禁用并说明原因。
- 依据：ui-interaction §8.1；决议 08 Q52
- Checklist：F11.1、F11.3
- 验收：`delivery-unknown` 中暂停目标 → 先显示“暂停请求已发送”，确认后显示“已暂停”。
- 验证：e2e

#### UI-STRIP-05 · 必须 · 子 Agent 与后台任务详情
- 条款：子 Agent 与后台任务详情在主区以只读方式打开，提供返回会话；后台任务可单独取消，且说明不影响当前回合。
- 依据：ui-interaction §8.1；决议 08 Q49
- Checklist：F12.2、F12.5、F12.6
- 验收：`running-approval` 中从子 Agent 卡片进入详情 → 输入区隐藏、可返回。
- 验证：e2e


#### UI-STRIP-06 · 必须 · 删除与清空队列
- 条款：可以删除单条未开始投递的队列项；清空队列须确认，只删除未开始投递的项，结果未知的项保留。
- 依据：ui-interaction §6.2；决议 07
- Checklist：F09.3
- 验收：`delivery-unknown` 中清空队列 → 确认 → 只剩结果未知的一项。
- 验证：e2e

#### UI-STRIP-07 · 必须 · 目标预算与用量
- 条款：目标的预算与用量逐项显示 Agent 报告的数值、单位与来源；未报告的项显示“未报告”，不得显示为 0，也不得用会话总用量代替。
- 依据：ui-interaction §8.1；决议 08
- Checklist：F11.2
- 验收：`delivery-unknown` 的目标浮层中，未报告的时长显示“未报告”。
- 验证：e2e
## 7. 审批与追问

#### UI-APPR-01 · 必须 · 就地审批与数字键
- 条款：审批卡在会话流中就地出现，选项按 Agent 提供的原文与顺序显示；聚焦卡片时数字键选择对应选项。
- 依据：ui-interaction §7.1；决议 07
- Checklist：F05.1、F05.2
- 验收：`running-approval` 中聚焦审批卡并按 1 → 进入“已提交”状态。
- 验证：e2e

#### UI-APPR-02 · 必须 · 提交后锁定
- 条款：提交后选项立即锁定，Agent 确认前显示“已提交，等待确认”；不得允许再次选择。
- 依据：ui-interaction §7.1
- Checklist：F05.1
- 验收：提交后审批卡内没有可点击的选项按钮。
- 验证：e2e

#### UI-APPR-03 · 必须 · 追问答案随状态保存
- 条款：追问与表单中正在填写的答案切换会话后保留，不得自动提交。
- 依据：ui-interaction §7.2；决议 05
- Checklist：F05.3
- 验收：`pending-question` 中选择一项 → 切换会话再切回，选择保留且问题仍待回答。
- 验证：e2e


#### UI-APPR-04 · 必须 · 模式转换
- 条款：Agent 请求切换模式时以审批卡呈现，选项照 Agent 原样；生效后执行设置栏的权限模式随之更新，会话流插入一条系统行。
- 依据：ui-interaction §7.3
- Checklist：F05.4
- 验收：用支持模式转换的 Agent 触发一次切换，人工核对。
- 验证：manual
## 8. 信任、生命周期与恢复

#### UI-LIFE-01 · 不得 · 重启后自动执行
- 条款：应用重启后恢复历史、草稿、队列和目标状态，但不得自动执行任何操作；恢复连接须由用户明确触发，并说明 Agent 可能因此继续执行。
- 依据：ui-interaction §10.3；决议 03、07
- Checklist：F03.4、F22.2
- 验收：`restart` 中出现恢复卡与“恢复连接”，等待 2 秒后会话没有进入执行中，也没有新增消息。
- 验证：e2e

#### UI-LIFE-02 · 必须 · 未信任工作区
- 条款：未信任工作区可以阅读历史、编辑草稿；发送入口不可用并说明原因，提供“信任此工作区”。
- 依据：ui-interaction §10.1；决议 07 Q39
- Checklist：F01.2
- 验收：`untrusted` 中显示信任提示，主按钮 `aria-disabled="true"`，输入框可输入。
- 验证：e2e

#### UI-LIFE-03 · 必须 · 关键保存失败
- 条款：关键保存失败时显示不可关闭的横幅，分别说明未保存的内容与停止未确认的活动；发送入口不可用。
- 依据：ui-interaction §10.4；决议 10 Q61
- Checklist：F22.1、F22.3
- 验收：`save-failure` 中横幅存在、没有关闭按钮，主按钮 `aria-disabled="true"`。
- 验证：e2e

#### UI-LIFE-04 · 必须 · 呈现故障回退与恢复入口
- 条款：外部呈现明确失败时自动回退默认呈现并告知原因；无响应时不自动回退，恢复入口（应用菜单、Dock/通知区域菜单、原生恢复窗口）必须在 WebView 之外可达。
- 依据：ui-interaction §10.5；决议 05 Q24
- Checklist：F20.7
- 验收：在 MyGO 中注入加载失败与页面死循环，分别观察自动回退与原生菜单可用。
- 验证：manual

#### UI-LIFE-05 · 必须 · 退出与后台运行
- 条款：关闭最后一个窗口时应用留在后台并提示一次；有活动时退出必须先确认，停止未确认时提供“继续等待”与“强制退出”。Windows 常驻通知区域图标。
- 依据：ui-interaction §10.2；决议 03
- Checklist：F22.2
- 验收：在 MyGO 的 macOS 与 Windows 构建中分别验证。
- 验证：manual


#### UI-LIFE-06 · 必须 · 诊断入口
- 条款：设置中显示诊断日志的占用、上限与保留期，并能打开日志目录；诊断内容不包含完整业务正文。
- 依据：ui-interaction §11；决议 10 Q62、Q65
- Checklist：F22.4、F22.5
- 验收：人工核对设置 › 诊断。
- 验证：manual

#### UI-LIFE-07 · 必须 · 宿主操作确认偏好
- 条款：设置中可配置归档、清空队列等宿主操作是否二次确认；“重新发送（可能重复执行）”始终确认，不可关闭。
- 依据：ui-interaction §11
- Checklist：F21.6
- 验收：关闭全部可选确认后，重新发送仍弹出确认（UI-STRIP-03 的 e2e 覆盖默认情况）。
- 验证：manual

#### UI-LIFE-08 · 必须 · 扩展包管理
- 条款：设置中分别管理 Agent 包、能力包与呈现包：本地安装、启用/停用、卸载、升级。升级显示受影响会话并说明在相关活动结束后生效；卸载 Agent 包时说明相关会话仍可阅读。呈现包可设为当前，切换后恢复共通呈现状态。
- 依据：ui-interaction §11；决议 09
- Checklist：F19.1、F19.2、F19.3、F20.4、F20.5、F20.6
- 验收：在 MyGO 构建中安装、升级、卸载各类包并切换呈现包，人工核对。
- 验证：manual

#### UI-LIFE-09 · 必须 · 平台验收
- 条款：本文所有 e2e 验收须在 MyGO 的 macOS 与 Windows 构建中通过，并完成 ui-interaction §15 的平台验证项。
- 依据：ui-interaction §15
- Checklist：F25.1、F25.2
- 验收：两个平台的 e2e 报告与 §15 验证记录。
- 验证：manual
## 9. 可访问性、键盘与语言

#### UI-A11Y-01 · 必须 · 自动可访问性检查
- 条款：各夹具的初始界面不得有 axe-core 判定为 serious 或 critical 的问题。
- 依据：ui-interaction §14
- Checklist：F20.1
- 验收：axe 检查通过。
- 验证：e2e

#### UI-A11Y-02 · 必须 · 弹层焦点
- 条款：菜单、对话框、浮层关闭后焦点回到触发元素。
- 依据：ui-interaction §14
- Checklist：F20.1
- 验收：打开会话菜单后按 Esc，焦点回到菜单按钮。
- 验证：e2e

#### UI-A11Y-03 · 必须 · 快捷键
- 条款：实现 ui-interaction §13 的快捷键；第一版不提供自定义。
- 依据：ui-interaction §13
- Checklist：F20.1
- 验收：`Mod+K` 打开搜索与命令面板，Esc 关闭。其余快捷键人工核对。
- 验证：e2e、manual

#### UI-I18N-01 · 必须 · 中英文与跟随系统
- 条款：界面提供中文、英文与跟随系统；切换即时生效；文案不得硬编码在组件中。
- 依据：ui-interaction §14
- Checklist：F21.1、F21.2、F21.3
- 验收：切换语言后无残留另一种语言的界面文字；代码评审确认文案经由语言资源。
- 验证：manual

## 10. 搜索

#### UI-SEARCH-01 · 必须 · 统一搜索与命令
- 条款：`Mod+K` 打开唯一的全局搜索与命令入口；结果按命令、会话、消息、文件、Skills、设置、扩展包分组；来自已归档或已移除来源的结果带标签。
- 依据：ui-interaction §9；决议 10 Q63
- Checklist：F10.2、F10.3、F10.5、F10.6、F10.7
- 验收：搜索已归档会话的标题，结果带“已归档”标签。
- 验证：e2e

## 11. 验收接口（测试契约）

开发与测试构建必须提供以下标识，供 `tools/ui-checks` 的 e2e 验收使用。标识只用于测试，不改变界面结构的自由度；同一含义的元素可以换成任意标签和布局。

### 11.1 `data-testid`

| testid | 附加属性 | 含义 |
| --- | --- | --- |
| `session-row` | `data-session-id`、`data-status` | 会话列表中的一行 |
| `session-status` | — | 会话行中的状态标记（须有可读名称） |
| `session-menu-button` | — | 会话行菜单按钮 |
| `session-head-menu` | — | 会话头菜单按钮 |
| `stream` | — | 会话流滚动容器 |
| `message-user` | `data-delivery` | 用户消息（气泡本体） |
| `message-agent` | — | Agent 回复 |
| `approval-card` | `data-state`：`pending`/`submitted`/`resolved`/`expired` | 审批卡 |
| `question-card` | `data-state` | 追问卡 |
| `stop-summary` | — | 停止各项结果 |
| `recovery-card` | — | 重启后的恢复卡 |
| `gap-marker` | — | 缺口或截断标示 |
| `new-content` | — | “有新内容”入口 |
| `subagent-card` | — | 会话流中的子 Agent 卡片 |
| `detail-view` | — | 主区中的子 Agent/后台任务详情 |
| `detail-back` | — | 返回会话 |
| `composer-input` | — | 输入框 |
| `composer-primary` | `aria-disabled` | 发送/排队 |
| `composer-steer` | `aria-disabled`、`aria-describedby` | 插话 |
| `composer-stop` | — | 停止 |
| `reference-chip` | — | 会话引用附件 |
| `attachment-chip` | — | 文件、目录、图片附件 |
| `trigger-picker` | — | `@ # /` 候选列表 |
| `model-select` | `data-applying` | 模型选择 |
| `status-strip` | — | 状态条 |
| `strip-chip` | `data-kind`：`pending`/`goal`/`sub`/`task`/`queue` | 状态条标签 |
| `strip-popover` | `data-kind` | 状态条浮层 |
| `queue-item` | `data-unknown` | 队列项 |
| `confirm-dialog` | — | 需要用户确认的对话框 |
| `session-details` | — | 会话详情 |
| `trust-block` | — | 未信任提示 |
| `banner-save-failure` | — | 关键保存失败横幅 |
| `tool-bar`、`tool-panel`、`tool-panel-source` | — | 工具面板区、面板、来源说明 |
| `command-palette` | — | 搜索与命令面板 |

### 11.2 夹具

页面地址加 `?fixture=<名称>` 时进入对应初始状态；不认识的名称必须报错而不是忽略。

| 名称 | 初始状态 |
| --- | --- |
| `running-approval` | 当前会话执行中；1 个待审批；队列 1 项；1 个子 Agent 执行中；Agent 不支持插话；输入框有 1 个会话引用 |
| `delivery-unknown` | 当前会话空闲；1 项投递结果未知且队列暂停（共 3 项）；持续目标推进中且支持暂停；1 个后台任务运行中 |
| `pending-question` | 当前会话有 1 个待回答的选择题 |
| `turn-failed` | 当前会话回合失败、队列暂停，含 1 处缺口标示 |
| `untrusted` | 当前会话所在工作区未信任 |
| `idle-empty` | 当前会话为空闲的新会话，无队列、目标与任务 |
| `restart` | 模拟应用重启后：当前会话有失效审批与恢复卡 |
| `save-failure` | 关键保存失败 |
| `history-10k` | 当前会话含 10,000 条历史（性能验收用，可选） |

目标在 `ui-checks.config.json` 中声明自己支持的夹具；不支持的夹具对应的测试显示为跳过，不算通过。

### 11.3 排除区与例外标记

- 非产品界面（开发工具栏、调试面板）放在 `data-ui-check="ignore"` 的容器内，运行时样式检查与 axe 检查会跳过它们。产品界面不得使用此标记。
- 静态检查的例外须在同一行或上一行注明 `ui-check-allow: <规则> <理由>`；整块的非产品样式可用 `ui-check-allow-begin: <规则|all> <理由>` 与 `ui-check-allow-end` 包围。规则名为 `color`、`left-accent`、`font-size`、`outline`。没有理由的例外视为违规。
