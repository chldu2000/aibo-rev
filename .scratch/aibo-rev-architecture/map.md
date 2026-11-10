# Aibo Rev 全 ACP 与框架中立架构

Status: open
Labels: wayfinder:map

## Destination

形成可据此实施的 Aibo Rev 架构规格：覆盖[完整版本 Checklist](../../docs/refactor-feature-inventory.md)当前 85 项必需功能，明确模块职责、进程与数据所有权、ACP 与呈现扩展接口、技术选型及验收方式。完成条件是实施前的重要架构问题均已有决策和依据；本图不编写产品代码。

## Notes

- 最高依据为本轮用户要求及 Q1–Q4 回答（2026-11-01）：全部 Agent 经 ACP 接入，不区分内置/外部；允许协商后的 ACP 命名空间扩展；不要求直接兼容旧 Agent 插件或呈现包；目标为可实施架构规格。
- 若采用 Web 呈现，默认 UI 可使用任一框架，但内核、交互状态及共享呈现接口不得依赖该框架。完整替换默认 UI 与使用不同框架的呈现包，均应无需修改内核；这必须有验收证据。
- 2026-11-04 用户撤销正式版本的特定设计系统/呈现风格要求。F20.1 保留内置默认呈现，暂不指定 Material 3 或其他风格；完整流程、可替换性、主题与状态恢复承诺继续成立。
- 2026-11-04 阶段收存：用户认可呈现原型基本通过，要求提交主架构文档；正式技术栈、用户界面与交互方案由用户后续单独设计。本次提交保存已确认的架构基线和证据，不自动认领后续选型或设计票。地图保留未完成的正式设计、验收与切换问题，不将本次提交称为全部实施规格已完成。
- 2026-11-10 用户决定技术栈：MyGO + 系统 WebView、Go 宿主、默认呈现 Svelte 5 + Bits UI + virtua（markdown-it、Shiki），见 [ADR 0003](../../docs/adr/0003-desktop-and-ui-stack.md)；11 号票其余问题仍开放。
- 主目标是降低维护成本。旧 aibo 的代码、合同、存储及语言不是新架构约束；已有行为和失败经验仍是参考。
- Checklist 是功能选择的唯一入口；83 项来自提交 `86b8efd`，2026-11-10 用户加选 F17.1（插件工具面板）与 F17.5（隐藏工具视图后保持后台运行），共 85 项。未选项不等于永久删除，旧文档中的 Git/编辑器/终端阶段建议不能扩大本图范围。数据正确性、恢复及必要保留规则随所选功能成立。
- 参考 [Zed 实现研究](../../docs/zed-acp-research.md)中的结构化会话内容、状态合并、类型驱动呈现和动态配置。其“保留宿主原生协议接入”的候选建议已被本轮全部 ACP 的要求取代。
- 历史与引用以本轮[业务历史与跨 Agent 引用的权威来源](issues/06-history-reference-authority.md#answer)为准，不直接继承终端路线对原生格式的要求；遵循 [CONTEXT](../../CONTEXT.md)，不能以归一化历史替代原始 transcript 的术语定义。
- 旧包格式兼容、新包升级的关联会话迁移、旧应用数据迁移是不同问题。旧数据迁移范围待决，不读取或清理真实运行库。
- 采用 [本地 Markdown tracker](../../docs/agents/issue-tracker.md)。查询子票获取 open/unblocked/unclaimed frontier，地图不重复列开放票。每张票必须先 claim，再开展工作；并发研究分别使用独立 worktree 和 `research/*` 分支。
- 每次会话先读本图及领域词汇；讨论票使用 [grilling](/Users/chldu/.agents/skills/grilling/SKILL.md)、[domain-modeling](/Users/chldu/.agents/skills/domain-modeling/SKILL.md) 和 [codebase-design](/Users/chldu/.agents/skills/codebase-design/SKILL.md)，研究票使用 [research](/Users/chldu/.agents/skills/research/SKILL.md)。人类决策由用户本人回答。
- 遵循 [wayfinder](/Users/chldu/.agents/skills/wayfinder/SKILL.md)：建图会话不解决人工决策票；后续每次最多解决一张，研究票除外。详细答案只存子票，地图只保留索引。

## Decisions so far

<!-- 一行对应一张已解决的子票；详细答案只在子票中。 -->

- [ACP 对必需会话能力的实际覆盖](issues/01-acp-capability-evidence.md)：基础能力已有共同接入依据，高级能力仍需逐合同协商；目标动作及原始记录可获取性存在差异。
- [跨框架呈现与桌面容器的可行边界](issues/02-presentation-platform-evidence.md)：通信与装载可跨框架，权限和故障隔离须分别验证；已据此形成后续原型问题。
- [工作台视图与会话运行的生命周期](issues/03-runtime-lifecycle.md)：运行随应用存活且独立于视图，关窗后台运行，退出停止推进；重启需明确恢复，归档/移除须先结束活动。
- [统一 ACP 接入与 Agent 能力承诺](issues/04-agent-capability-contract.md)：统一基础门槛、逐项交付平台功能；扩展独立映射，缺失能力说明原因，正式支持绑定已验证组合。
- [框架中立的工作台呈现接口](issues/05-presentation-interface.md)：单一完整呈现包复用无界面交互，共通状态跨包恢复；默认列表导航，明确失败自动回退，无响应由独立入口手动恢复。2026-11-10 补充：完整包须提供后台状态展示、工具面板承载位置与会话详情入口。
- [业务历史与跨 Agent 引用的权威来源](issues/06-history-reference-authority.md)：宿主独立保存原文，加入引用时固定范围；显式授权读取同版本的历史、相关子任务及已保存附件，来源失效不破坏已有引用。
- [ACP 会话控制与投递证据的实际边界](issues/14-acp-control-evidence.md)：请求结算、消费与原生活动停止需分开取证；插话回退及恢复存在具体实现限制，ACP 不提供通用沙箱或重试幂等保证。
- [会话控制、授权与消息投递的责任](issues/07-session-control-and-delivery.md)：宿主持有队列与尝试证据，取消/失败暂停推进；按确认配置投递，未知只允许明确手动重发，信任撤销与停止核实分开。
- [ACP 目标控制与子任务补读的实际边界](issues/15-acp-goal-task-evidence.md)：目标动作与回合停止语义不同；父会话恢复可能启动执行，不能替代独立只读补读。
- [持续目标、子 Agent 与后台任务的职责](issues/08-goals-and-background-work.md)：Agent 原生推进目标，宿主分别协调目标/回合/队列/任务；保存已推送历史，额外内容按需只读补齐，子 Agent 沿用父会话引用范围。
- [现有插件职责与升级边界的事实基线](issues/16-package-lifecycle-evidence.md)：旧能力包不限于 MCP；旧迁移会恢复原生会话，宿主状态回退不保证外部数据与副作用可逆。
- [Agent 包、能力包与呈现包的生命周期](issues/09-package-lifecycle.md)：可信扩展分别提供 ACP 接入、工具/上下文和完整呈现；执行包在相关活动结束后单版本变更，升级先迁移宿主副本，卸载保留恢复条件。2026-11-10 修订：能力包可贡献工具面板（F17.1、F17.5）。
- [持久数据、搜索与诊断的所有权](issues/10-data-ownership-and-search.md)：业务内容先保存后发布，引用共享固定内容；索引可重建、诊断独立限额，无主托管资产延迟复核后回收，关键保存故障暂停推进并核实停止。
- [跨框架替换与呈现故障回退的最小原型](issues/13-presentation-conformance-prototype.md)：macOS 上已取得两种框架替换及故障恢复证据，用户认可原型基本通过；正式技术栈、界面与交互留待单独设计，原型实现不成为产品选型。

## Not yet specified

- 真实接入与协议样本验证可能暴露固定源码研究之外的适配和兼容问题；已有产品承诺见上述能力决议，后续按具体证据为新增缺口建票。
- 真实负载与故障实验可能暴露当前合同之外的容量或恢复问题；已知验证场景由后续架构与验收票承接，仅为新增疑问另建决策票。

## Out of scope

- 本轮产品实现、部署发布、对正式数据执行迁移或清理。
- 宿主直接接入非 ACP 的 Agent 协议，以及按内置/外部 Agent 分两条执行路径：用户已明确排除。
- 直接运行旧格式 Agent 插件或呈现包的兼容层：Q4 已明确不要求。
- 为未勾选的 Git 工作流、源码编辑器、IDE 语言服务、终端面板、完整原生 Agent 管理或数据管理 GUI 制定详细实现方案；今后可由新范围进入其他地图。
- 终端优先、VS Code 扩展或 TUI 转 GUI 作为产品主线。
