# ACP v1 类型生成与 stdio 连接层

Status: open
Type: task
Blocked by: 01

## Question

从官方 ACP v1 schema 生成 Go 类型，放在按协议版本区分的包中（ADR 0010 第 2 条）。实现 stdio 上的 JSON-RPC 连接层（ADR 0005 §3）：先落库再发送、区分已发送未结算、按连接代际拒绝迟到结果、保留未识别字段原文。用模拟 Agent 测试，并用本机 Codex（codex-acp 2.6.0）跑通 initialize 与 session/new。
