---
status: accepted
date: 2026-11-15
revises: ADR 0005 §4 的“语言运行时”（11 号票 Q104）
---

# 官方 Agent 包与 Node 运行时

用户于 2026-11-15 决定：官方 Agent 包不带 CLI 本身，只带安装指引或文档链接，检测到没有安装 CLI 时提供给用户；官方 Agent 包尽量是用 JS 脚本就能运行的形式；需要 Node 时，先检测用户已有的 Node，或者提供入口让用户为 Aibo 下载一份最新的 LTS 版 Node。同时确认官方 Agent 包随安装程序预装（12 号票 2026-11-15 记录）。

本文取代 ADR 0005 §4 中“宿主不内置 Node、不自动下载；官方 Agent 包把 Node 适配器编译为单文件可执行文件”一句，并因此勾选 F02.7（Node 运行时自动查找、手动选择与按需下载）。其余包合同（zip 交付、清单、启动/检查/迁移命令）不变。

## 依据（2026-11-15 读取 npm 包元数据与源码，未运行）

- `@agentclientprotocol/claude-agent-acp` 0.106.0 要求 Node ≥ 22；依赖 `@anthropic-ai/claude-agent-sdk` 0.3.322，后者以各平台的 optional 依赖携带 Claude CLI。适配器优先使用环境变量 `CLAUDE_CODE_EXECUTABLE` 指定的 CLI（`dist/claude-cli.js`）。
- `@agentclientprotocol/codex-acp` 2.6.0 依赖 `@openai/codex`（npm 上的 Codex CLI）；`CODEX_PATH` 可指向本机的 `codex`（[`acp-concurrency`](../../prototypes/acp-concurrency/README.md) 已这样运行）。
- 旧 Aibo 已实现 Node 查找与下载（`aibo/src-tauri/src/node_runtime.rs`）：候选顺序为手动选择、系统、Aibo 管理的副本；只下载锁定版本并核对 SHA-256；探测时继承 `NVM_DIR`、`VOLTA_HOME`、`ASDF_DATA_DIR` 等变量。

## 决定

1. **官方 Agent 包的形式。** 包内是打包成单个 JS 文件的适配器（含其 JS 依赖），由 Node 运行。打包时排除 CLI：不包含 `@openai/codex` 与 Claude Agent SDK 的平台 optional 依赖。启动时由宿主把检测到的 CLI 路径经 `CODEX_PATH` / `CLAUDE_CODE_EXECUTABLE` 传给适配器。
2. **外部依赖与指引。** 清单声明两类外部依赖：CLI（名称、版本范围、检测方式、安装指引与官方文档链接）与 Node（版本范围，取各适配器 `engines` 的要求）。检查命令报告“已满足 / 缺失 / 未知”；缺 CLI 时界面显示包内的指引与链接，不代为安装。认证状态同样由检查命令说明（F02.6）。
3. **Node 的选择顺序。** 用户手动选择的 Node → 用户系统中已有的 Node → Aibo 下载的 Node。只采用满足包版本要求的候选。系统 Node 的查找不能只看 GUI 进程的 `PATH`（macOS 从 Dock 启动的应用不继承 shell 的 `PATH`）：另查常见安装位置与版本管理器（Homebrew、nvm、fnm、Volta、asdf 等），以 `node` 自报的版本与真实路径为准。
4. **下载入口。** 只在用户操作时下载：缺少可用 Node 时的提示，以及设置中的 Node 运行时一项。下载 nodejs.org 当时最新的 LTS 版（按官方 `index.json` 的 `lts` 字段），按同目录的 `SHASUMS256.txt` 核对后解压到数据目录下按版本区分的目录；新版本校验并探测通过后才切换，旧版本在没有进程使用后删除。不后台自动更新；设置中显示当前版本与来源，并提供“更新到最新 LTS”。
5. **预装。** Codex 与 Claude Code 的官方 Agent 包随安装程序预装，仍按普通包管理，可以升级与卸载（F19）。首次启动的欢迎页因此能列出这两个 Agent 及其检查结果（[界面方案 §12](../design/ui-interaction.md#12-首次使用与空态)）。

## 后果

- 宿主新增一个 Node 运行时模块（查找、探测、下载、校验、选择），这是 ADR 0005 当初想避免的维护成本；换来的是官方包不必为每个平台编译可执行文件，包体积小，适配器升级只是替换一个 JS 文件。
- 下载的 Node 由用户决定何时更新，宿主不跟随其安全更新；设置中须显示版本，便于用户判断。
- 第三方 Agent 包也可以声明 Node 依赖并共用这套选择逻辑。

## 尚未验证

- 两个适配器打包成单个 JS 文件后能否正常运行，以及排除 CLI 后用 `CODEX_PATH` / `CLAUDE_CODE_EXECUTABLE` 指向用户安装的 CLI 是否完整可用（Claude Code 部分并入 [20 号票](../../.scratch/aibo-rev-architecture/issues/20-acp-same-connection-concurrency-claude-code.md)）。
- Windows 上 `node.exe` 及其子进程在 Job Object 中随应用结束（17 号票验证的是其他子进程）。
