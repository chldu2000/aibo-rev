# Windows 11 x64 验收报告（17 号票）

2026-11-12；自动证据与用户人工验收通过，17 号票已收口。源码基线 `930240bb02616497edca8b20afb94a06664424fd`（2026-11-15 为移除误提交的二进制改写了历史，该提交现为树完全相同的 `5522151c34a31de3cb1494631b1942b1f5c1525c`；`win-*.json` 中的 `sourceCommit` 保留实测时的原值），实验包含本次 Windows 适配；具体产物与源码哈希见 `win-build-manifest.json`。环境见 `win-environment.json`。Go 1.27.3、MyGO v0.9.0、CGO_ENABLED=0。

## 已有实机证据

- `../mygo-validation/reports/win-children-result.json`：专用宿主加入带 KILL_ON_JOB_CLOSE 的 Job Object 后，启动读 stdin、不读 stdin、派生孙进程的子进程。脚本先核实四个进程均存活，再分别使用 Stop-Process -Force、taskkill /F、等待正常退出；三组均无残留。stdin 管道由宿主持有，避免 GC 提前关闭。
- `win-renderer-destroy.json`、`win-renderer-hide.json`：真实可见窗口中面板宽度与窗口尺寸同步、收起后页面保留且计时器被节流。单 renderer 死循环时工作台继续流式更新；结束 renderer 后 Reload 恢复。关闭窗口后宿主持有的会话、草稿、面板宽度、滚动条目恢复。
- `win-renderer-destroy.json`：108 帧，发布到构建平均 3.982 ms，最大 8.043 ms；主线程投递平均 0.200 ms，最大 1.498 ms。panic 注入后侧栏与工具卡局部显示故障，其他构建继续。
- `../native-ui/reports/win-auto.json`：流式 107 帧，发布到构建平均 3.327 ms、最大 8 ms；加载全部历史 10,004 条。无窗口测试另行通过，不作为人工 IME 证据。
- `../native-ui/reports/win-memory.json` 与 `win-memory-auto.json`：另一次同场景运行采样 65 次，Working Set 峰值 86,851,584 字节（约 82.8 MiB），最终加载 10,004 条，进程以 0 退出。这是本机进程的内存采样，不与 macOS 的统计直接排名。
- `../mygo-validation/reports/win-two-panels-recreate.json`：两个包 origin 不同，均为安全上下文、可调用 Go；共享浏览器、独立 renderer。A 死循环 12 秒期间 B 可用；本次没有观察到 unresponsive 回调。结束 A renderer 后 B 仍可用、A Reload 可恢复。结束浏览器后两面板均失效且 Reload 失败；销毁并重建 WebView 后均恢复。
- `win-visible-destroy.json` 与对应 `.events.txt`：保留了结束浏览器进程导致 Reload 失败的原始样本。`win-auto-destroy.json` 是最初隐藏启动的实验，未产生有效绘制帧，不能作为界面或性能通过证据。
- 无响应直接重建实验已完成两次（`win-loop-recreate.json`、`win-loop-recreate-binding.json` 位于 mygo-validation/reports）：创建调用约 2 ms；2 秒后的探测确认原 renderer 已退出，1 秒后再次确认；另一包面板不受影响。第二次还记录两面板重建前后的 Go 绑定调用均返回 ok。用户已确认采用 ADR 0008 的恢复策略。

## 帧耗时

真实窗口，MYGO_FRAME_STATS=all。日志中的 PERF-BEGIN/END 区分流式与滚动阶段，按帧耗时排序；中位数取排序后中间位置的较低值，p95 取 ceil(N×0.95)。启动单列，未混入稳态统计。GPU 的具体适配器选择尚未从运行时证明；本机 GPU 列表不等于 D3D11 实际选择。

| 绘制/外观 | 流式中位/p95/最大 ms | 滚动中位/p95/最大 ms | 启动最大 ms |
| --- | --- | --- | --- |
| GPU 标准 | 1.0 / 1.5 / 4.5 | 1.0 / 1.5 / 2.5 | 73.0 |
| GPU 玻璃 | 1.5 / 2.0 / 5.5 | 1.5 / 2.5 / 3.5 | 507.7 |
| MYGO_GPU=0 强制 CPU 玻璃 | 20.0 / 21.5 / 27.5 | 20.0 / 21.0 / 23.1 | 73.0 |

数据：`win-perf-summary.json`、`win-perf-cpu-summary.json`；原始合并日志为 `win-perf-*.txt`。CPU 结果是本机强制关闭 MyGO GPU 绘制的实测，不代表另一台无 GPU 机器的表现。CPU 日志未出现 drawn on the GPU。

## 构建与检查

- 外壳与 native-ui 的 go test 通过；mygo-validation 构建、go test（无测试文件）及 go vet 通过；外壳 go vet 通过。
- go -C tools/ui-checks run . all 通过。Windows CRLF 检出曾导致 token-sync 失败，按要求运行 tokens 生成器后通过；令牌内容未变。
- go tool mygo build -platform windows/amd64 成功：原型 exe 约 13.4 MB，NSIS 安装程序约 4.3 MB。通知点击定位已由用户人工确认通过；用户未单独说明是否使用安装版，不声称已执行安装程序。
- 本轮最终 go vet、go test 输出位于三个原型各自的 `reports/win-vet.txt`、`win-tests.txt`；界面门槛输出为 `win-ui-checks.txt`。Windows 调整后的 mygo-validation 也已交叉编译 darwin/arm64，未将其视为 macOS 实测。
- `win-build-manifest.json` 固定基线提交、未提交改动事实与每个 Windows exe/安装程序的 SHA-256；独立实验分支为 `prototype/windows-validation`。`win-manual-process-tree.json` 为人工检查窗口所属的实际 WebView2 子树；`win-input-methods.json` 为系统配置的输入法原始记录，不代表已做输入法实测。
- 新增直接依赖 golang.org/x/sys v0.48.0（已由 MyGO 间接依赖）：用于 Job Object 与进程状态的 Windows API，保持无 cgo。
- 人工夹具补齐了“验证 → 打开附件文件选择器”，原型此前只提供目录选择器、粘贴和拖放。新版为 `bin/shell-next.exe`，独立打包目录为 `dist/manual-next/windows-amd64/`；文件/目录选择器随人工清单由用户确认通过。

## 人工验收与决议

2026-11-12 用户确认：“人工验证通过；ADR 0008 恢复策略可以按照建议修订”。据此，17 号票中的字体与缩放、原生与面板输入法、粘贴与拖放、文件/目录选择器、深浅色、标题栏与贴靠、焦点往返与快捷键、托盘、通知点击定位、讲述人、Mica 与透明效果关闭等人工条款通过。证据来源为用户实际操作反馈；自动报告不替代这些人工结果。

ADR 0008 已接受，并将无响应面板直接重建、共享浏览器退出时重建全部受影响面板的规则回写 ADR 0005。17 号票、地图、ADR 0003/0005/0006/0007 和界面方案 §15 已同步。

## 适用范围与保留限制

- 结论适用于本次 Windows 11 x64、锁定工具链与 WebView2 组合；不扩展为 Windows 10、ARM64 或所有 GPU/驱动的正式支持证明。
- 用户未单列第三方输入法名称/版本、显示器组合和通知安装方式；按其对人工清单的整体确认记录通过，不补写未知环境详情。第二显示器和无 GPU 虚拟机是条件项；CPU 路径采用本机 MYGO_GPU=0 的独立实测。
- 首帧玻璃最大 507.7 ms 与 CPU 约 20 ms/帧保留在报告中。恢复时间中的约 2 ms 是销毁/创建调用耗时，页面可用性是在 2 秒后探测，不能当作完整恢复延迟。
- 原型不是产品实现。真实 Agent、存储故障、产品接线与完整端到端验收继续由 12 号票承接；产品长历史前插锚定和完整焦点/选区恢复不由本票推定完成。
