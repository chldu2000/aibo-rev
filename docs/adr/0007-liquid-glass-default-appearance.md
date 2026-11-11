---
status: accepted
date: 2026-11-12
---

# 默认外观采用 Liquid Glass

## 决定

新 Aibo 默认界面的外观风格定为 Liquid Glass（用户 2026-11-12）。它取代 2026-11-04 起“暂不指定设计系统或呈现风格”的状态。用 MyGO 的窗口材质与 `plugins/glass` 实现，不另外引入设计系统或组件库。

玻璃只用于外框层，内容层保持不透明：

| 区域 | 做法 |
| --- | --- |
| 侧栏 | 窗口材质 `VibrancySidebar`，侧栏与根元素透明。系统材质的模糊程度不能调整，保持系统默认 |
| 会话区顶部 | 会话流延伸到标题行下方，标题行下用磨砂滚动边缘 `glass.ScrollEdge{Hard: true}`。渐隐的 Soft 边缘会让文字与标题重叠，不采用 |
| 输入区 | 状态条与输入框放在同一块浮动玻璃 `glass.Glass{}` 上（圆角令牌 `radius.glass` = 22px），下方内容用底部滚动边缘渐隐 |
| 内容层 | 消息、工具卡、审批卡、结果未知卡、代码块等保持不透明，沿用现有令牌 |
| 工具面板 | 面板区不透明；面板页面不受本决定约束（ADR 0005 第 5 节） |

## 依据

- 原型 [`workbench-shell`](../../prototypes/workbench-shell/README.md)：用户看过真实窗口后确认侧栏与浮动输入区的样式。
- 性能：Metal 上流式输出的帧耗时中位数由 1.3 ms 增至 1.7 ms，滚动由 2.5 ms 增至 2.8 ms，p95 都在 4 ms 以内。
- 可访问性：开启“减少透明度”时，玻璃改为不透明的 `Surface`；开启“增强对比度”时，玻璃加亮边缘并为不透明玻璃加边框（MyGO 内置）。

## 后果

- 按钮、输入框、选择器、状态标签与徽记等组件按 Liquid Glass 风格设计：2026-11-12 用户选定方案 A“玻璃胶囊”，状态标签悬浮在输入面板之外，弹出层统一为玻璃面板（界面方案 §2.5，UI-VIS-11、UI-VIS-12）。
- 内容层不透明，所以 UI-VIS-02 的对比度检查仍然有效；玻璃上的文字对比度随下方内容变化，由 UI-VIS-09 约束，并在截图中人工核对。
- `plugins/glass` 用自己的着色器绘制，macOS 与 Windows 外观一致；窗口材质在 Windows 11 上，`VibrancySidebar` 对应 Mica。没有 GPU 时退回 CPU 绘制，玻璃下内容滚动的一帧可达 30–70 ms（MyGO 文档）。

## 尚未验证

- Windows 上侧栏的 Mica 效果、`plugins/glass` 在 Direct3D 11 上的绘制与帧耗时（[17 号票](../../.scratch/aibo-rev-architecture/issues/17-windows-presentation-validation.md)）。
- 没有 GPU 或 GPU 不可用时 CPU 绘制的实际帧耗时。
