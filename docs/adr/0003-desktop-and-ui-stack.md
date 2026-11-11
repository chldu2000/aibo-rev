---
status: accepted
date: 2026-11-10
---

# 桌面容器、宿主语言与默认界面技术栈

用户于 2026-11-10 决定新 Aibo 采用以下技术栈：

| 层 | 选择 |
| --- | --- |
| 桌面容器 | [MyGO](https://github.com/egoist/mygo)，走系统 WebView 路线：macOS WKWebView、Windows WebView2 |
| 宿主（运行核心与桌面宿主） | Go |
| 默认呈现（F20.1） | Svelte 5 + Bits UI + virtua；流式 Markdown 用 markdown-it，代码高亮用 Shiki |

这项决定取代此前“正式技术栈留待单独设计”的状态，回答 [11 号票](../../.scratch/aibo-rev-architecture/issues/11-module-and-desktop-architecture.md)中的桌面容器、宿主语言与默认界面组件栈三项；该票其余问题仍然开放（见下文）。

## 依据

- **MyGO + WebView。** 外部呈现包须能用任意前端框架交付完整工作台（[05 号决议](../../.scratch/aibo-rev-architecture/issues/05-presentation-interface.md#answer)），只有 WebView 路线能承接；MyGO 的 Go 原生 UI 只用于呈现失效时的恢复入口等少量宿主窗口。调研见 [MyGO 桌面方案调研](../research/mygo-desktop-evaluation.md)。
- **Go 宿主。** 与 MyGO 同一语言，去掉旧宿主的 Rust/Cargo 构建链，回应降低维护成本的主目标（[ADR 0002](0002-aibo-refactor-direction.md)）。
- **默认呈现组件栈。** 无样式的 Bits UI 便于实现自有视觉规范；virtua 承担长会话虚拟化与向前加载。会话流与输入区已在 Chromium 与 WebKit 上验证，见 [验证记录](../design/ui-interaction.md#151-验证记录会话流与输入区2026-11-10) 与 [`prototypes/stream-composer`](../../prototypes/stream-composer/README.md)。

## 约束

1. **框架中立不变。** Svelte 只用于默认呈现包。运行核心、共享无界面交互模块与公开呈现接口不得依赖 Svelte 的组件、响应式对象或生命周期；外部呈现包仍可使用任意框架（05 号决议）。默认呈现与外部包使用同一公开接口。
2. **重计算离开主线程。** 代码高亮等在 Web Worker 中运行；MyGO 的资源协议须支持模块 Worker（验证记录发现 1）。
3. **锁定版本。** MyGO 处于早期（调研时 v0.7.4，补丁版本含 breaking change），锁定框架版本及其要求的 Go 工具链版本（v0.7.4 为 Go 1.27.3），升级须经过真实流程回归。优先保持无 cgo；引入必须使用 cgo 的依赖前单独评估跨平台构建成本。
4. **界面硬性要求。** 默认呈现须满足 [界面硬性要求](../design/ui-requirements.md)，并通过 `node tools/ui-checks/run.mjs all`。

## 尚未验证

这些是选型后的必做验证，不是撤回选型的预设条件；结果若推翻某项选择，须新开 ADR 取代本文：

- 在 MyGO 的 WKWebView 与 WebView2 中运行默认呈现：中文 IME、剪贴板图片、拖放、模块 Worker、长会话性能（[界面方案 §15](../design/ui-interaction.md#15-需要在-mygo-上验证的交互)）。
- Windows 上的呈现故障与恢复入口（[17 号票](../../.scratch/aibo-rev-architecture/issues/17-windows-presentation-validation.md)）。

## 后续决定

11 号票的其余问题已于 2026-11-11 决定：共享交互状态与公开呈现协议见 [ADR 0004](0004-presentation-protocol.md)；进程部署、存储与索引、ACP 接入实现、包交付与工具面板装载见 [ADR 0005](0005-runtime-storage-and-packages.md)。
