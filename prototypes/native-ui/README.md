# 原生 UI 原型：会话流与输入区（MyGO native UI）

一次性实验，不是产品实现。目的：评估把默认界面从 WebView（Svelte + Bits UI + virtua）换成 MyGO 原生 UI 是否可行，对照 [`stream-composer`](../stream-composer/README.md) 与 [界面硬性要求](../../docs/design/ui-requirements.md)。

## 运行

```sh
go generate                                   # 经 tools/ui-checks 从 docs/design/tokens/tokens.json 生成 tokens_gen.go
CGO_ENABLED=0 go test ./...                   # 无窗口交互检查（ui.Tester），含 tools/ui-checks 验收套件（TestAcceptance）
SHOTS=1 CGO_ENABLED=0 go test -run TestShots  # 深浅色截图到 reports/
CGO_ENABLED=0 go build -o bin/native . && ./bin/native
MYGO_FRAME_STATS=all ./bin/native -auto reports/auto.json > reports/frames.log 2>&1   # 真实窗口自动测量
```

| 文件 | 内容 |
| --- | --- |
| `tokens_gen.go` | 生成的设计令牌：`Palette`（深浅两套）与字号、间距、圆角、尺寸常量；生成器在 `tools/ui-checks` |
| `model.go` | 会话模型：10,000 条历史（尾部 200 条先载入，向上滚动时每次前插 80 条）、模拟流式回复、审批、队列、候选 |
| `view.go` | 界面：会话列表、`ui.List` 会话流、`ui/markdown` 流式回复、工具卡、审批卡、结果未知卡、状态条与队列浮层、输入区 |
| `view_test.go` | 无窗口交互检查 |
| `acceptance_test.go` | 作为 `tools/ui-checks` 的验收目标，提供 `idle-empty` 与 `history-10k` 夹具 |
| `shots_test.go` | 截图 |

## 样式怎么落地

- **令牌**：`tokens.json` 仍是唯一来源，生成 Go 代码。能对上的映射到 `ui.Theme`（背景、表面、边框、文字、强调色、状态色、选中、焦点、提示框、圆角、字号、间距单位），其余（关注/危险/未知底色、气泡、代码背景等）直接从 `Palette` 取。深浅色由 `mygo.Theme.IsDark()` 与 `OnUpdated` 切换 `Palette`。
- **无样式基件**：会话列表行、候选行、标签移除按钮、队列标签用 `ButtonBase`；输入框用 `TextAreaBase`；队列浮层用 `PopoverBase`；候选弹层用 `Overlay` + `AttachTo`（不夺焦点）。
- **规范对照**：
  - 审批卡与“结果未知”卡用整圈边框 + 浅底（后者虚线），不用左侧色条（§2.4）。
  - `ui/markdown` 的引用块默认带 3 DIP 左侧竖条（`QuoteBarWidth`），原型设为 0。
  - 禁用的发送按钮用 `AccessDisabled` + 静态底色/文字 + `Tooltip` + `Description` 说明原因，不用半透明。

## 结果（2026-11-11，macOS 27 arm64，MyGO v0.9.0）

### 无窗口交互检查（`go test`，全部通过）

| 检查 | 结果 |
| --- | --- |
| Enter 发送、Shift+Enter 换行 | 通过 |
| 输入法组字（`tt.Compose`）时 Enter 不发送，提交后再 Enter 发送 | 通过 |
| `@src/co` 弹出候选，焦点留在输入框，↓ + Enter 选中变为标签并删去触发文字，不发送 | 通过 |
| `#` 候选 Esc 关闭、草稿不变，继续输入后重新出现 | 通过 |
| 执行中 Enter 入队；队列浮层打开、Esc 关闭、焦点回到“队列 1” | 通过 |
| 审批卡内联、允许后状态条提示消失 | 通过 |
| 设置应用中 Enter 不发送、草稿保留；读屏看到“发送”为禁用且有原因 | 通过 |
| 所有按钮有可读名称 | 通过 |
| 10,000 条历史只构建可见行 | 7 条用户消息被构建 |
| 流式 108 帧（含 CPU 绘制） | 平均 2.45 ms，最长 3.9 ms |
| 流式期间输入一个字 | 2.5 ms |

### 真实窗口（Metal，`MYGO_FRAME_STATS=all`）

| 项目 | 结果 |
| --- | --- |
| 401 帧（空闲、流式长回复、载入全部 10,000 条并跳到中部） | 中位 0.9 ms，p95 2.2 ms；仅首帧 35.9 ms、跳到中部那一帧 11.6 ms 超过 8 ms |
| 流式：发布 → 下一次构建 | 平均 5.4 ms，最长 8.2 ms（WebView 原型为“发布→下一帧”平均 9 ms、最长 18 ms，口径不完全相同） |
| 视图构建 | 平均 0.13 ms，最长 0.58 ms |
| 进程内存（启动 5 秒后 RSS） | 120 MB（单进程；WebView 方案另有 WebContent 等进程，未在此对比） |
| 可执行文件 | 15.7 MB，无 cgo |

### 截图（`reports/shot-light.png`、`shot-dark.png`）

深浅两套令牌生效；中文与代码字体、行内代码底色、代码块、虚线边框、弹层阴影与圆角正常。发现：`ButtonBase` 内容默认居中，列表行需 `Justify(ui.Start)`。

## 手动验证（2026-11-11，用户）

以下项目用户确认无问题：中文输入法（候选确认的 Enter 不发送、候选窗位置）、`@` 候选的键盘与点击选择、⌘V 粘贴截图成为图片标签、拖入文件时边框高亮并成为标签、跨消息选中复制、流式长回复期间向上滚动（停止跟随、出现“有新内容”）、随系统切换深浅色。

用户意见：原生组件（按钮、下拉选择等）的外观之后需要单独调整，本原型只用了默认组件加令牌颜色。

## 未验证

- ~~Windows（Direct3D 11）上的绘制、字体回退与输入法~~ 2026-11-12 [17 号票](../../.scratch/aibo-rev-architecture/issues/17-windows-presentation-validation.md#answer)已完成 Windows 11 x64 自动实验与用户人工确认，报告见 `reports/win-auto.json`、`win-memory.json` 及[汇总](../workbench-shell/reports/win-validation-progress.md)。
- 工具面板：原生窗口中嵌入能力包 WebView 已在 [`mygo-validation`](../mygo-validation/README.md) 验证可行（独立进程、可调用 Go），未与本原型合并。
- 屏幕阅读器的实际朗读（测试器只检查了无障碍树）。

## 2026-11-11 用户决定

默认界面改用原生 UI（[ADR 0006](../../docs/adr/0006-native-ui.md)）。本原型同时作为 `tools/ui-checks` 的第一个验收目标：为契约补上了“用户消息”“Agent 回复”两个无障碍分组；弹层阴影颜色暂以注明例外的变量保留，待令牌生成器支持阴影。
