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
