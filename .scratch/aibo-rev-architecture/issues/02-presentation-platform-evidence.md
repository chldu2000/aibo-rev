# 跨框架呈现与桌面容器的可行边界

Status: resolved
Type: research
Labels: wayfinder:research
Parent: ../map.md
Assignee: chldu2000 (Codex research agent)
Blocked by: none

## Question

哪些浏览器标准接口与桌面容器机制，能让默认 UI 和外部呈现包独立选择框架，同时满足隔离、卸载、故障回退与 macOS/Windows 交付？

调查结构化消息、独立 document/iframe/WebView、DOM/Web Components 等接口的适用范围，以及 Tauri 和 Electron 对这些机制的约束。核对跨容器输入焦点、弹层、剪贴板/文件、资源加载、渲染进程故障和高频状态传递的限制。不要把 Shadow DOM 视为安全或崩溃隔离，也不要把无框架数据模型视为已证明整个工作台可替换。

输出事实与待原型验证的问题，不替用户选择容器、语言、隔离方式或默认 UI 框架。平台能力不等于框架中立已经实现。

范围依据：F20.1、F20.3–F20.7、F21、F25；框架中立须用完整替换及跨框架呈现包验证。

研究资产：`docs/research/framework-neutral-presentation.md`；上下文：`research/framework-neutral-presentation` 分支。

## Comments

2026-11-01：已由 chldu2000 的 Codex 研究会话认领。从 `86b8efd` 建立独立分支，工作树为 `/private/tmp/aibo-rev-ui-research.ldtMBb`。本票不选择技术栈；root 汇总研究资产并记录 Answer，呈现接口与选型留给后续讨论。

## Answer

2026-11-01：完成浏览器标准与桌面容器接口核查，见 [框架中立呈现研究](../../../docs/research/framework-neutral-presentation.md)。研究分支 `research/framework-neutral-presentation`，最终提交 `8d57aba0d01a5291caaf2c05b55466de1e9e76c7`，同版文档已收录到主工作区。

- 可序列化数据、命令与事件可以跨框架；工作台交互状态和完整替换合同仍需应用自行定义。
- 同文档组件、iframe 和原生多视图提供不同程度的封装；框架独立、权限隔离和进程故障隔离不能互相代替。
- 核查的 Tauri 2.12.5 子 WebView API 仍需 unstable，相关进程终止回调不是双平台统一接口；Electron 有 renderer 故障通知，但视图可能共享进程，子视图资源也需要显式释放。
- 由官方接口不能推断中文输入、焦点、弹层、大历史流式更新及回退均通过验收。报告列出了可执行的最小验证场景。

未选择桌面容器、语言或 UI 框架，也未运行原型。证据进入 [框架中立的工作台呈现接口](05-presentation-interface.md)；原先模糊的验证方向已明确为 [跨框架替换与呈现故障回退的最小原型](13-presentation-conformance-prototype.md)，后者将为最终选型提供人工评审依据。
