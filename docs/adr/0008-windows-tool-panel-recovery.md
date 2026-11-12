---
status: accepted
date: 2026-11-12
revises: ADR 0005 §5 的 Windows 工具面板故障恢复
---

# Windows 工具面板区分 renderer 故障与浏览器进程退出

用户于 2026-11-12 确认“人工验证通过；ADR 0008 恢复策略可以按照建议修订”。本文修订 ADR 0005 §5 的 Windows 恢复合同；macOS 的已验证恢复方式保持有效。

## 实机证据

[17 号票](../../.scratch/aibo-rev-architecture/issues/17-windows-presentation-validation.md)在 Windows 11 x64 26300.9711、MyGO v0.9.0、Go 1.27.3、WebView2 157.0.4341.46 上得到：

- 同一原生窗口里的两个不同包面板使用不同 origin、不同 renderer，共享浏览器进程。
- A 面板死循环 12 秒时，B 面板仍可执行脚本；期间没有观察到 `unresponsive` 回调。这只证明本次观察窗口内未触发，不能认定永不触发。
- 结束 A 的 renderer，收到 `crashed`，B 仍可用，A 经 `Reload()` 恢复。
- 结束共享浏览器进程，两面板均收到 `browser-exited`；原 WebView 上 `Reload()` 后仍返回 HRESULT `0x8007139f`。
- 在原生窗口内销毁并重新创建两个 WebView 后，两面板均恢复，origin 与安全上下文保持正确。
- 无响应面板直接 `Destroy()` 再创建同包 WebView 也可恢复：创建调用约 2 ms，2 秒后页面已可用，原 renderer 不再存活；B 面板始终可用。见 [win-loop-recreate.json](../../prototypes/mygo-validation/reports/win-loop-recreate.json)。这不需要结束共享浏览器或枚举 renderer。

原始报告：[win-two-panels-recreate.json](../../prototypes/mygo-validation/reports/win-two-panels-recreate.json)。

## 决定

1. Windows 上把 `browser-exited` 作为共享故障：标记所有受影响面板不可用。用户选择“重新加载面板”时，销毁并重新创建这些失效的 WebView，再按包与工作区恢复展示状态。保留当前原生窗口、会话执行和包进程。
2. renderer 的崩溃仍先用 `Reload()` 恢复。不得将 `Page.ProcessID()` 当作单个面板 renderer 的 PID 并结束它；该 API 在 Windows 返回共享浏览器 PID。
3. 面板无响应仍用面板心跳判断，不能仅依赖 `unresponsive` 回调。用户选择“重新加载面板”时，直接销毁并重建该 WebView，释放原 renderer；不结束共享浏览器，不枚举进程命令行。renderer 已崩溃时仍可先尝试 Reload。
4. F17.5 的后台工作仍由包进程承担。WebView 重建只恢复展示，不代替用户发送消息或重启 Agent。

## 验收范围与限制

- 用户已接受共享浏览器退出时重建全部受影响面板，以及无响应面板直接重建的策略。直接重建还在 [win-loop-recreate-binding.json](../../prototypes/mygo-validation/reports/win-loop-recreate-binding.json) 中复核了原 renderer 退出、另一面板与 Go 绑定持续可用。
- [17 号票](../../.scratch/aibo-rev-architecture/issues/17-windows-presentation-validation.md#answer)以自动报告和用户人工验收收口，范围为本次 Windows 11 x64 组合。
- 这些原型证明恢复方式可行；产品实现仍须按本合同接线，并在 12 号票定义的端到端验收中验证。重建不得改变授权、触发 Agent 执行或丢失宿主持有的业务状态。
