# 桌面候选在 Windows 的关键验证

Status: open
Type: prototype
Labels: wayfinder:prototype
Parent: ../map.md
Assignee: unassigned
Blocked by: 11, 13

## Question

前序原型与模块架构讨论采用的桌面候选，能否在 Windows 上满足完整呈现替换、输入与原生交互、明确失败自动回退、页面卡死时独立手动恢复的共同合同？是否需要修正候选或接口才能完成双平台选型？

Q71 已允许分平台取得证据并带条件继续讨论，本票补齐架构收口前所需的 Windows 关键证据。它不重新决定功能范围或故障策略，不以 macOS 结果推定 Windows 行为。

使用[跨框架替换与呈现故障回退的最小原型](13-presentation-conformance-prototype.md)的实验资产，依据[模块接口、技术栈与桌面部署形态](11-module-and-desktop-architecture.md)的候选结论选择实际运行组合。若后者更换容器，本票按新候选验证，不保留已失效的 Electron 假设。

必须记录：

- Windows、容器、引擎、运行时及两种 UI 框架的版本，实际运行方式和测量口径。
- 完整关键流程替换、草稿/引用/审批/阅读状态恢复、未知投递不自动重发，以及旧实例失效。
- 实际加载失败、渲染进程退出、页面死循环与原生菜单/快捷键恢复；单纯执行菜单处理函数不能替代实际入口可达性的验证。
- 中文 IME、候选确认与提交键、焦点/选区、图片剪贴板、拖放、原生文件/目录选择器。
- 长历史及并发样本的输入表现、消息积压、重同步、进程归属和资源释放；不将不同系统内存统计直接作等价排名。
- 用户操作反馈，以及失败后需要修正的合同或明确保留的限制。

Agent 与存储可继续使用前序模拟模型；真实 ACP 消费、停止、持久性和完整产品覆盖仍由最终验收规格承接。实验代码留在可丢弃分支，结果与版本固定到本票。

## 执行清单

2026-11-12 整理，尚未执行。范围：用户 2026-11-12 决定只验证 Windows 11 x64，不覆盖 Windows 10 与 ARM64。取代上文 Question 与 Comments 中已取消的项目（完整呈现替换、两种 UI 框架、自动回退、页面卡死时的原生菜单恢复）。开始前先按 tracker 约定 claim 本票。

### 0. 出发前在 macOS 上完成的准备

到 Windows 机器上应只需运行，不再改代码。

- [ ] **P1 子进程与 Job Object 实验**：`mygo-validation` 不能在 Windows 编译（`auto.go`、`webviews.go` 中的 `syscall.Kill`），`-children` 用 `/bin/cat`、`/bin/sh`，`crash-children.sh` 依赖 POSIX，仓库中还没有 Job Object 代码。需要补齐：结束进程改为跨平台写法；Windows 版子进程实验（读 stdin 的子进程、不读 stdin 的子进程、子进程派生的孙进程），放入带 `JOB_OBJECT_LIMIT_KILL_ON_JOB_CLOSE` 的 Job Object；PowerShell 版崩溃脚本（`Stop-Process -Force` 后检查各进程是否残留）。
- [ ] **P2 面板渲染进程的结束方式**：MyGO v0.9.0 在 Windows 上的 `Page.ProcessID()` 是 WebView2 共享的浏览器进程（`get_BrowserProcessId`），没有提供单个渲染进程号。`workbench-shell` 的 `killPanel` 与 `auto.go` 第 6 步在 Windows 上会结束全部 WebView，和 macOS 的测法不同。改为分别测量两种情况：(a) 结束浏览器进程：记录同时打开两个面板时的影响范围与恢复耗时；(b) 实验手段：在浏览器进程的子进程中按命令行 `--type=renderer` 找到面板的渲染进程，只结束它。
- [ ] **P3 面板无响应信号**：MyGO 把 WebView2 `ProcessFailed` 的 unresponsive 映射为 `OnRenderProcessGone("unresponsive")`。在 `auto.go` 中记录面板死循环后这个回调是否触发、耗时多久，作为 Windows 上自动发现卡死的依据。
- [ ] **P4 记录 origin 与安全上下文**：MyGO 在 Windows 上把自定义协议映射为 `http://<scheme>.localhost` 与 `http://*.<scheme>.localhost`。实验要记录面板页面的 `location.origin` 与 `isSecureContext`，并比较两个包的 origin 是否不同。
- [ ] **P5 构建产物**：重新交叉编译 `workbench-shell`（`go tool mygo build -platform windows/amd64 -skip-installer`）、`native-ui` 与 `mygo-validation`。另外再构建一个带安装程序的版本：未打包的 exe 可能无法发送或点击 toast 通知（需要 AppUserModelID）。

### 1. 环境记录

- [ ] Windows 11 版本与内部版本号（x64）、GPU 与驱动、显示缩放（至少 100% 与 150%，如有可能接一台缩放不同的第二显示器）。
- [ ] WebView2 Runtime 版本、Go 1.27.3、MyGO v0.9.0、各原型的提交号。
- [ ] 输入法（微软拼音，加一款第三方输入法）、读屏（讲述人，可选 NVDA 及其版本）。

### 2. 自动检查（PowerShell）

```powershell
cd prototypes\workbench-shell
$env:CGO_ENABLED=0; go test ./...                         # 无窗口检查也在 Windows 上跑一次
.\bin\shell.exe -auto reports\win-auto-destroy.json
.\bin\shell.exe -close hide -auto reports\win-auto-hide.json
$env:MYGO_FRAME_STATS='all'
.\bin\shell.exe -perf reports\win-perf-standard.done
.\bin\shell.exe -glass -perf reports\win-perf-glass.done
cd ..\native-ui; go test ./...; .\bin\native.exe -auto reports\win-auto.json
cd ..\mygo-validation; .\scripts\crash-children.ps1        # P1 产物
```

通过口径与 macOS 结果对照（`workbench-shell`、`native-ui` README）：帧耗时中位与 p95 同一量级；面板死循环期间工作台界面线程延迟没有明显上升；结束进程后能用 `Reload` 恢复；关窗再打开后，宿主持有的状态（所选会话、草稿、面板宽度、滚动位置）都能恢复。

### 3. 手动检查

**原生界面（D3D11）**
- [ ] 中文、英文、emoji 混排时字体回退正确，没有方框；150% 缩放下文字清晰；窗口在不同缩放的显示器之间拖动后重新布局。
- [ ] 微软拼音和第三方输入法：组字时按 Enter 不发送；候选窗跟随光标；Shift 切换中英；确认候选后再按 Enter 才发送。
- [ ] 用 Win+Shift+S 截图后粘贴；从资源管理器拖入图片、文件和文件夹。
- [ ] “添加工作区”打开系统目录选择器；附件使用文件选择器。
- [ ] 在运行中切换系统深浅色，界面随之更新。

**标题栏与窗口**
- [ ] 窗口按钮在右上角，不压住标题行内容；标题行里的按钮可以点击，不会触发拖动。
- [ ] 能拖动窗口；双击最大化；Win+方向键贴靠；悬停最大化按钮时出现贴靠布局（非客户区命中测试）。

**工具面板（WebView2）**
- [ ] 用任务管理器或 Process Explorer 记录面板所在的 `msedgewebview2.exe` 进程树。
- [ ] 焦点往返：从输入框按 Tab 进入面板再移出；面板有焦点时 Ctrl+\ 能收起面板。
- [ ] 面板内中文输入法正常，确认候选的 Enter 不触发其他操作。
- [ ] 面板死循环时工作台仍能输入和滚动；“重新载入”能恢复（按 P2、P3 的口径记录）。
- [ ] 收起面板 3 秒，以及隐藏窗口期间，记录页面计时器的走动次数（节流程度，对应 F17.5）。
- [ ] 两个包的 origin 不同（P4）。

**后台、托盘与通知**
- [ ] 关闭最后一个窗口后应用留在后台，窗口被销毁；从托盘重新打开，宿主状态能恢复。
- [ ] 托盘菜单在工作台卡死时仍可用。
- [ ] toast 通知：点击后打开窗口并定位到对应会话。若未打包的 exe 不行，就用安装版再测，并记录原因。

**进程**
- [ ] 用 `taskkill /F` 结束应用后，子进程和孙进程全部结束（Job Object）；正常退出时也没有残留。

**读屏（讲述人）**
- [ ] 按 macOS 手动清单第 6 项逐条朗读：侧栏会话、“用户消息 / Agent 回复”分组、输入框、禁用的“发送”及原因、面板网页内容、分隔条。

**Liquid Glass（ADR 0007）**
- [ ] 侧栏的 Mica 效果。
- [ ] 关闭“设置 → 个性化 → 颜色 → 透明效果”后，界面回退为不透明。
- [ ] 若有不带 GPU 加速的虚拟机，测量 CPU 绘制的帧耗时，顺便关闭 ADR 0007 的“无 GPU”一项。

### 4. 结果回写

- [ ] 报告放在各原型的 `reports/win-*`。结论写入本票 `## Answer`，并在地图 Decisions-so-far 中加一行。
- [ ] 回写对应位置：ADR 0003、0005、0006、0007 的“尚未验证”一节，以及[界面方案 §15](../../../docs/design/ui-interaction.md#15-需要在-mygo-上验证的交互) 各项的 Windows 状态。
- [ ] 任何一项推翻已接受的选择时，先停下来报告并提出新 ADR，不要自行绕过。

## Comments

2026-11-04，根据用户 Q71 及首轮 macOS 原型新建。当前环境未提供 Windows 实测证据，本票未认领、未运行；它阻止最终架构收口，不阻止前序有条件的候选讨论。

首轮可复用资产在 `prototype/presentation-conformance` 分支，提交 `83ac7e8e6b4f0020aff88d74e76a6e8db90d1ba1`，入口为 `prototypes/presentation/README.md`。Windows 还须检查最后窗口关闭后的运行与恢复入口可达性，不能沿用 macOS 应用菜单/Dock 的行为假设。

同日，用户已接受前序原型基本通过，正式技术栈与 UI/交互后续单独设计。本票保持 open/unassigned，待正式候选明确后再推进；不阻塞本次主架构文档提交，也不把用户的原型接受解释为 Windows 已通过。

2026-11-11：[11 号票](11-module-and-desktop-architecture.md#answer)已解决，候选为 MyGO + WebView2 + Go 宿主（[ADR 0003](../../../docs/adr/0003-desktop-and-ui-stack.md)），本票不再按 Electron 假设验证。除上文各项外，须一并验证 [ADR 0005](../../../docs/adr/0005-runtime-storage-and-packages.md#尚未验证) 中与 Windows 相关的项目：WebView2 中跨 origin iframe 的进程归属与死循环影响、自定义协议的每包独立 origin、`unresponsive` 事件与心跳中断的原生提示、菜单与托盘恢复入口在页面卡死时可达，以及 Job Object 在应用崩溃时结束全部子进程。

2026-11-11（ADR 0006）：默认界面改用 MyGO 原生 UI，外部呈现包与全局恢复入口取消。本票在 Windows 上要验证的内容改为：

- 原生界面：Direct3D 11 绘制与帧耗时、中文字体回退、中文 IME（候选确认与提交键、候选窗位置）、图片剪贴板、拖放、原生文件/目录选择器、深浅色切换；`prototypes/native-ui` 可直接复用。
- 工具面板：原生窗口中嵌入的 WebView2 是否在独立进程中运行、面板死循环是否影响工作台、结束渲染进程后重新载入能否恢复、每包独立 origin。
- 进程：Job Object 在应用崩溃时结束全部子进程；最后一个窗口关闭后托盘重新打开工作台。

上文中“完整呈现替换”“两种 UI 框架”“明确失败自动回退”“页面卡死时的原生菜单恢复”等项不再需要验证。

2026-11-12：macOS 上的对应项已由[原生工作台外壳在 macOS 上的组合验证](18-macos-workbench-shell-validation.md)完成，`prototypes/workbench-shell` 可直接在 Windows 上复用（`-auto` 自动检查）。Windows 上补充核对：WebView2 面板收起与窗口隐藏时的计时器节流程度；“结束面板渲染进程”须改用 WebView2 的方式（`ProcessID` 为共享浏览器进程）；关闭最后一个窗口时销毁窗口、托盘重新打开后的状态恢复；通知点击定位（通知区域气泡）。

2026-11-12：默认外观改为 Liquid Glass（[ADR 0007](../../../docs/adr/0007-liquid-glass-default-appearance.md)）。Windows 上另需核对：侧栏窗口材质（`VibrancySidebar` 在 Windows 11 上为 Mica）、`plugins/glass` 在 Direct3D 11 上的绘制与帧耗时（`workbench-shell -glass -perf`），以及“透明效果”关闭时的不透明回退。
