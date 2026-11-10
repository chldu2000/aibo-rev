# 会话流与输入区原型（Svelte 5 + Bits UI + virtua）

一次性验证，不是产品实现。目的：在[默认工作台界面方案](../../docs/design/ui-interaction.md)确定后，验证候选组件栈能否承担会话流和输入区里最难的部分。

## 运行

```bash
pnpm install
pnpm build && pnpm exec vite preview --host 127.0.0.1 --port 5199   # 或 pnpm dev
node scripts/webkit-check.mjs http://127.0.0.1:5199/                  # WebKit 自动检查（需 playwright-core 对应的 WebKit）
```

右上角“验证面板”可以载入全部历史、开始长回复、调整发布间隔与流速，并显示帧率、慢帧、发布→下一帧延迟、挂载行数与 IME 事件日志。

## 结构

| 路径 | 职责 |
| --- | --- |
| `src/core/session.ts` | 与框架无关的会话模型：条目、队列、草稿/附件/引用、展开状态、执行设置；`subscribe/getSnapshot` 发布，流式正文按间隔合并发布 |
| `src/core/markdown.ts` | 流式 Markdown：按空行和代码围栏切块，已渲染块按文本缓存 |
| `src/core/highlight*.ts` | Shiki（JS 正则引擎）在 Web Worker 中高亮，主线程只缓存结果 |
| `src/lib/store.svelte.ts` | Svelte 适配层，仅用 `createSubscriber` 接入订阅 |
| `src/lib/Stream.svelte` | virtua `VList`：底部跟随、向上加载更早记录（`shift`）、“有新内容”提示 |
| `src/lib/Composer.svelte` | 输入区：IME 安全的 Enter、`@ # /` 候选（Bits UI Popover + `customAnchor`）、引用菜单（DropdownMenu）、队列浮层（Popover）、模型/权限选择（Select）、禁用原因（Tooltip） |

## 结果（2026-11-10，macOS arm64）

Playwright Chromium 与 WebKit 27.2（Playwright 1.65 构建，与 WKWebView 同引擎但不等于 MyGO 内的 WKWebView）。会话含 10,003 条历史，长回复约 1,000 字、含 TS/Go/Bash 三个代码块，发布间隔 50ms、流速 6 字/16ms。

| 项目 | Chromium | WebKit |
| --- | --- | --- |
| 10,000 条历史时挂载的行 / 会话流 DOM 节点 | 8–10 / 60–170 | 9 / 91 |
| 流式期间帧率 | 80–120 fps | 60 fps |
| 流式期间 >50ms 的帧 | 2 | 0 |
| 发布→下一帧（平均 / 最大） | 12 / 88 ms（高亮移入 Worker 前） | 9 / 16 ms |
| 跟随底部 | 是 | 是 |
| 向前加载 80 条后可见条目的位移 | 0 px（5 次） | 0 px（3 次） |
| 代码块高亮 | 3/3 | 3/3 |
| `@` 候选弹层 / 焦点留在输入框 / 键盘选择 | 通过 | 通过 |
| 执行中 Enter 入队、队列浮层、Esc 收起 | 通过 | 未单独测 |
| 引用菜单（单选范围、禁用项、关闭后焦点回到触发器） | 通过 | 弹出通过 |
| 设置应用中阻止发送、确认后恢复 | 通过 | 通过 |

### 发现

1. **代码高亮必须离开主线程。** 首个 TS 代码块闭合时，Shiki 在 JavaScriptCore 上首次编译语法，阻塞主线程 175ms；用 `'x'` 预热无效（语法规则按命中惰性编译）。移入 Web Worker 后最大延迟降到 16ms。Chromium 上同一问题约 88ms。
2. **“跟随底部”不能只看滚轮。** 拖动滚动条或程序定位不产生 wheel 事件；第一版因此在加载更早记录时被拉回底部。改为每次 scroll 事件按“是否仍在底部”重新判定（流式增长本身不触发 scroll 事件）。
3. **virtua 的 `shift` 能正确处理向前加载。** 每次前插 80 条，可见条目位移为 0。
4. **Bits UI 的 Popover `customAnchor` 能把候选列表锚定到输入框，同时保持文本框焦点**（`onOpenAutoFocus` 阻止默认聚焦、`trapFocus=false`、对输入框的点击不当作外部点击）。
5. **Bits UI Select 在 `pointerdown` 打开。** 用合成 `click()` 的测试脚本会误判；产品行为正常。
6. **禁用但需要解释原因的按钮用 `aria-disabled`**，保留焦点与提示，不用 `disabled`。
7. 主包 360 kB（gzip 122 kB），Shiki 与语法在 Worker 包中按需加载。

### 未解决 / 未验证

- WebKit 下出现一次 `ResizeObserver loop completed with undelivered notifications`，Chromium 无此提示；不影响行为，来源未定位。
- 未在 MyGO 内运行：本机 Go 为 1.24.1，MyGO v0.7.4 需要 Go 1.27.3，且未安装 MyGO CLI。
- 中文 IME：自动化无法模拟真实输入法。验证面板会记录 `composition*` 与 Enter 的 `isComposing/keyCode`，需要在 MyGO（WKWebView、WebView2）中手动输入确认。
- 未测 Windows / WebView2、剪贴板图片粘贴、拖放、屏幕阅读器。
- `markdown.ts` 按空行切块，会把松散列表拆成多段；正式实现应按 markdown-it token 切块。
