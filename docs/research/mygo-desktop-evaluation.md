# MyGo 桌面方案调研

研究日期：2026-11-04。对象为 [egoist/mygo](https://github.com/egoist/mygo)，不是同名项目。

本次核查官方文档与源码，没有安装 MyGo、构建示例、运行原型或测量 Aibo 性能。源码依据固定的 main 提交 [`6842ad5e0aa2cd168166e0a4ae12c0cd0f527a3d`](https://github.com/egoist/mygo/tree/6842ad5e0aa2cd168166e0a4ae12c0cd0f527a3d)；官网为滚动文档，其页脚版本不保证每项功能已进入同名发布版。本文不形成技术栈决议。

> 2026-11-10 更新：用户已决定采用 MyGO 的系统 WebView 路线与 Go 宿主，见 [ADR 0003](../adr/0003-desktop-and-ui-stack.md)。下文“建议的验证边界”转为选型后的必做验证。

## 结论与适用范围

MyGo 同时提供 Go + 系统 WebView 和 Go 自绘原生 UI。对继续使用 Web 呈现包的 Aibo，值得验证的是前一条路线：它可能减轻 Rust 工具链与原生编译依赖的维护成本，但不改变 macOS 使用 WebKit、Windows 使用 WebView2 的事实。原生 UI 是另一种界面实现路线，不能直接承接任意 JavaScript 框架的呈现包。[官方介绍](https://mygo.egoist.dev/docs)

当前 Aibo 依据 2026-11-03 的[呈现接口决议](../../.scratch/aibo-rev-architecture/issues/05-presentation-interface.md#answer)：每次启用一套完整呈现包、用户信任其前端代码、共用无界面交互模块；明确失败自动回退，持续无响应由 UI 外部入口手动恢复。本文据此评估，不额外要求敌对插件沙箱或逐包独占进程。

## 构建与运行边界

| 方面 | 已核查的机制 | 对 Aibo 的含义或限制 |
| --- | --- | --- |
| Go 与本机 API | 框架默认以 `CGO_ENABLED=0` 构建；macOS/Linux 通过 purego 在运行时调用本机库，Windows 通过 syscall | 无 cgo 不等于不用系统库；可以省去框架自身的目标 C 编译工具链 |
| Web UI | macOS WKWebView、Windows WebView2、Linux WebKitGTK | 仍需双引擎兼容验证；没有核查到打包固定 Chromium 的官方路线 |
| 原生 UI | Go 的 `ui` 包自绘，后端使用 Metal、Direct3D 11、OpenGL 等 | 不启动 WebView，但需要用 Go 实现视图；不是原有 Web UI 的无成本替换 |
| Bun | 用于前端及开发工具，不随应用作为运行时分发 | 不应把它等同于 Electrobun 的 Bun 主进程 |

上述机制来自[官方架构](https://mygo.egoist.dev/docs/architecture)。本次固定源码的 [go.mod](https://github.com/egoist/mygo/blob/6842ad5e0aa2cd168166e0a4ae12c0cd0f527a3d/go.mod) 要求 Go **1.27.3**，实施时应按锁定版本检查工具链。

“任意机器编译其他平台”主要适用于 Go 可执行文件。发行流程仍有条件：macOS 签名与 DMG 要在 Mac 上完成；其他系统生成 Windows 安装程序需要 NSIS。应用若引入必须使用 cgo 的依赖，则要重新准备目标 C 编译器。官方 CLI 支持显式启用 `CGO_ENABLED=1`。[分发与交叉编译](https://mygo.egoist.dev/docs/distribution#platforms)

功能依赖也会增加产物：官方 SQLite 插件使用预编译 SQLite 原生库；Terminal 插件加载 libghostty-vt。CLI 可按平台下载、校验并打包原生库，缺少对应平台资产会阻止构建。因此“无 cgo”“小 Go 程序”不能外推为任意功能组合都只有一个几 MB 文件。[插件介绍](https://mygo.egoist.dev/docs/plugins)、[原生库打包](https://mygo.egoist.dev/docs/distribution#native-libraries-of-packages)

## 前端接口与执行生命周期

`mygo.Bind` 将 Go 服务的导出方法暴露给前端，生成 TypeScript 调用端及类型；方法调用运行在各自 goroutine 中。`context.Context` 可识别调用页面，并在页面导航或窗口关闭时取消；共享数据仍需宿主自行处理并发。[Bindings 文档](https://mygo.egoist.dev/docs/bindings)

`Channel[T]` 提供有序流式结果，方法结束时关闭流；前端停止消费、关闭 Channel 或页面消失可以取消对应调用。源码包含 ACK 背压，未消费量超过约 1 MiB 时生产者等待，但主线程不能阻塞等待；这不是整个应用的内存上限。[channel.go](https://github.com/egoist/mygo/blob/6842ad5e0aa2cd168166e0a4ae12c0cd0f527a3d/channel.go)

**对 Aibo 的推论：**适合把 UI 订阅放在页面 context 下；Agent 执行必须由独立生命周期管理，不能直接把页面调用的 context 传给长期运行的 Agent 进程，否则重载呈现可能同时取消执行。快照、游标、恢复、审批校验和旧呈现实例失效仍需 Aibo 的业务协议。

前端资源可以嵌入 Go 程序，也可以通过 `Protocol.Handle` 的 `http.Handler` 提供运行时资源。内置页面的 origin 在 macOS/Linux 与 Windows 不同；文档提示不要用浏览器本地存储承担重要持久数据。[前端与资源加载](https://mygo.egoist.dev/docs/frontend)

MyGo 的 `Plugin` 是 Go 服务与 JavaScript 包的组合，`Use` 要在 `App.Run` 前注册。它不等同于 Aibo 用户运行时安装、升级、卸载的呈现包机制；后者仍需独立的包格式、兼容与装载实现。[plugin.go](https://github.com/egoist/mygo/blob/6842ad5e0aa2cd168166e0a4ae12c0cd0f527a3d/plugin.go)

## 呈现装载、信任与恢复

MyGo 能加载不同页面，也支持在原生 UI 窗口内创建多个 `WebView`，通过 `Destroy` 销毁；视图未在当前 UI 中展示时只是隐藏，页面仍保留。可以据此研究整套工作台替换，但多视图不是本轮产品必需条件，也不自动意味着每页独占 OS 进程。[WebView 文档](https://mygo.egoist.dev/docs/ui/web-view)、[webview.go](https://github.com/egoist/mygo/blob/6842ad5e0aa2cd168166e0a4ae12c0cd0f527a3d/webview.go)

可信页面可调用已绑定 Go 服务。`PageOptions.Untrusted` 另外提供强制禁用通用 IPC、限制默认导航等机制，源码也有相应测试；但当前 Aibo 已选信任前端包，不必为采用 MyGo 改成敌对插件模型。真正需要保留的是业务入口校验、明确资源授权以及 UI 外的恢复入口。[page.go](https://github.com/egoist/mygo/blob/6842ad5e0aa2cd168166e0a4ae12c0cd0f527a3d/page.go)、[untrusted_test.go](https://github.com/egoist/mygo/blob/6842ad5e0aa2cd168166e0a4ae12c0cd0f527a3d/untrusted_test.go)

固定 main 源码已有 `OnDidFailLoad` 和 `OnRenderProcessGone`。后者三平台均有接线，但事件细节不同：

| 平台 | 已核查的后端接线 | 恢复判断的注意点 |
| --- | --- | --- |
| macOS | `webViewWebContentProcessDidTerminate` → `terminated` | 可作为明确终止证据，不能推导为通用无响应监测 |
| Windows | WebView2 `ProcessFailed` → `browser-exited`、`crashed` 或 `unresponsive` | **不能收到 `OnRenderProcessGone` 就一律自动回退**；无响应也走此事件 |
| Linux | `web-process-terminated` → 崩溃、内存超限或 API 终止 | 仍需核对目标系统行为；Linux 不是本轮 Aibo 必需目标 |

来源：[公开事件接口](https://github.com/egoist/mygo/blob/6842ad5e0aa2cd168166e0a4ae12c0cd0f527a3d/window.go#L1887)、[macOS 接线](https://github.com/egoist/mygo/blob/6842ad5e0aa2cd168166e0a4ae12c0cd0f527a3d/internal/darwin/window.go#L2538)、[Windows 接线](https://github.com/egoist/mygo/blob/6842ad5e0aa2cd168166e0a4ae12c0cd0f527a3d/internal/windows/webview.go#L282)、[Linux 接线](https://github.com/egoist/mygo/blob/6842ad5e0aa2cd168166e0a4ae12c0cd0f527a3d/internal/linux/window.go#L1958)。这里只核查了源码，没有注入崩溃验证。

额外核对发布标签可见，Windows 将无响应并入此事件的行为已存在于 [v0.7.4 源码](https://github.com/egoist/mygo/blob/v0.7.4/internal/windows/webview.go#L279)，并非仅 main 未发布改动。

`Page.ProcessID` 也不是统一的渲染进程标识：macOS 通过私有 WebKit API 读取网页内容进程，Windows 返回多个 WebView 共享的浏览器进程，Linux 返回 0。因此不能直接据该值实现跨平台逐页面故障归属或内存统计。[进程标识说明](https://mygo.egoist.dev/docs/windows#the-pages-process)

**对 Aibo 的推论：**MyGo 已提供恢复所需的部分事件和重建入口；是否满足明确失败自动回退、卡死后原生菜单仍可用、重建后订阅与草稿恢复，需要在 macOS/Windows 上验证。WebView 渲染故障与 Go 宿主进程故障应分别定义，不能把事件接口存在当作完整恢复通过。

## 发布、签名与更新

CLI 文档包含 macOS app/DMG、Windows EXE/安装程序、Linux 归档与 Debian 包；macOS Developer ID 签名与公证、Windows 证书或自定义签名命令均有配置入口。这降低了集成工作量，但仍需开发者提供证书和各平台所需工具。[发行指南](https://mygo.egoist.dev/docs/distribution)

更新机制支持 GitHub Releases 或 HTTPS 服务器，使用嵌入公钥验证更新；差量失败时可改下载完整归档。更新 UI 可以使用官方插件或自行实现。Aibo 的数据迁移、业务恢复与更新后的兼容验证不能由这些机制代替。[自动更新](https://mygo.egoist.dev/docs/updates)

## 发布成熟度与核查版本

研究日期为 2026-11-04。GitHub 仓库创建于 2026-09-26；本次检索到的最新框架 release 是 2026-11-01 发布的 v0.7.4，许可为 MIT。仓库创建时间只能说明公开仓库历史较短，不能据此断言全部研发始于该日。下述成熟度判断是工程评估：可以进入原型候选，但缺少长期生产运行证据，尚不足以直接承诺 Aibo 的稳定交付。[仓库元数据](https://api.github.com/repos/egoist/mygo)、[v0.7.4 release](https://github.com/egoist/mygo/releases/tag/v0.7.4)

官方概览写 Go 1.27+，但已发布 v0.7.4 的 `go.mod` 实际要求 **Go 1.27.3**，原型应按精确要求准备工具链。[v0.7.4 go.mod](https://github.com/egoist/mygo/blob/v0.7.4/go.mod)

v0.7.2 的补丁版本包含默认页面手势行为的 breaking changes，不能按版本号假定升级没有兼容性变化；v0.7.3 和 v0.7.4 还在修复 Windows 原生窗口缩放、首帧和其他交互问题。应锁定框架版本、阅读 release notes，并让升级经过 Aibo 的真实流程回归。[v0.7.2](https://github.com/egoist/mygo/releases/tag/v0.7.2)、[v0.7.3](https://github.com/egoist/mygo/releases/tag/v0.7.3)、[v0.7.4](https://github.com/egoist/mygo/releases/tag/v0.7.4)

本次也检查了 main 的固定提交 `6842ad5e0aa2cd168166e0a4ae12c0cd0f527a3d`（2026-11-03）。该提交的 CI 配置包含 macOS、Linux、Windows 测试与 GUI e2e，另有 Windows ARM64 job；这是自动化覆盖的积极证据，但本次没有重跑上游 CI，也未确认每个 job 的近期执行结果。main 文档与发布标签不应混为一份已交付合同。[固定提交 CI 配置](https://github.com/egoist/mygo/blob/6842ad5e0aa2cd168166e0a4ae12c0cd0f527a3d/.github/workflows/ci.yml)

## 性能数字与构建成本的证据边界

首页列出约 7 MB 二进制、44 MB 原生 UI 内存、62 MB 含 WebView 内存。首页源码说明这些数字来自 **macOS 的示例程序**，且以固定值展示；它们不是 Aibo 的测量，也不是三平台统一保证。[首页指标源码](https://github.com/egoist/mygo/blob/6842ad5e0aa2cd168166e0a4ae12c0cd0f527a3d/website/src/routes/index.tsx#L19)

独立的持续基准有可检查的实现：构建 `examples/hello` 与 `examples/counter-native` 的 release 程序，默认重复六次取中位数，同时运行 Go 与 WebView IPC 等基准。它跟踪 MyGo 自身的变化，没有在该脚本中进行 Electron、Tauri 或 Wails 的同功能对照。[基准脚本](https://github.com/egoist/mygo/blob/6842ad5e0aa2cd168166e0a4ae12c0cd0f527a3d/scripts/bench.ts)、[WebView e2e 基准](https://github.com/egoist/mygo/blob/6842ad5e0aa2cd168166e0a4ae12c0cd0f527a3d/internal/e2e/bench_test.go)

内存采样默认在启动六秒后进行，包含应用与相关 WebView 进程；macOS 用 physical footprint，Linux 用 PSS，Windows 用 private working set。不能把不同系统的值当作同一统计口径直接排名；macOS 的进程归属识别实现也应一并纳入复现记录。[内存采样实现](https://github.com/egoist/mygo/blob/6842ad5e0aa2cd168166e0a4ae12c0cd0f527a3d/internal/idlemem/main.go)、[macOS 归属与 footprint](https://github.com/egoist/mygo/blob/6842ad5e0aa2cd168166e0a4ae12c0cd0f527a3d/internal/idlemem/mem_darwin.go)

**对构建空间的推论：**如果重构后的宿主用 Go 实现且依赖不再要求 Rust 编译，可去掉旧宿主对应的 Cargo 构建链；MyGo 的 no-cgo 设计进一步减少跨平台 C 工具链负担。但 Go 模块/构建缓存、前端依赖、打包工具以及 Agent 自己的运行时仍占空间；保留 Rust sidecar 时也保留相应构建成本。本次未安装 MyGo、未测冷构建/增量构建耗时与磁盘增量，因此不承诺具体节省比例。前次检查旧 Aibo 的 133 GB target 是历史累积占用，不能拿来与 MyGo 的首次空项目构建做公平对比。

## 对 Aibo 的建议

将 MyGo 列入**小体积 Go 宿主的优先原型候选**，采用 Web 前端路线评估。这个建议基于上述构建与接口机制，不代表已经选定迁移：

- 它直接回应减少 Rust 宿主构建负担的诉求；若 Aibo 仍保留 Rust 实现，则要把剩余 Cargo 成本一并计算。
- 继续接受系统 WebView 差异，换取较小分发体积；它没有为 Web 呈现提供 Electron 那样随应用固定 Chromium 的一致性。
- 原生 UI 可作为未来独立工具窗口或恢复入口的研究选项；将主工作台全部改为 Go 自绘，会改变呈现包的实现方式，不能因为有原生 UI 就认为现有跨框架方案可直接复用。
- 官方生成绑定作为桌面适配层使用。Aibo 对呈现包公开的业务协议应自行版本化，不因暴露某个 Go 方法就把宿主内部结构变成长期合同。
- 对当前以完整工作台、共享状态与故障回退为重点的范围，先验证一条真实会话流程，比追求示例的最小二进制更有决策价值。Bun 不随框架分发也不意味着外部 ACP Agent 不再需要它们自己的运行时。

## 建议的验证边界

以下是由上述证据推导的原型验收项，尚未实施：

1. 用同一业务协议接入两种框架的完整关键流程，验证 Go 生成的调用端不会成为包内部 UI 框架耦合。
2. 并发流式更新时替换或重载当前呈现，确认执行继续，快照与增量重新衔接，无重复提交或迟到审批。
3. 分别注入加载失败、页面 JavaScript 死循环、渲染进程退出；记录实际事件，以及原生菜单恢复是否仍可达。Windows 的 `unresponsive` 不自动触发崩溃策略。
4. 在 macOS/Windows 验证中文 IME、焦点/选区、图片剪贴板、拖放、菜单和文件选择。
5. 分别测干净构建、热构建及缓存增长、安装包与按需资源总量、完整进程树内存、长历史输入响应；不能以空窗口指标决定迁移。

研究结论限于文档与源码层面的候选可行性，尚不证明已满足 Aibo 全部平台、性能与恢复承诺。
