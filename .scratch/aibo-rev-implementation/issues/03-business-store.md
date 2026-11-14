# 业务库：SQLite、单写入 goroutine 与格式迁移

Status: open
Type: task
Blocked by: 01

## Question

按 ADR 0005 §2 与 [10 号决议](../../aibo-rev-architecture/issues/10-data-ownership-and-search.md#answer)实现业务库：`modernc.org/sqlite`、WAL、宿主内单一写入 goroutine、版本化迁移（迁移前 `VACUUM INTO` 副本）、数据目录与开发用覆盖变量。第一条链路需要的表：工作区与信任、会话、条目（原文）、投递尝试。关键写入失败按 UI-LIFE-03 暂停推进。
