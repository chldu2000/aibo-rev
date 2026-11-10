/* Aibo Rev 默认工作台原型 · 示例数据（全部为虚构样例，不代表真实 Agent 行为） */
const ICONS = {
  search: '<circle cx="11" cy="11" r="7"/><path d="m20 20-3.5-3.5"/>',
  plus: '<path d="M12 5v14M5 12h14"/>',
  settings: '<path d="M4 6h9M17 6h3M4 12h3M11 12h9M4 18h11M19 18h1"/><circle cx="15" cy="6" r="2"/><circle cx="9" cy="12" r="2"/><circle cx="17" cy="18" r="2"/>',
  sidebar: '<rect x="3" y="4" width="18" height="16" rx="2"/><path d="M9 4v16"/>',
  panel: '<rect x="3" y="4" width="18" height="16" rx="2"/><path d="M15 4v16"/>',
  down: '<path d="m6 9 6 6 6-6"/>',
  right: '<path d="m9 6 6 6-6 6"/>',
  left: '<path d="m15 6-6 6 6 6"/>',
  stop: '<rect x="6" y="6" width="12" height="12" rx="1.5"/>',
  pause: '<path d="M9 5v14M15 5v14"/>',
  play: '<path d="M7 5l12 7-12 7z"/>',
  send: '<path d="M12 19V5M5 12l7-7 7 7"/>',
  clip: '<path d="M21 11.5 12.5 20a5 5 0 0 1-7-7l8.5-8.5a3.3 3.3 0 0 1 4.7 4.7L10.2 17.7a1.7 1.7 0 0 1-2.4-2.4L15.5 7.6"/>',
  image: '<rect x="3" y="4" width="18" height="16" rx="2"/><circle cx="9" cy="10" r="2"/><path d="m21 16-5-5-9 9"/>',
  x: '<path d="M6 6l12 12M18 6 6 18"/>',
  check: '<path d="m5 12 5 5 9-10"/>',
  alert: '<path d="M12 3 2 20h20z"/><path d="M12 10v4M12 17h.01"/>',
  help: '<circle cx="12" cy="12" r="9"/><path d="M9.5 9.5a2.5 2.5 0 1 1 3.5 2.3c-.6.3-1 .9-1 1.6V14M12 17h.01"/>',
  clock: '<circle cx="12" cy="12" r="9"/><path d="M12 7v5l3 2"/>',
  target: '<circle cx="12" cy="12" r="9"/><circle cx="12" cy="12" r="5"/><circle cx="12" cy="12" r="1"/>',
  bot: '<rect x="4" y="8" width="16" height="12" rx="2"/><path d="M12 4v4M9 13h.01M15 13h.01"/>',
  term: '<rect x="3" y="4" width="18" height="16" rx="2"/><path d="m7 9 3 3-3 3M13 15h4"/>',
  more: '<circle cx="5" cy="12" r="1"/><circle cx="12" cy="12" r="1"/><circle cx="19" cy="12" r="1"/>',
  sun: '<circle cx="12" cy="12" r="4"/><path d="M12 2v2M12 20v2M2 12h2M20 12h2M5 5l1.5 1.5M17.5 17.5 19 19M5 19l1.5-1.5M17.5 6.5 19 5"/>',
  moon: '<path d="M20 14.5A8 8 0 1 1 9.5 4a6.5 6.5 0 0 0 10.5 10.5z"/>',
  archive: '<rect x="3" y="4" width="18" height="4" rx="1"/><path d="M5 8v11a1 1 0 0 0 1 1h12a1 1 0 0 0 1-1V8M10 12h4"/>',
  folder: '<path d="M3 7a2 2 0 0 1 2-2h4l2 2h8a2 2 0 0 1 2 2v8a2 2 0 0 1-2 2H5a2 2 0 0 1-2-2z"/>',
  file: '<path d="M14 3H6a1 1 0 0 0-1 1v16a1 1 0 0 0 1 1h12a1 1 0 0 0 1-1V8z"/><path d="M14 3v5h5"/>',
  refresh: '<path d="M20 12a8 8 0 1 1-2.3-5.7L20 8M20 4v4h-4"/>',
  lock: '<rect x="5" y="11" width="14" height="10" rx="2"/><path d="M8 11V7a4 4 0 0 1 8 0v4"/>',
  hash: '<path d="M5 9h14M5 15h14M10 3 8 21M16 3l-2 18"/>',
  at: '<circle cx="12" cy="12" r="4"/><path d="M16 8v5a3 3 0 0 0 6 0v-1a10 10 0 1 0-4 8"/>',
  slash: '<path d="M16 4 8 20"/>',
  copy: '<rect x="8" y="8" width="13" height="13" rx="2"/><path d="M16 8V5a2 2 0 0 0-2-2H5a2 2 0 0 0-2 2v9a2 2 0 0 0 2 2h3"/>',
  quote: '<path d="M7 7h10M7 12h10M7 17h6"/>',
  spark: '<path d="M12 3v4M12 17v4M3 12h4M17 12h4M6 6l2.5 2.5M15.5 15.5 18 18M6 18l2.5-2.5M15.5 8.5 18 6"/>',
  tool: '<path d="M14.7 6.3a4 4 0 0 0-5.4 5.4L3 18l3 3 6.3-6.3a4 4 0 0 0 5.4-5.4l-2.6 2.6-2.4-.6-.6-2.4z"/>',
  shield: '<path d="M12 3 4 6v6c0 5 3.5 8 8 9 4.5-1 8-4 8-9V6z"/>',
  edit: '<path d="M4 20h4L19 9l-4-4L4 16z"/>',
  trash: '<path d="M4 7h16M9 7V4h6v3M6 7l1 13h10l1-13"/>',
  bolt: '<path d="M13 2 4 14h7l-1 8 9-12h-7z"/>',
  grip: '<circle cx="9" cy="6" r="1"/><circle cx="15" cy="6" r="1"/><circle cx="9" cy="12" r="1"/><circle cx="15" cy="12" r="1"/><circle cx="9" cy="18" r="1"/><circle cx="15" cy="18" r="1"/>',
  list: '<path d="M8 6h13M8 12h13M8 18h13M3 6h.01M3 12h.01M3 18h.01"/>',
  branch: '<circle cx="6" cy="5" r="2"/><circle cx="6" cy="19" r="2"/><circle cx="18" cy="7" r="2"/><path d="M6 7v10M18 9c0 5-7 4-11 8"/>',
  pin: '<path d="M12 16v6M8 3h8l-1 6 3 4H6l3-4z"/>',
  info: '<circle cx="12" cy="12" r="9"/><path d="M12 11v6M12 7.5h.01"/>',
  layers: '<path d="m12 3 9 5-9 5-9-5z"/><path d="m3 13 9 5 9-5"/>',
};

const AGENTS = {
  claude: {
    name: 'Claude Code', short: 'CC', color: '#C46A47', adapter: 'claude-agent-acp 0.9.2', verified: true, ok: true,
    steer: false, steerWhy: '此 Agent/适配器组合的插话存在回退歧义，尚未启用。可以排队，或停止后发送。',
    goal: true, goalPause: false, goalPauseWhy: '此 Agent 只提供设置与清除目标，不支持暂停/恢复。',
    models: ['Claude Opus 5.5', 'Claude Sonnet 5.5', 'Claude Haiku 5.5'], reasoning: ['低', '中', '高'],
    fast: false, fastWhy: '此 Agent 未提供服务档位选项。', modes: ['询问', '自动编辑', '计划', '完全访问'], fullRead: true,
  },
  codex: {
    name: 'Codex', short: 'CX', color: '#2F7A6A', adapter: 'codex-acp 0.4.1', verified: true, ok: true,
    steer: true, goal: true, goalPause: true,
    models: ['gpt-5.5-codex', 'gpt-5.5'], reasoning: ['低', '中', '高', '极高'], fast: true,
    modes: ['只读', '询问', '自动', '完全访问'], fullRead: true,
  },
  gemini: {
    name: 'Gemini CLI', short: 'GM', color: '#4C6FD6', adapter: '通用 ACP 模板', verified: false, ok: true,
    steer: false, steerWhy: '此 Agent 未声明插话能力。', goal: false, goalWhy: '此 Agent 未提供持续目标能力。',
    goalPause: false, goalPauseWhy: '此 Agent 未提供持续目标能力。',
    models: ['gemini-3-pro'], reasoning: [], reasoningWhy: '此 Agent 未提供推理强度选项。',
    fast: false, fastWhy: '此 Agent 未提供服务档位选项。', modes: ['默认', '自动'], fullRead: false,
  },
  opencode: {
    name: 'OpenCode', short: 'OC', color: '#777', adapter: '通用 ACP 模板', verified: false, ok: false,
    okWhy: '未找到可执行文件 opencode（已检查 PATH 与 ~/.local/bin）。',
  },
};

const FILES = [
  'src/core/session-control.ts', 'src/core/queue.ts', 'src/core/delivery-evidence.ts', 'src/core/',
  'src/presentation/shared/draft.ts', 'src/presentation/shared/', 'src/presentation/default/',
  'docs/design/ui-interaction.md', 'docs/refactor-feature-inventory.md', 'package.json', 'go.mod', 'cmd/aibo/main.go',
];

const COMMANDS = [
  { n: '/goal', d: '设置持续目标', src: 'Aibo' },
  { n: '/review', d: '审查当前工作区改动', src: 'Agent' },
  { n: '/compact', d: '压缩上下文', src: 'Agent' },
  { n: '/init', d: '生成 AGENTS.md', src: 'Agent' },
  { n: '/skill:write-tests', d: '为改动补充测试', src: 'Skill' },
  { n: '/skill:grilling', d: '逐题追问并记录决议', src: 'Skill' },
];

const MD1 = `<p>共享交互模块的草稿状态已经从组件里移出来了。现在的结构是：</p>
<ul><li><code>DraftStore</code> 以会话 ID 为键保存正文、附件与固定引用，输入停顿 400ms 后合并写入；</li>
<li>呈现包只订阅 <code>draft.snapshot</code> 与 <code>draft.changed</code>，不持有权威状态；</li>
<li>重挂载时先读快照，再按游标接续增量。</li></ul>
<div class="codeblock"><div class="ch"><span>ts</span><button class="ib" data-a="copy" aria-label="复制代码"><svg class="i sm" viewBox="0 0 24 24">${ICONS.copy}</svg></button></div><pre>export interface DraftSnapshot {
  sessionId: SessionId
  text: string
  attachments: AttachmentRef[]
  references: PinnedReference[]
  savedVersion: number
}</pre></div>
<p>接下来我准备跑一下 core 包的测试，确认队列与草稿之间的保留关系没有断开。</p>`;

const MD3 = `<p>我查了一下缩放问题，原因是 WebView2 在每显示器 DPI 切换时没有重新计算客户区。可以有三种修法，差别主要在改动范围：</p>
<ol><li>在窗口的 <code>WM_DPICHANGED</code> 里主动重设 WebView 边界；</li>
<li>升级 MyGO 到 v0.7.4，它修了一部分首帧缩放问题；</li>
<li>两者都做。</li></ol>`;

function seed() {
  return {
    workspaces: [
      { id: 'rev', name: 'aibo-rev', path: '~/Workspace/aibo-dev/aibo-rev', trusted: true, open: true },
      { id: 'mygo', name: 'mygo-playground', path: '~/Workspace/mygo-playground', trusted: false, open: true },
    ],
    sessions: [
      {
        id: 's1', ws: 'rev', title: '重构共享交互模块的草稿状态', agent: 'claude',
        model: 'Claude Opus 5.5', reasoning: '高', fast: false, mode: '询问', ctx: [76, 200],
        running: true, startedAt: Date.now() - 134000, updated: '刚刚', pinned: true,
        queue: [{ id: 'q11', text: '完成后顺便补上快照游标的边界测试，尤其是缺口重同步那条路径。', chips: [], at: '14:41' }],
        queuePaused: null, goal: null,
        subagents: [{ id: 'sa1', name: '检查快照协议的测试覆盖', status: 'running', bg: false, since: '1 分钟',
          items: [{ t: 'user', text: '列出 snapshot/cursor 相关的测试，并指出没有覆盖的分支。', d: 'received' },
                  { t: 'tool', name: 'Grep', target: 'cursor|snapshot  src/**/*.test.ts', st: 'done', dur: '0.4s', out: '12 处匹配，分布在 4 个文件' },
                  { t: 'agent', html: '<p>目前缺少“游标缺口后重新同步”的测试，正在继续检查队列部分…</p>' }] }],
        tasks: [],
        items: [
          { t: 'user', text: '把共享交互模块里的草稿状态从 Svelte store 里移出来，改成框架无关的实现。呈现包只能订阅，不能持有权威状态。', d: 'received', chips: [{ type: 'file', label: 'src/presentation/shared/' }] },
          { t: 'thinking', secs: 12, text: '需要先确认草稿目前在哪些组件里被直接修改，再看保存时序与队列转移的关系……' },
          { t: 'group', label: '读取了 4 个文件', items: ['src/presentation/shared/draft.ts', 'src/presentation/default/Composer.svelte', 'src/core/queue.ts', 'docs/design/ui-interaction.md'] },
          { t: 'tool', name: 'Edit', target: 'src/presentation/shared/draft.ts', st: 'done', dur: '1.2s', out: '+84 −37' },
          { t: 'agent', html: MD1 },
          { t: 'subagent', ref: 'sa1' },
          { t: 'approval', id: 'ap1', tool: 'Bash', cmd: 'pnpm test --filter @aibo/core', cwd: '~/Workspace/aibo-dev/aibo-rev', options: ['允许一次', '本会话始终允许', '拒绝'], state: 'pending' },
        ],
      },
      {
        id: 's2', ws: 'rev', title: '队列与投递证据', agent: 'codex',
        model: 'gpt-5.5-codex', reasoning: '高', fast: true, mode: '自动', ctx: [118, 272],
        running: false, updated: '6 分钟前',
        queue: [
          { id: 'qu', unknown: true, text: '把 promptRequired 的歧义结果按“未知”处理，不要自动转普通投递。', chips: [], at: '14:22', why: '连接在投递过程中断开，无法确认 Agent 是否已接收。' },
          { id: 'q21', text: '再检查一遍取消后队列暂停的状态迁移。', chips: [], at: '14:24' },
          { id: 'q22', text: '整理成测试清单发给我。', chips: [], at: '14:25' },
        ],
        queuePaused: '投递结果未知，队列已暂停',
        goal: { text: '让队列相关的 12 个失败用例全部通过', status: 'active', since: '13:58', budget: [['Token', '58k / 200k', 'Agent 报告'], ['时长', '未报告', ''], ['回合', '7', 'Agent 报告']] },
        subagents: [],
        tasks: [{ id: 't1', cmd: 'cargo test -p queue --no-fail-fast', status: 'running', since: '14:19', last: '刚刚',
          out: 'running 48 tests\ntest delivery::unknown_is_not_resent ... ok\ntest delivery::steer_fallback_once ... ok\ntest queue::pause_on_cancel ... FAILED\ntest queue::resume_explicit ... ok\n' }],
        items: [
          { t: 'user', text: '给投递尝试加上独立 ID，区分“已开始投递但没有结果”和“确认未消费”。', d: 'received' },
          { t: 'tool', name: 'Shell', target: 'rg "DeliveryAttempt" -n src', st: 'done', dur: '0.3s', out: 'src/core/delivery-evidence.ts:12\nsrc/core/queue.ts:88' },
          { t: 'agent', html: '<p>已拆出 <code>DeliveryAttempt</code>，状态分为 <code>not_started</code> / <code>started</code> / <code>unconsumed_verified</code> / <code>accepted</code> / <code>settled</code> / <code>unknown</code>。超时与断连只会进入 <code>unknown</code>，不会退回 <code>not_started</code>。</p>' },
          { t: 'user', text: '把 promptRequired 的歧义结果按“未知”处理，不要自动转普通投递。', d: 'unknown', qid: 'qu' },
        ],
      },
      {
        id: 's3', ws: 'rev', title: '修复 Windows 窗口缩放', agent: 'codex',
        model: 'gpt-5.5-codex', reasoning: '中', fast: false, mode: '询问', ctx: [41, 272],
        running: false, updated: '22 分钟前', queue: [], queuePaused: null, goal: null, subagents: [], tasks: [],
        items: [
          { t: 'user', text: '在 4K + 1080p 双屏之间拖动窗口时，WebView 内容会被裁掉一块。查一下原因。', d: 'received' },
          { t: 'group', label: '读取了 3 个文件', items: ['internal/windows/webview.go', 'internal/windows/window.go', 'go.mod'] },
          { t: 'agent', html: MD3 },
          { t: 'question', id: 'qq1', title: '采用哪种修复方案？', options: ['只处理 WM_DPICHANGED', '只升级 MyGO 到 v0.7.4', '两者都做'], extra: '补充说明（可选）', state: 'pending', sel: null },
        ],
      },
      {
        id: 's4', ws: 'rev', title: '整理数据生命周期盘点', agent: 'claude',
        model: 'Claude Sonnet 5.5', reasoning: '中', fast: false, mode: '自动编辑', ctx: [23, 200],
        running: false, failed: true, updated: '1 小时前',
        queue: [{ id: 'q41', text: '然后把 D04 和 D06 的验收问题合并成一张表。', chips: [], at: '13:02' }],
        queuePaused: '回合失败，队列已暂停', goal: null, subagents: [], tasks: [],
        items: [
          { t: 'user', text: '把 data-lifecycle-inventory.md 里每一项对应到 10 号决议的数据所有者。', d: 'received', chips: [{ type: 'file', label: 'docs/data-lifecycle-inventory.md' }] },
          { t: 'tool', name: 'Read', target: 'docs/data-lifecycle-inventory.md', st: 'done', dur: '0.1s', out: '75 行' },
          { t: 'gap', text: '此处有 2 条工具输出未能保存（应用在保存前中断）' },
          { t: 'error', msg: 'Agent 返回错误：上游服务过载（HTTP 529），回合已结束。', detail: 'ACP session/prompt → stopReason: "refusal"\n_meta.error: { code: "overloaded_error", status: 529, request_id: "req_01HX…" }' },
        ],
      },
      {
        id: 's5', ws: 'mygo', title: 'MyGO 原生菜单试验', agent: 'gemini',
        model: 'gemini-3-pro', reasoning: '', fast: false, mode: '默认', ctx: null,
        running: false, updated: '昨天', queue: [], queuePaused: null, goal: null, subagents: [], tasks: [],
        items: [
          { t: 'user', text: 'WebView 卡死时，原生菜单还能响应吗？', d: 'received' },
          { t: 'agent', html: '<p>MyGO 的菜单事件在 Go 主线程处理，不经过页面 JavaScript。只要宿主进程没有阻塞，WebView 内的死循环不会让菜单失去响应。这仍需要在 macOS 和 Windows 上实际注入故障来确认。</p>' },
        ],
      },
      {
        id: 's6', ws: 'rev', title: '旧终端路线讨论', agent: 'claude', archived: true,
        model: 'Claude Opus 5.5', reasoning: '高', fast: false, mode: '询问', ctx: [12, 200],
        running: false, updated: '11-01', queue: [], queuePaused: null, goal: null, subagents: [], tasks: [],
        items: [
          { t: 'user', text: '如果只保留跨 Agent 引用，能不能直接用 Ghostty 跑各家 CLI？', d: 'received' },
          { t: 'agent', html: '<p>可以把终端交互与会话共享分开：原生 CLI 通过 PTY 运行，会话桥从可靠的原生记录获取历史……</p>' },
        ],
      },
    ],
  };
}
