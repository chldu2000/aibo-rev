# ui-checks：界面硬性要求的本地验证

对应 [docs/design/ui-requirements.md](../../docs/design/ui-requirements.md)。不依赖产品代码，可以先对原型运行，实现建立后换成实现目标。

## 准备

```bash
cd tools/ui-checks
pnpm install                                        # playwright-core、axe-core
npx playwright-core install webkit                  # 默认浏览器，接近 macOS WKWebView
npx playwright-core install chromium-headless-shell # 可选
```

## 运行

在仓库根目录：

| 命令 | 内容 |
| --- | --- |
| `node tools/ui-checks/run.mjs all` | 依次运行下面四项；任一失败以非零状态退出 |
| `node tools/ui-checks/run.mjs selftest` | 用 `selftest/bad` 与 `selftest/good` 确认每条静态规则能命中违规、不误报 |
| `node tools/ui-checks/run.mjs static` | 令牌完整性、对比度、令牌同步、左侧信号条、硬写颜色、最小字号、焦点环、减少动态效果、禁用文案 |
| `node tools/ui-checks/run.mjs requirements` | 要求文档的编号、字段、Checklist 引用与测试映射 |
| `node tools/ui-checks/run.mjs e2e [--target 名称] [--browser webkit\|chromium] [--grep 编号]` | 按测试契约运行验收测试 |
| `node tools/ui-checks/build-tokens.mjs` | 修改 `tokens.json` 后重新生成 `tokens.css` |

e2e 会在本机 `127.0.0.1` 的随机端口启动静态服务；受限的沙箱环境可能需要允许本地监听。

## 结构

| 路径 | 职责 |
| --- | --- |
| `ui-checks.config.json` | 检查范围（`styleRoots`、`copyRoots`）、禁用文案、验收目标 |
| `checks/tokens.mjs` | `tokens-complete`、`contrast`、`token-sync` |
| `checks/styles.mjs` | `no-left-accent`、`no-hardcoded-color`、`min-font-size`、`focus-outline`、`reduced-motion` |
| `checks/copy.mjs` | `forbidden-copy` |
| `checks/requirements.mjs` | 解析要求文档并核对 Checklist 与测试 |
| `e2e/harness.mjs` | 静态服务、浏览器、按夹具打开页面、axe 与运行时样式检查 |
| `e2e/acceptance.test.mjs` | 验收测试，标题以要求编号开头 |

## 接入实现

1. 实现提供要求文档 §11 的 `data-testid` 与 `?fixture=` 夹具（开发与测试构建）。
2. 在 `ui-checks.config.json` 的 `targets` 中新增目标，例如：

   ```json
   "default-ui": { "url": "http://127.0.0.1:5173/", "fixtures": ["running-approval", "delivery-unknown", "pending-question", "turn-failed", "untrusted", "idle-empty", "restart", "save-failure", "history-10k"] }
   ```

   有 `serve` 字段时脚本自行启动静态服务；有 `url` 时连接已运行的开发服务器。可把 `defaultTarget` 改为该目标。
3. 把实现的源码目录加入 `styleRoots` 与 `copyRoots`。
4. 运行 `node tools/ui-checks/run.mjs all`。

## 新增或修改要求

- 在要求文档中新增条款，`验证` 字段写 `e2e`、`perf`、`static:<检查名>` 或 `manual`。
- 声明 `e2e`/`perf` 时，在 `acceptance.test.mjs` 中加入标题以该编号开头的测试；`requirements` 检查会核对两边是否一致。
- 新增静态规则时在 `run.mjs` 的 `STATIC` 中注册，并在 `selftest/bad` 与 `selftest/good` 中各加样例。

## 当前状态（2026-11-10，静态原型目标）

- 规则自检、静态检查、要求文档检查全部通过。
- e2e：48 项，WebKit 与 Chromium 均为 46 通过、2 跳过（`UI-STREAM-04/05` 需要 `history-10k` 夹具，静态原型不提供）。
- 这只证明脚本与契约可用；静态原型不是实现，通过不代表产品已满足要求。
