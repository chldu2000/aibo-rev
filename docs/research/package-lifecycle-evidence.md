# 现有插件贡献与升级边界的事实基线

研究日期：2026-11-04。支持[Agent 包、能力包与呈现包的生命周期][decision]及[现有插件职责与升级边界的事实基线][ticket]，只提供事实与待定边界，不选择新包格式、技术栈或迁移机制。
本次只读相邻仓库的文档与源码，没有启动插件、安装依赖、读取真实会话库或修改旧仓库。下文“实现”不等于本次验收。

## 来源基线

| 仓库 | 读取时 HEAD | 工作区状态 |
| --- | --- | --- |
| [aibo][host-readme] | `de70665d635c5fbc2ed79f254cb43843b28621a1` | 脏：CONTEXT、ADR、package.json 及未跟踪文件；下列实现、合同与现行插件文档未列为修改 |
| [aibo-plugins][plugins-readme] | `53e05985b48fac07e5787d210fb39fddb9f99811` | 干净 |
| [aibo-rev][rev-readme] | `86b8efdac75c96e6c6e8daa5fa7020647b61b726` | 脏：当前术语、Checklist、架构票及研究文档含未提交内容；本文采用主工作区的当前决议，不把它们归入该 HEAD |

相邻源码链接按 `aibo-dev/{aibo-rev,aibo,aibo-plugins}` 布局解析。旧验收记录只说明其记录的版本、平台和场景；本轮未重跑。[旧支持矩阵][matrix]

## 1. capability plugin 的真实含义与范围缺口

**现成反例：它不等于 MCP 包。** `aibo-plugins/plugins/capability` 是开发起点，不是 Agent：1.0.2 清单声明 application 作用域的只读 `capabilityProvider`，以及 `command` 入口的 `semanticView/detail`；Worker 返回 `state/view/actions`，展示 greeting 并提供 refresh。它走 Capability Runtime 2.0，没有 MCP 服务声明。[说明][greeting-readme]、[清单][greeting-manifest]、[实现][greeting-worker]

| 概念 | 旧实现中的事实与边界 |
| --- | --- |
| 能力提供者 | 以版本化操作提供服务、业务数据或读写行为；声明作用域、schema、effect 与权限，声明本身不授予执行权。写入夹具确有 workspace 作用域的 append 操作。[开发指南][development]、[写入夹具][write-fixture] |
| Agent／session provider | 旧 Agent 是提供 `aibo.session.open` 等会话操作的能力提供者，使用 Runtime 2.1；原生 ID/recovery 属于提供者，宿主掌握会话、回合、授权和持久历史。不能因新架构全 ACP 而把这个旧包分类当成新格式要求。[开发指南][development] |
| 语义视图与动作 | `workspace.tool`、`session.context`、`session.action`、`settings.page`、`command` 是语义贡献入口；collection/detail/settings/inspector 是数据结构，宿主负责绘制与动作校验。`workspace.tool` 不是“向 Agent 注册一个工具”。[合同][manifest]、[开发指南][development] |
| workspace action／工程 task | test/lint/build/custom 是宿主 `project_actions` 的定义、执行和记录路径；不能把已有工程任务直接算作插件贡献。语义写动作可以调用能力操作，但二者不是同一合同。[工程任务实现][project-actions]、[写入夹具][write-fixture] |
| 子 Agent task | Cursor 扩展将原生 task 工具及 `cursor/task` 通知映射为子 Agent 状态；它不是上述 test/build 任务，也不说明存在独立的 task 包类型。[扩展实现][cursor-extension] |
| 宿主工具与 MCP | 历史读取由宿主工具目录、授权和读取器提供；插件接动态工具注册或 MCP bridge。ACP Worker 把 bridge 配置交给 `session/new/load` 的 `mcpServers`；MCP 是这条接线使用的协议，不是能力插件的总定义。[宿主工具设计][host-tools] |
| toolView | 能力包可另带独立交互应用的前端、原生后端和领域状态；终端是实际消费者。宿主管窗口/工作区身份、通信与停止监督，Agent 工具目录不暴露该通道。[toolView 合同][tool-view]、[终端说明][terminal] |
| 管理面板 | 旧实现另支持 `managementPanel`，与 capabilityProvider、semanticView、toolView 分别判定；并不是每个能力插件都必须提供管理中心。[激活实现][manifest-code] |
| ACP extension mapping | Cursor Worker 调共享 `serveAcpAgent` 并传扩展对象；扩展处理厂商模式、参数、追问、计划与 task 通知。它是协议解释角色，不能与 MCP 工具或可见工具面板混称。[Worker][cursor-worker]、[扩展实现][cursor-extension] |

**当前范围缺口：**F19.1 已选本地安装 Agent 和能力插件，但 F13–F18 的细项全部未选，F17 的工具面板/终端也全部未选。greeting 是教学样例，工程任务、Git、面板和完整 Agent 设置中心不能据旧接口存在自动纳入；目前尚无一个具体已选的“非 Agent 能力插件贡献场景”。这是待用户定界的空白，不是删除能力插件或改称 MCP 包的依据。[Checklist][checklist]
F07.6 的按需读取已保存历史确为已选，但旧证据中它是宿主固定工具经 Agent 接线，不能单凭这一项推导通用第三方工具安装平台。[Checklist][checklist]、[宿主工具设计][host-tools]

## 2. 安装、启停、替换与迁移实际处理什么

| 操作 | 当前旧实现的可核查行为 |
| --- | --- |
| 本地安装与启用 | 安装拷贝包并登记 release，首次默认禁用；同版本同 digest 幂等，同版本异内容为替换。启用校验宿主/合同/平台、可执行依赖与包依赖，普通启用不额外迁移会话。登记、启用和运行握手是不同阶段。[边界文档][boundaries]、[registry][registry]、[支持矩阵][matrix] |
| 停用 | 入口先处理 toolView 关闭，再禁用、使依赖贡献失效并停止对应运行；不删除宿主历史或改绑会话。仅停用与卸载后永久仅历史不是同一后果。[生命周期入口][commands]、[边界文档][boundaries] |
| 升级／替换 | 预览绑定包 digest 与引用快照，运行调用、未决候选和固定依赖会阻止替换；保存替换记录，暂存候选并复制私有实例数据，逐会话验证后才提交当前版本。源码保留旧包供失败恢复或有限撤销。[替换实现][replace]、[替换记录][replacement] |
| 已建立绑定的会话 | 候选以旧 recovery 执行 resume，须保持同一 nativeSessionId、执行权限后端及原有基础能力；先停妥校验实例，再写新绑定并清空 runtime generation。候选失败不提交该会话的新绑定。[迁移实现][migration] |
| 关联会话的不同路径 | 默认可跳过归档会话并转为仅历史；未建立原生绑定者验证贡献后重绑；空会话，以及严格证明首条失败输入 `not_sent` 的有限情况可重建并恢复草稿，不自动重发。其余恢复仍需成功。这些是旧产品取舍，非新方案默认值。[边界文档][boundaries]、[替换实现][replace] |
| 失败与崩溃恢复 | 替换失败路径先停止候选，再恢复整体宿主绑定/启用状态并清理；停止或清理失败仍可能留下待恢复工作，启动处理未完成记录。独立 `migrate_release` 只接受同 pluginId 的更新版本，逐会话返回 migrated/failed，不能将它误称为整个批次原子迁移或跨 Agent 状态转换。[替换实现][replace]、[替换记录][replacement]、[迁移实现][migration] |
| 卸载 | 重验影响范围，阻止破坏固定依赖；先禁用并停止运行，再删除程序与该 release 私有目录。保留安装身份、消息、附件、执行记录和会话配置；明确保留历史并停用的会话永久仅历史，重装不会恢复执行。[迁移实现][migration]、[registry][registry]、[边界文档][boundaries] |

## 3. 四类数据不能合并为“插件数据”

| 数据 | 实际所有权及迁移边界 |
| --- | --- |
| 包私有实例数据 | 宿主管理 `plugin-data/<plugin>/<installation>/v1/<instance>` 路径，插件解释内容；迁移复制目录、改写 recovery 内旧路径，拒绝冲突的非空来源。它不等于所有 Agent 原生存储。[迁移实现][migration]、[替换实现][replace] |
| 会话绑定与 recovery | 宿主 binding 固定 installation/contribution、nativeSessionId 和带 schema/version 的 recovery；恢复数据由提供者产生。它是恢复依据，不是完整 transcript 或原生状态备份。[绑定合同][binding]、[会话迁移说明][session-migration] |
| 宿主业务历史与共通状态 | 持久消息、事件、附件及宿主身份独立于插件可运行性；旧替换快照只保存 release、session binding/capabilities 与能力选择，不回滚业务历史。新决议另明确草稿、附件/引用及共通阅读状态归共享层，包私有呈现状态按包保留。[替换记录][replacement]、[边界文档][boundaries]、[呈现决议][presentation] |
| Agent 原生数据 | 通用 ACP 恢复调用 Agent 的 `session/load`，并重新选择模式/模型/参数；recovery 保存 ID、工作区、配置与后台状态引用，并不复制 Agent 的全部存储。旧 Pi 也有 `sessionFile` 路径恢复，说明原生数据位置随提供者而异。[ACP 实现][acp-session]、[Pi 实现][pi-engine] |

**回退的保证限于有记录的宿主管理状态。**`replacement::restore` 恢复数据库中的安装、绑定、能力选择及 retirement 状态，安排候选目录清理；没有恢复 Agent 外部会话库、配置或已执行工具副作用的通用操作。且候选 resume 自身会启动 Agent、加载会话并设置模式/模型。由此推论：恢复旧绑定不证明原生状态可逆，更不证明旧包之后仍能恢复；若升级触及外部原生数据，需要另给可验证的兼容/失败承诺。[替换记录][replacement]、[ACP 实现][acp-session]

## 4. 对职责与升级取舍的输入

- **已定职责继续有效。**标准 ACP 由共享模块处理，厂商扩展解释可随 Agent 包交付；宿主继续拥有授权、队列、历史、投递与恢复事实。包交付映射不转移这些责任，也不要求旧 Runtime/SDK 继续存在。[Agent 决议][agent-decision]、[生命周期前置约束][decision]
- **呈现包已独立定义。**新方案是用户信任的单一完整包，共享交互流程与共通呈现状态；旧受限视觉树、可选 surface 继承不能重新覆盖这个选择。其升级须区分共同状态兼容与包私有状态迁移。[呈现决议][presentation]、[旧支持矩阵][matrix]
- **维护负担来自交叉状态。**旧实现同时处理版本/digest、包依赖、固定会话与能力绑定、私有目录冲突、归档/空会话、投递证据、运行代际、校验实例清理，以及跨文件/数据库的中断恢复。即使不复制旧机制，这些已选行为对应的责任仍需落点；具体哪些复杂度来自未选功能应单独剔出。[边界文档][boundaries]、[替换实现][replace]、[迁移实现][migration]
- **失败后回旧包与日后撤销升级是两项承诺。**旧撤销窗口会被新版能力调用（包括可能写私有数据的读）、配置/会话变化或停用卸载结束；F19.4 撤销 GUI、F19.5 显式降级均未选，不能由 F19.3 的失败处理自动带入。[边界文档][boundaries]、[Checklist][checklist]
- **依赖检查与自动获取分开。**旧 registry 用可执行文件及 `--version` 检查范围，Node 另有按需下载路径；F02.6 的运行环境/认证检查不能自行扩展成自动下载 Agent/runtime。包本地安装、在线更新检查和已验证组合也不同；F24.7–F24.10 均未选。[registry][registry]、[支持矩阵][matrix]、[Checklist][checklist]

留给本票的实质选择是：非 Agent 能力包首个必要贡献是什么；升级能承诺转换哪一层状态、如何证明兼容；遇到不可恢复会话时如何阻止或显式降为仅历史。上述源码不能替用户选择“全部迁移成功才升级”、允许部分迁移或其他新机制。[范围与问题][decision]
验证：核对本地文档、清单和上述函数，检查 Markdown 相对目标与 diff；未执行旧测试、原生恢复或双平台探针。旧矩阵本身也区分 macOS arm64 证据与 Windows/Linux 未完成范围，不能外推为新版本验收。[支持矩阵][matrix]

[decision]: ../../.scratch/aibo-rev-architecture/issues/09-package-lifecycle.md
[ticket]: ../../.scratch/aibo-rev-architecture/issues/16-package-lifecycle-evidence.md
[agent-decision]: ../../.scratch/aibo-rev-architecture/issues/04-agent-capability-contract.md#answer
[presentation]: ../../.scratch/aibo-rev-architecture/issues/05-presentation-interface.md#answer
[checklist]: ../refactor-feature-inventory.md
[rev-readme]: ../../README.md
[host-readme]: ../../../aibo/README.md
[plugins-readme]: ../../../aibo-plugins/README.md
[greeting-readme]: ../../../aibo-plugins/plugins/capability/README.md
[greeting-manifest]: ../../../aibo-plugins/plugins/capability/plugin.json
[greeting-worker]: ../../../aibo-plugins/plugins/capability/worker.ts
[cursor-worker]: ../../../aibo-plugins/plugins/cursor/worker.mjs
[cursor-extension]: ../../../aibo-plugins/plugins/cursor/cursor-session.mjs
[terminal]: ../../../aibo-plugins/plugins/terminal/README.md
[development]: ../../../aibo/docs/plugin-development.md
[matrix]: ../../../aibo/docs/plugin-platform-support-matrix.md
[manifest]: ../../../aibo/contracts/plugin-manifest.v2.schema.json
[manifest-code]: ../../../aibo/src-tauri/src/plugin_manifest.rs
[write-fixture]: ../../../aibo/fixtures/plugins/semantic-write/plugin.json
[project-actions]: ../../../aibo/src-tauri/src/project_actions.rs
[host-tools]: ../../../aibo/docs/session-history-tool-design.md
[tool-view]: ../../../aibo/docs/tool-view-contract.md
[boundaries]: ../../../aibo/docs/plugin-boundaries-and-regression.md#插件卸载与会话迁移
[session-migration]: ../../../aibo/docs/capability-session-migration.md
[commands]: ../../../aibo/src-tauri/src/lib.rs
[registry]: ../../../aibo/src-tauri/src/plugin_registry.rs
[replace]: ../../../aibo/src-tauri/src/session_plugin_replacement.rs
[replacement]: ../../../aibo/src-tauri/src/plugin_replacement.rs
[migration]: ../../../aibo/src-tauri/src/session_plugin_lifecycle.rs
[binding]: ../../../aibo/contracts/session-binding.v2.schema.json
[acp-session]: ../../../aibo/packages/acp-adapter/session.mjs
[pi-engine]: ../../../aibo/src-tauri/capability-plugins/pi/engine.mjs
