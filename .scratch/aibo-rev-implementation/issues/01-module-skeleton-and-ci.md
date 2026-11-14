# 产品 module 骨架、依赖方向检查与 CI

Status: open
Type: task
Blocked by: none

## Question

按 [ADR 0010](../../../docs/adr/0010-implementation-start.md) 建立单个 Go module `github.com/chldu2000/aibo-rev`（Go 1.27.3、MyGO v0.9.0、`CGO_ENABLED=0`），四个顶层包目录对应 ADR 0006 的运行核心、交互层、界面包、桌面宿主；可启动一个空的工作台窗口。依赖方向（ADR 0006 约束 1）由自动检查强制，违反即失败。产品界面包接入 `tools/ui-checks` 成为验收目标。配置 GitHub Actions：macOS arm64 上 vet/test/ui-checks，Windows x64 上编译与 test。
