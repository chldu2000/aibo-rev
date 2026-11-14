# 官方 Codex Agent 包与 Node 运行时选择

Status: open
Type: task
Blocked by: 01

## Question

按 [ADR 0009](../../../docs/adr/0009-official-agent-packages-and-node-runtime.md)：把 codex-acp 打包为单个 JS 文件（排除 `@openai/codex`），做成预装的 `.aibopkg`（清单含 CLI 指引与文档链接、Node 版本范围、检查命令）。宿主实现 Node 的选择顺序（手动 → 系统含版本管理器 → Aibo 下载）、用户触发的最新 LTS 下载与 SHA-256 校验、CLI 检测并经 `CODEX_PATH` 传入。macOS 上验证从 Dock 启动时能找到系统 Node。
