# 跨框架呈现与桌面容器的可行边界

研究日期：2026-11-01。对应架构研究票：《跨框架呈现与桌面容器的可行边界》。

本次只核查浏览器标准、Tauri 和 Electron 的公开接口；没有运行原型、测量性能或验证 macOS / Windows 行为。
本文不决定 UI 框架、桌面容器、实现语言或呈现包格式；候选方式与接口建议均需后续决策。

## 已知约束与研究结论

用户已接受：默认 UI 可以使用框架，但内核、共享交互状态和呈现接口必须独立于它。
需要通过完整默认 UI 替换、异框架呈现包加载验证独立性；呈现包可定制布局、控件、业务视图，并支持故障回退。
目标系统为 macOS、Windows；不要求直接兼容旧包。

官方机制足以支撑跨框架通信与独立装载，但不自动保证工作台可替换、权限隔离或故障恢复。
必须区分三个问题：数据和交互契约是否独立、代码处于什么文档与权限边界、故障会影响哪些进程。
下面的证据支持建立这些边界；不能据此宣称任一候选已通过产品验收。

## 浏览器提供的边界

### 数据、命令与事件

结构化序列化支持跨 realm 传递特定数据；函数、部分平台对象、Proxy 等会触发 `DataCloneError`，并非所有对象细节都会保留。
因此组件实例、框架响应式代理和闭包不能直接当作通用消息载荷。[WHATWG：结构化数据](https://html.spec.whatwg.org/multipage/structured-data.html)

`window.postMessage` 支持跨文档通信，`MessageChannel` / `MessagePort` 支持独立通道。
规范要求接收方验证来源与消息内容，并提醒限制消息处理频率；通道本身不定义应用状态恢复协议。[WHATWG：Web Messaging](https://html.spec.whatwg.org/multipage/web-messaging.html)

**对 Aibo 的推论：**可以用可序列化数据、带结果的命令和显式事件组织呈现契约，但这些是应用协议设计，不是浏览器现成的工作台模型。
只把业务逻辑移到纯函数，仍不足以证明独立性：若草稿、权限请求、引用选择或会话切换只能通过某框架 store / hook 操作，替换 UI 仍会改动共享实现。

### DOM、文档、进程不能混为一层

Custom Elements 提供通用 DOM 元素及连接、断开等生命周期回调；不同框架可围绕这一接口封装组件。
断开回调提供清理入口，不会替应用取消订阅、后台任务或外部资源。[HTML：Custom Elements](https://html.spec.whatwg.org/multipage/custom-elements.html)

Shadow DOM 提供树与样式封装，仍属于宿主文档；它不是独立 JavaScript 运行环境，也不是进程或权限沙箱。
样式封装和授权应分别设计。[DOM：Shadow trees](https://dom.spec.whatwg.org/#shadow-trees)、[CSS Shadow：封装模型](https://drafts.csswg.org/css-shadow-1/)

`iframe` 提供子文档，可用 `sandbox` 限制能力；同源内容同时开放 `allow-scripts` 与 `allow-same-origin`，可能移除 sandbox 属性再重载以逃离限制。
iframe 的标准接口不能证明每个子文档独占一个 OS 进程。[HTML：iframe sandbox](https://html.spec.whatwg.org/multipage/iframe-embed-object.html#attr-iframe-sandbox)

焦点管理、顺序键盘导航和用户激活属于浏览器交互模型；它们不会因消息接口与框架无关而消失。
跨文档输入、快捷键和浮层协作仍需验证。[HTML：用户交互](https://html.spec.whatwg.org/multipage/interaction.html#focus)

剪贴板规范包含焦点、权限和用户激活条件；当前规范草案不能替代各系统 WebView 的兼容性验证。
呈现包不宜把“能收到点击事件”当作“可以完成所有剪贴板操作”的证明。[W3C：Clipboard API](https://www.w3.org/TR/clipboard-apis/)

## 三种可研究的呈现接入形态

这些是候选边界，不是互斥选型，也不是按优先级排序；整套默认 UI 与局部呈现包可以采用不同边界。

| 形态 | 已有机制 | 可以据此得到的能力 | 不能直接承诺的结果 |
| --- | --- | --- | --- |
| 同一文档挂载独立组件，可用 Custom Elements / Shadow DOM | DOM 接口、组件生命周期、样式封装 | 各包自行选择内部框架，经共同 DOM / 数据接口接入 | 全局 JS、权限及故障隔离；包卡死后的可靠回退 |
| 独立 iframe 文档 | 独立文档、sandbox、消息通道 | 单独加载构建产物，显式规定跨文档交互 | 每包独占进程；天然一致的焦点、弹层、剪贴板体验 |
| 原生窗口内组合多个 WebView / WebContentsView | 容器的视图、生命周期与 IPC API | 宿主管理视图边界，桥接受控原生能力 | 所有平台相同的故障通知；每视图独占进程；自动资源释放 |

前两行由上述浏览器标准支持；第三行的具体限制见下方容器证据。
“独立框架构建产物”不等于“可完整替换工作台”：还要验证导航、输入、审批和状态恢复能否都走共同契约。

## Tauri v2：能力与需要单独验证的部分

Tauri 的 Core 管理原生能力，前端运行于系统 WebView；Windows 使用 WebView2，macOS 使用 WKWebView。
因此同一套前端 API 不能替代对两个系统引擎的兼容性验证。[Tauri：进程模型](https://v2.tauri.app/concept/process-model/)

Capabilities 可按窗口 / WebView 配置 IPC 权限，多个 capability 的权限会合并。
应用经 `invoke_handler` 注册的自定义命令默认允许所有窗口 / WebView 使用；限制这些命令需配置应用命令清单等机制。
仅声明呈现包 capability，不足以证明自定义文件或会话命令已收紧。[Tauri：Capabilities](https://v2.tauri.app/security/capabilities/)

官方 Events 机制适用于较少数据，不面向高吞吐、低延迟；事件载荷为 JSON 字符串，缺少事件数据 / 通道的细粒度 capability 控制。
Channels 面向快速、有序的数据传递；是否适合大量会话并发更新仍需实测。[Tauri：调用前端](https://v2.tauri.app/develop/calling-frontend/)

本次查阅的 API 文档版本为 Tauri **2.12.5**：子视图 `WebviewBuilder` 仍要求 `unstable` feature。
文档还指出 Windows 同步命令 / 事件处理器中创建 WebView 的死锁限制，需要异步命令或独立线程等处理。
这项限制针对该子 WebView API，不能泛化为所有 Tauri 窗口 API 都不稳定。[Tauri：WebviewBuilder](https://docs.rs/tauri/2.12.5/tauri/webview/struct.WebviewBuilder.html)

同版本 `on_web_content_process_terminate` 回调受 macOS / iOS 条件编译控制，文档明确 Windows 不支持该接口。
这说明该回调不能直接形成双平台统一故障检测；并不证明 Windows 无其他恢复办法。[Tauri：应用接口源码](https://docs.rs/tauri/2.12.5/src/tauri/app.rs.html#1903-1919)

Asset protocol 通过配置的 scope 控制磁盘文件向 WebView 暴露，运行时选择的路径也必须满足相应范围。
呈现包资源、会话附件和用户选择的文件应分别明确授权来源，不能仅以磁盘路径作为权限依据。[Tauri：Asset protocol scope](https://v2.tauri.app/security/asset-protocol/)

Tauri 的 Isolation Pattern 是通过隔离应用过滤前端到 Core 的 IPC；它不等同于每个呈现包拥有独立渲染进程。
不能仅凭名称推导故障隔离级别。[Tauri：Isolation Pattern](https://v2.tauri.app/concept/inter-process-communication/isolation/)

## Electron：能力与需要单独验证的部分

Electron 区分主进程与渲染进程，原生能力应通过明确的进程间接口提供。
`contextIsolation` 分离 preload 与页面全局环境；官方示例要求为具体功能封装接口，避免直接暴露任意 IPC 发送能力。
[Electron：进程模型](https://www.electronjs.org/docs/latest/tutorial/process-model)、[Context Isolation](https://www.electronjs.org/docs/latest/tutorial/context-isolation)

`contextBridge` 可以代理函数，其他值会复制并冻结；桥接后修改源对象不会同步修改另一侧。
它不能直接当成跨边界响应式 store，更新需要显式协议。[Electron：contextBridge](https://www.electronjs.org/docs/latest/api/context-bridge)

主进程提供 `MessagePortMain` / `MessageChannelMain`；端口通过 `postMessage` 转移，普通 `send` / `invoke` 不负责端口转移。
消息通道提供传输能力，订阅注销、慢消费者处理和重连状态仍是应用职责。[Electron：MessagePorts](https://www.electronjs.org/docs/latest/tutorial/message-ports)

`WebContentsView` 可组合独立 WebContents 视图；使用 `BaseWindow` 时，关闭窗口不会自动销毁其子视图的 WebContents，必须自行清理。
这是呈现包卸载协议需要覆盖的资源生命周期。[Electron：WebContentsView](https://www.electronjs.org/docs/latest/api/web-contents-view)、[BaseWindow](https://www.electronjs.org/docs/latest/api/base-window)

WebContents 提供 `render-process-gone`、`unresponsive`、`responsive`、`destroyed` 等事件。
官方同时说明有些 WebContents 会共享渲染进程；故障事件存在，不代表任意视图组合已实现进程级隔离。[Electron：WebContents](https://www.electronjs.org/docs/latest/api/web-contents)

自定义资源协议注册绑定 Electron session；若视图使用自定义 partition / session，需在对应 session 注册协议。
资源装载与视图身份、生命周期不能作为完全独立的问题处理。[Electron：protocol](https://www.electronjs.org/docs/latest/api/protocol)

## 对后续架构设计的推论

以下是由证据得到的待讨论约束，不是本次新增架构决议：

- 共享契约要覆盖数据、命令结果、持续事件与版本协商；不能要求包导入默认 UI 的组件类型、store 或路由。
- 区分需要跨呈现恢复的状态与局部视图状态；草稿、引用、待处理审批等是否跨 UI 保留，需要逐项定义，不能隐含在组件生命周期里。
- 明确装载、就绪、订阅、卸载、失联、重挂载的责任；旧实例迟到的回复不能影响新实例或错误会话。
- 为快照、增量更新、顺序、批处理、积压和重新同步定义行为；“支持流式”不足以描述大历史和并发会话。
- 将原生操作及资源访问限定为明确能力；框架独立、权限隔离和故障隔离应分别验收。
- 回退 UI 需要可读取的共享状态和独立恢复入口；仅捕获组件异常不足以覆盖进程退出或事件循环卡死。

## 最少必要的原型验收

原型应同时记录 macOS / Windows 的 OS、WebView / Electron 版本；以下尚未执行，也没有据此设定性能结论。

| 场景 | 需要拿到的证据 |
| --- | --- |
| 完整替换默认 UI，并加载另一框架的呈现包 | 两套实现都完成会话列表、流式消息、输入草稿、跨 Agent 引用、权限审批；核心与共享契约无需修改 |
| 流式生成时切换、重载、卸载呈现包 | 执行与共享状态按约定继续；重新订阅可恢复；无重复提交、迟到审批、残留订阅或资源泄漏 |
| 包异常、死循环、渲染进程退出 | 记录实际进程共享关系、通知路径及受影响范围；能否进入默认回退视图，而不只验证组件异常捕获 |
| 输入与原生交互 | 中文 IME、焦点循环、快捷键、选区、跨边界菜单 / 弹层、图片剪贴板、拖放和文件选择在双平台的结果 |
| 大量历史与并发更新 | 测量首屏、输入响应、内存、消息积压；慢消费者及断连后能否重新同步，随后据需求确定预算 |
| 权限与资源生命周期 | 包只获得指定资源与命令；卸载后失效；更换 session / partition 或重启后，附件与包资源的授权行为可解释 |

## 尚待决策的问题

- 呈现包的信任级别，以及“故障回退”需要覆盖组件异常、事件循环阻塞还是整个渲染进程退出。
- 完整默认 UI 与局部呈现包是否采用相同边界；跨文档 / 原生视图组合增加的交互工作是否值得。
- UI、ACP 会话执行和持久化各自的存活周期，以及 UI 关闭 / 重启时用户预期的行为。
- 需要支持的历史规模、并发流量与可接受延迟；在这些要求明确前，无法比较候选方式的维护和性能成本。

结论适用范围：标准与官方 API 层面的可行性和限制；不构成选型结论，也不代表已经证明跨框架、双平台或故障恢复验收通过。
