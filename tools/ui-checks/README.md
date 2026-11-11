# ui-checks：界面硬性要求的本地验证

对应 [docs/design/ui-requirements.md](../../docs/design/ui-requirements.md)。Go 实现（2026-11-11 起，默认界面为 MyGO 原生 UI，见 [ADR 0006](../../docs/adr/0006-native-ui.md)）；需要 Go 1.27.3，`CGO_ENABLED=0`。

## 运行

在仓库根目录：

| 命令 | 内容 |
| --- | --- |
| `go -C tools/ui-checks run . all` | 依次运行下面四项；任一失败以非零状态退出 |
| `go -C tools/ui-checks run . selftest` | 用 `selftest/bad` 与 `selftest/good` 确认每条源码规则能命中违规、不误报 |
| `go -C tools/ui-checks run . static` | 令牌完整性、对比度、令牌同步，以及界面 Go 源码的规则：硬写颜色、最小字号、左侧信号条、焦点环、减少动态效果、禁用文案 |
| `go -C tools/ui-checks run . requirements` | 要求文档的编号、字段、Checklist 引用与验收用例映射 |
| `go -C tools/ui-checks run . uitest [go test 参数]` | 在 `targets` 的每个目录中运行 `go test -run <run>` |
| `go -C tools/ui-checks run . tokens` | 修改 `tokens.json` 后重新生成 `tokens.css` 与 `tokensGo` 中的 Go 令牌文件 |

## 结构

| 路径 | 职责 |
| --- | --- |
| `ui-checks.config.json` | 检查范围（`goRoots`、`copyRoots`）、生成的令牌文件、禁用文案、验收目标 |
| `tokens.go` | `tokens-complete`、`contrast`、`token-sync`，以及 `tokens.css` 与 Go 令牌的生成 |
| `gosrc.go` | 对界面 Go 源码的静态规则（`go/ast`）；跳过 `_test.go` 与生成文件 |
| `requirements.go` | 解析要求文档并核对 Checklist 与验收用例 |
| `acceptance/` | 验收套件：`Target` 接口、夹具、契约名称（`contract.go`）与用例（`cases.go`），用 MyGO `ui.Tester` 驱动 |

## 接入实现

1. 在实现的测试中提供一个 `acceptance.Target`，并调用 `acceptance.Run(t, target)`（参考 `prototypes/native-ui/acceptance_test.go`）；实现的 `go.mod` 用 `replace aibo-rev/ui-checks => <相对路径>` 引用本模块。
2. 界面使用要求文档 §11.1 的无障碍名称。
3. 在 `ui-checks.config.json` 中：把实现的源码目录加入 `goRoots` 与 `copyRoots`，把其令牌文件加入 `tokensGo`，在 `targets` 中新增目标；补齐用例后把 `requireAllUITests` 改为 `true`。
4. 运行 `go -C tools/ui-checks run . all`。

## 新增或修改要求

- 在要求文档中新增条款，`验证` 字段写 `uitest`、`perf`、`static:<检查名>` 或 `manual`。
- 声明 `uitest`/`perf` 时，在 `acceptance/cases.go` 中加入 `ID` 为该编号的用例。
- 新增源码规则时在 `main.go` 的 `sourceRules` 中注册，并在 `selftest/bad/bad.go` 与 `selftest/good/good.go` 中各加样例（两个文件带 `//go:build ignore`，只被解析，不编译）。

## 当前状态（2026-11-11）

- 规则自检、静态检查全部通过；要求文档检查通过，提示 41 条要求的验收用例尚未编写。
- 验收用例 5 条（UI-GEN-04、UI-STREAM-04、UI-STREAM-05、UI-COMP-01、UI-COMP-04），在唯一的目标 `prototypes/native-ui` 上全部通过。原型只提供 `idle-empty` 与 `history-10k`；其余用例须等产品实现成为验收目标后编写，原 WebView 版的 48 项 DOM 测试保留在 git 历史中（`757b6eb`）作参考。
- 原型不是实现，通过不代表产品已满足要求。
