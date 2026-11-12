# MyGO 实机验证（macOS）

一次性实验，不是产品实现。验证 [ADR 0003](../../docs/adr/0003-desktop-and-ui-stack.md#尚未验证) 与 [ADR 0005](../../docs/adr/0005-runtime-storage-and-packages.md#尚未验证) 中可在 macOS 上取得证据的项目。

## 运行

```sh
./scripts/build.sh                                   # 构建默认界面原型并嵌入，产出 bin/validate（CGO_ENABLED=0）
./bin/validate -auto reports/auto.json -log reports/auto.log   # 自动检查，约 2 分钟后退出
./scripts/crash-children.sh                          # 应用被 kill -9 后子进程的去向
./scripts/single-instance.sh                         # 单实例
./bin/validate -log reports/manual.log               # 手动验证：菜单“恢复”与托盘中有测试入口，实验页在“打开实验页”
(cd sqlite-check && CGO_ENABLED=0 go run .)          # 纯 Go SQLite、WAL 批量提交与 FTS5 trigram
```

| 路径 | 内容 |
| --- | --- |
| `main.go` | 宿主：绑定 `Lab` 服务（Channel 推送、回显、心跳、日志）、`aibopkg` 自定义协议（每个包一个 host）、菜单与托盘恢复入口、心跳中断提示、单实例、后台运行、子进程 |
| `auto.go` | 自动检查：默认界面原型、实验页、页面卡死与恢复、渲染进程结束、独立窗口 |
| `webviews.go` | 同一原生窗口中两个 WebView 的进程归属与故障隔离 |
| `web/lab/` | 实验页：Channel、iframe 面板、IME、粘贴与拖放 |
| `panels/panel.html` | 模拟工具面板 |
| `sqlite-check/` | SQLite 独立实验（单独模块） |

## 环境

macOS 27.0（Darwin 27.0.0，arm64），Go 1.27.3，MyGO v0.9.0（2026-11-11 发布；调研时为 v0.7.4），系统 WKWebView。文档站已迁至 mygo.run。本机 Command Line Tools 与 macOS 27 SDK 不匹配，开 cgo 时外部链接失败；MyGO 本身要求 `CGO_ENABLED=0`，按此构建。

## 结果（2026-11-11）

### 默认界面原型（Svelte 5 + Bits UI + virtua + Shiki Worker）

在 MyGO 的 WKWebView 中，经 `mygo://localhost` 加载嵌入资源：

| 项目 | 结果 |
| --- | --- |
| Shiki 模块 Worker | 可用，3/3 代码块高亮 |
| 流式期间帧率 / 发布→下一帧 | 60 fps；平均 9 ms、最大 18 ms；无 >50 ms 帧 |
| 10,005 条历史 | 挂载 9 行、会话流 DOM 113 个节点 |
| 向前加载锚定 | 3 次位移均为 0 px |
| `@` 候选弹层、焦点留在输入框 | 通过（合成 input 事件） |
| 中文 IME（手动） | 用户确认正常；记录到 `compositionupdate` 先于 `compositionstart` 触发一次，IME 安全的 Enter 判断不能依赖事件先后，应以 `isComposing` / `keyCode 229` 为准 |
| 剪贴板图片粘贴（手动） | `image/png` 文件 583 KB，可读取并预览 |
| 拖放图片文件（手动） | `Files` 类型，文件名、类型、大小正确 |

### 宿主 ↔ 页面

| 项目 | 结果 |
| --- | --- |
| 绑定调用往返 | 平均 0.08 ms（300 次） |
| Channel 快速推送 | 20,000 条 × 200 B，宿主 40 ms 发完、页面 46 ms 收完，顺序无误 |
| 背压 | 页面每条忙等 2 ms 时，宿主 `Send` 最长阻塞 256 ms（14 次超过 5 ms）；全部送达 |
| 页面关闭 Channel | 宿主 `Send` 返回关闭错误、context 取消 |
| 按 token 节奏推送（每 16 ms） | 延迟最大 5 ms |
| 每包独立 origin | `aibopkg://pkg-a`、`aibopkg://pkg-b` 互不相同 |

### 工具面板：iframe 不可行，独立 WebView 可行

| 项目 | 结果 |
| --- | --- |
| iframe 中的 `window.mygo` | 同源与跨 origin 的 iframe 均没有，与文档“只有顶层页面能调用 Go”一致 |
| 经父页面中转调用 | 往返平均 0.16 ms |
| iframe 忙等 3 秒 | 父页面定时器最大间隔 3000 ms——同源与跨 origin iframe 都与父页面同线程 |
| iframe 死循环 | 父页面随之卡死，无法移除该 iframe |
| 独立窗口中的面板 | 独立 WebContent 进程；面板死循环或被结束，主窗口照常 |
| 同一原生窗口中的第二个 WebView | 独立 WebContent 进程；顶层页面可直接调用 Go；面板死循环或被结束不影响工作台，面板 `Reload` 后恢复 |

### 卡死、崩溃与恢复入口

| 项目 | 结果 |
| --- | --- |
| 页面死循环时原生主线程 | `RunOnMain` 最大延迟 0.03–0.17 ms；菜单与托盘可用（手动确认） |
| 死循环页面 `Reload()` / `LoadURL()` | **均不能恢复**（等待 6–15 s 未重新载入） |
| 结束 WebContent 进程后 `Reload()` | `OnRenderProcessGone("terminated")`，约 40 ms 恢复，新进程号 |
| 心跳中断提示 | 约 4.1–4.7 s 检测到，原生提示可选“继续等待 / 重新载入 / 切回默认界面”（手动确认）。发现：窗口关闭后重新打开时，因旧心跳过期误报一次；正式实现须在窗口创建与 DOM 就绪时重置心跳 |
| 菜单“重新载入界面 / 切回默认界面” | 页面无响应时先结束渲染进程再载入，手动确认可恢复 |
| 关闭最后一个窗口 | 应用留在后台；Dock 或托盘“显示工作台”重新打开（手动确认） |
| 单实例 | 第二个实例把参数 `open-session-demo` 交给第一个实例后退出 |

### 子进程（Q87）

应用被 `kill -9` 后：

| 子进程 | 是否残留 |
| --- | --- |
| 读 stdin、EOF 即退出（ACP Agent 的典型行为） | 否 |
| 不读 stdin | **是** |
| 读 stdin 的子进程派生的、不读 stdin 的孙进程 | **是** |

每个子进程以 `Setpgid` 建立独立进程组；按记录的进程组号 `killpg` 后，残留的子进程与孙进程全部结束。macOS 没有父进程退出信号，“随应用结束”须由“stdin EOF + 重启时按进程组清理”组合实现；进程号可能复用，记录时应同时保存启动时间并核对。

### SQLite（`modernc.org/sqlite` v1.62.0，SQLite 3.54.1，无 cgo）

| 项目 | 结果 |
| --- | --- |
| 8 个会话并发流式写入，单写入者按 100 ms / 64 KiB 合并提交 | 200,000 行（每行 400 字）0.70 s，约 28 万行/秒；1,231 次提交，最长 7 ms |
| 分段读取（某会话 50 条一页） | 平均 0.06 ms |
| FTS5 trigram（独立索引库，contentless） | 可用；重建 200,000 行 5.2 s |
| 中文查询 | “子任务补读” 3.3 ms、“暂停，恢复” 2.1 ms |
| 英文高频词 | “approval” 17 万命中 50 ms |
| 两字查询 | trigram 无结果（符合预期）；有上限的 `LIKE` 扫描 0.5 ms |
| 体积 | 业务库 89 MB、索引库 119 MB |

## 未验证

- Windows / WebView2：当前有效的原生界面、工具面板与进程验证已由 [17 号票](../../.scratch/aibo-rev-architecture/issues/17-windows-presentation-validation.md#answer)收口；旧默认 WebView 呈现不在范围。WebView2 的 `ProcessID` 是共享浏览器进程，恢复按已接受的 ADR 0008 执行，不能照搬 macOS 的结束单面板进程方式。
- 原生窗口 + 多 WebView 作为主窗口时，面板区域随窗口缩放与面板拖拽调整的同步表现、跨 WebView 拖放。
- 屏幕阅读器、真实 Agent 负载下的长会话、打包后的 `.app`（本实验直接运行可执行文件）。
