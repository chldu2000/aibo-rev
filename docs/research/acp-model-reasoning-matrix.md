# ACP 下“每个模型支持的推理强度”的取得方式

研究日期：2026-11-12。起因：用户决定正式版本先用两个下拉选择模型与推理强度（[界面方案 §5.3](../design/ui-interaction.md)），并要求调研开源社区里有没有基于 ACP 的实现，能拿到 Agent 支持的每个模型及其推理强度，从而启用矩阵。本文只整理证据，不改变这个决定。

## 范围与证据等级

- **规范**：ACP 官方文档（v1 与 v2 的 Session Config Options、Session options RFD 草案）。
- **适配器源码**：读的是已发布的 npm 包，没有执行：`@agentclientprotocol/codex-acp` 2.6.0（`prototypes/acp-concurrency/adapter` 中已安装）、`@agentclientprotocol/claude-agent-acp` 0.106.0（`npm pack` 解包后只读）。
- **社区实现**：GitHub 上的公开 issue 与 PR，只读了页面描述，没有读它们的代码，也没有运行。
- **旧 Aibo**：`aibo` 仓库 `de70665`。

## 结论

1. **ACP 规范没有这项能力。** `thought_level` 是会话配置项的保留类别，只描述当前状态；规范的处理方式是在模型变化后，由 `session/set_config_option` 的响应或 `config_option_update` 送回新的完整配置。规范没有“某个选项的取值依赖另一个选项”的元数据，也没有 RFD 或 issue 提出过。[Session Config Options v2][S-v2]、[v1][S-v1]
2. **Session options RFD（草案，2026-11-03，11-10 修订）也帮不上。** 新增的 `session/options` 能在建会话前取得配置项，但参数只有 `cwd`、`additionalDirectories` 与可选的 `sessionId`，不能指定模型，所以拿到的仍是默认模型的推理强度。它对依赖的处理方式是：`session/new` 的 `configOptions` 按顺序应用，先选模型、再选推理强度。[Session options RFD][S-opts]
3. **两个正式支持的适配器内部都有完整映射，但只公布当前模型那一行。**
   - codex-acp 2.6.0：`availableModels[].supportedReasoningEfforts` 来自 Codex app-server 的 `model/list`；`createReasoningEffortConfigOption` 只用当前模型的列表生成 `reasoning_effort`（`category: thought_level`），换模型后重建该选项。
   - claude-agent-acp 0.106.0：`modelInfos[].supportedEffortLevels` 来自 Claude Agent SDK；`buildEffortConfigOption` 只取当前模型（`session-effort.js`），选项 id 为 `effort`，另有 `default`（交回 CLI 决定）。
   - 两者的 `_meta` 只用于一个 `air` 命名空间的扩展，例如推荐值 `recommendedValue`，不包含每个模型的推理强度。
4. **社区里只有两类补法，都没有成为标准。**
   - **逐个切换模型去探测。** bb 用一个专门的探测会话，对每个模型调用 `session/set_config_option`，再读回 `thought_level`；acp-bridge 的做法相同。代价：每次切换 115–150 ms，bb 的 5 秒上限只能覆盖约 40–80 个模型，没探到的模型显示假的“Medium”并被真正发给 Agent；只要一次探测失败，后面的模型全部退化。[bb #4367][B-4367]、[bb #4154][B-4154]、[acp-bridge #38][A-38]
   - **私有 `_meta` 扩展。** Lody 定义了 `_meta.lody.modelCapabilities.models[<model>].effortValues`，对应 PR 仍未合并；bb 在 oh-my-pi 的分叉中给每个模型选项加 `_meta.thoughtLevels`，说是“之后向上游提议”，目前没有看到上游合并。另有一个对 5 个 harness 的比较（Codex、Claude Code、pi、OpenCode、deepseek-harness），结论是全部只公布当前模型的列表。[Rx #73][R-73]
   - 其他客户端（OpenHands SDK、dagmar、SrudAgent、kule 等）都按规范只读当前模型的选项。Cursor 把推理强度编进模型 ID，每个模型只有一档。[OpenHands #6555][O-6555]、[dagmar #47][D-47]、[open-webui #34167][W-34167]
5. **旧 Aibo 能显示矩阵，是因为 Codex 没走 ACP。** 它直接调 Codex app-server 的 `model/list`（`src-tauri/capability-plugins/codex/engine.mjs`），目录缓存 60 秒；Claude Code 等 ACP Agent 声明 `parameterScope: "current-model"`，只能先选模型、再选推理强度。新版要求全部经 ACP 接入，这条路不能沿用。

## 对 Aibo Rev 的含义

- 正式版本的两个下拉，与社区里的通行做法一致，不需要维护任何扩展。
- **探测不可取。** 它会建额外的会话、改变会话配置，与界面方案 §5.3“宿主不得逐个切换模型来读取”冲突；Codex 与 Claude Code 是否把空会话写进各自的历史，也没有证据。
- **以后要启用矩阵，代价最小的路线是上游扩展。** 两个适配器内部已经有数据，在模型选项上加一个 `_meta`（例如每个取值附带支持的推理强度与默认值）改动很小；更好的是先在 ACP 提出“选项取值依赖”的通用元数据，避免每家一个私有键。这需要有人维护提议，按 [04 号决议](../../.scratch/aibo-rev-architecture/issues/04-agent-capability-contract.md#answer)记录扩展合同与维护责任，目前不在范围内。

## 来源

[S-v2]: https://agentclientprotocol.com/protocol/v2/session-config-options
[S-v1]: https://agentclientprotocol.com/protocol/v1/session-config-options
[S-opts]: https://agentclientprotocol.com/rfds/session-options
[B-4367]: https://github.com/get-bb/bb/issues/4367
[B-4154]: https://github.com/get-bb/bb/issues/4154
[A-38]: https://github.com/anesthetised/acp-bridge/pull/38
[R-73]: https://github.com/samzong/Rx/issues/73
[O-6555]: https://github.com/OpenHands/software-agent-sdk/issues/6555
[D-47]: https://github.com/srozov/dagmar/issues/47
[W-34167]: https://github.com/open-webui/open-webui/issues/34167

- [Session Config Options（v2）][S-v2]、[（v1）][S-v1]
- [Session options RFD][S-opts]
- [get-bb/bb #4367][B-4367]、[#4154][B-4154]
- [anesthetised/acp-bridge #38][A-38]
- [samzong/Rx #73][R-73]
- [OpenHands/software-agent-sdk #6555][O-6555]
- [srozov/dagmar #47][D-47]
- [open-webui/open-webui #34167][W-34167]
