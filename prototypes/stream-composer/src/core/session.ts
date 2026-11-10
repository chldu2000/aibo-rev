/**
 * 与框架无关的会话模型（原型）。
 * 只用普通 TypeScript：任何 UI 框架都通过 subscribe/getSnapshot 读取，并调用方法表达意图。
 * Agent 行为为模拟；“保存后批量发布”用 publishInterval 合并增量。
 */

export type Delivery = 'queued' | 'delivering' | 'received' | 'unknown';

export interface Chip {
  id: string;
  type: 'file' | 'ref' | 'image';
  label: string;
  range?: '近 3 轮' | '全文';
  at?: string;
  chars?: number;
  stale?: boolean;
}

export interface UserItem { kind: 'user'; id: string; text: string; chips: Chip[]; delivery: Delivery; steer?: boolean }
export interface AgentItem { kind: 'agent'; id: string; text: string; done: boolean }
export interface ToolItem { kind: 'tool'; id: string; name: string; target: string; status: 'running' | 'done' | 'failed'; output: string; dur?: string }
export interface ApprovalItem { kind: 'approval'; id: string; tool: string; cmd: string; options: string[]; state: 'pending' | 'submitted' | 'resolved'; choice?: string }
export type Item = UserItem | AgentItem | ToolItem | ApprovalItem;

export interface QueueItem { id: string; text: string; chips: Chip[]; at: string }

export interface Settings { model: string; mode: string; applying: readonly string[] }

export interface Snapshot {
  readonly version: number;
  readonly publishedAt: number;
  readonly items: readonly Item[];
  readonly olderRemaining: number;
  readonly running: boolean;
  readonly queue: readonly QueueItem[];
  readonly queuePaused: string | null;
  readonly draft: string;
  readonly chips: readonly Chip[];
  /** 共通呈现状态：按稳定条目 ID 记录的展开状态 */
  readonly expanded: ReadonlySet<string>;
  readonly settings: Settings;
}

let seq = 0;
const uid = (p: string) => `${p}${++seq}`;
const now = () => new Date().toTimeString().slice(0, 5);

export class Session {
  /** 流式增量的发布间隔（ms）；0 表示每次立即发布 */
  publishInterval = 50;
  /** 模拟流式时每 16ms 追加的字符数 */
  streamRate = 6;

  #items: Item[] = [];
  #older: Item[] = [];
  #queue: QueueItem[] = [];
  #queuePaused: string | null = null;
  #running = false;
  #draft = '';
  #chips: Chip[] = [];
  #expanded = new Set<string>();
  #settings: Settings = { model: 'Claude Opus 5.5', mode: '询问', applying: [] };
  #timers = new Set<ReturnType<typeof setTimeout>>();
  #listeners = new Set<() => void>();
  #flushTimer: ReturnType<typeof setTimeout> | 0 = 0;
  #version = 0;
  #snap: Snapshot;

  constructor() {
    this.#chips = [{ id: uid('c'), type: 'ref', label: '队列与投递证据', range: '近 3 轮', at: '14:32', chars: 9800, stale: true }];
    this.#snap = this.#build();
  }

  subscribe(fn: () => void): () => void {
    this.#listeners.add(fn);
    return () => this.#listeners.delete(fn);
  }
  getSnapshot(): Snapshot { return this.#snap; }

  // ---------- publish ----------
  #build(): Snapshot {
    return {
      version: ++this.#version,
      publishedAt: performance.now(),
      items: this.#items.slice(),
      olderRemaining: this.#older.length,
      running: this.#running,
      queue: this.#queue.slice(),
      queuePaused: this.#queuePaused,
      draft: this.#draft,
      chips: this.#chips.slice(),
      expanded: new Set(this.#expanded),
      settings: { ...this.#settings },
    };
  }
  #flush() {
    if (this.#flushTimer) clearTimeout(this.#flushTimer);
    this.#flushTimer = 0;
    this.#snap = this.#build();
    for (const fn of this.#listeners) fn();
  }
  /** 结构性变化立即发布；流式正文按 publishInterval 合并发布 */
  #commit(batched = false) {
    if (!batched || this.publishInterval <= 0) return this.#flush();
    if (!this.#flushTimer) this.#flushTimer = setTimeout(() => this.#flush(), this.publishInterval);
  }
  #later(ms: number, fn: () => void) {
    const t = setTimeout(() => { this.#timers.delete(t); fn(); }, ms);
    this.#timers.add(t);
  }

  // ---------- history ----------
  seedHistory(total: number, visible = 60) {
    const all = generateHistory(total);
    this.#older = all.slice(0, Math.max(0, total - visible));
    this.#items = all.slice(Math.max(0, total - visible));
    this.#items.push(
      { kind: 'user', id: uid('u'), text: '把共享交互模块里的草稿状态移出 UI 框架，呈现包只能订阅。', chips: [], delivery: 'received' },
      { kind: 'tool', id: uid('t'), name: 'Edit', target: 'src/presentation/shared/draft.ts', status: 'done', output: '+84 −37', dur: '1.2s' },
      { kind: 'approval', id: uid('a'), tool: 'Bash', cmd: 'pnpm test --filter @aibo/core', options: ['允许一次', '本会话始终允许', '拒绝'], state: 'pending' },
    );
    this.#running = true;
    this.#commit();
  }
  /** 向前加载更早的记录（会话流顶部） */
  loadOlder(count = 80) {
    if (!this.#older.length) return;
    const chunk = this.#older.splice(Math.max(0, this.#older.length - count), count);
    this.#items = chunk.concat(this.#items);
    this.#commit();
  }
  loadAllHistory() {
    this.#items = this.#older.concat(this.#items);
    this.#older = [];
    this.#commit();
  }

  // ---------- expanded / draft / chips (shared presentation state) ----------
  toggleExpanded(id: string) {
    if (this.#expanded.has(id)) this.#expanded.delete(id); else this.#expanded.add(id);
    this.#commit();
  }
  setDraft(text: string) { this.#draft = text; this.#commit(true); }
  addChip(c: Omit<Chip, 'id'>) { this.#chips.push({ ...c, id: uid('c') }); this.#commit(); }
  updateChip(id: string, patch: Partial<Chip>) {
    this.#chips = this.#chips.map((c) => (c.id === id ? { ...c, ...patch } : c));
    this.#commit();
  }
  removeChip(id: string) { this.#chips = this.#chips.filter((c) => c.id !== id); this.#commit(); }
  popChip() { this.#chips.pop(); this.#commit(); }

  // ---------- settings ----------
  applySetting(key: 'model' | 'mode', value: string) {
    this.#settings = { ...this.#settings, applying: [...this.#settings.applying, key] };
    this.#commit();
    this.#later(800, () => {
      this.#settings = { ...this.#settings, [key]: value, applying: this.#settings.applying.filter((k) => k !== key) };
      this.#commit();
    });
  }

  // ---------- input & control ----------
  /** 执行中或队列暂停时进入队列，否则立即投递 */
  send(text: string) {
    const chips = this.#chips;
    if (!text && !chips.length) return;
    this.#chips = [];
    this.#draft = '';
    if (this.#running || (this.#queuePaused && this.#queue.length)) {
      this.#queue.push({ id: uid('q'), text: text || '（仅附件）', chips, at: now() });
      this.#commit();
      return;
    }
    this.#deliver(text, chips);
  }
  steer(text: string) {
    const it: UserItem = { kind: 'user', id: uid('u'), text, chips: [], delivery: 'delivering', steer: true };
    this.#items.push(it);
    this.#draft = '';
    this.#commit();
    this.#later(400, () => { this.#replace(it.id, { delivery: 'received' }); this.#commit(); });
  }
  stop() {
    for (const t of this.#timers) clearTimeout(t);
    this.#timers.clear();
    this.#items = this.#items.map((i) =>
      i.kind === 'agent' && !i.done ? { ...i, text: i.text + '\n\n*（已停止）*', done: true }
      : i.kind === 'tool' && i.status === 'running' ? { ...i, status: 'failed', output: '已随回合停止' }
      : i.kind === 'approval' && i.state === 'pending' ? { ...i, state: 'resolved', choice: '已失效' } : i);
    this.#running = false;
    if (this.#queue.length) this.#queuePaused = '已停止当前执行，队列已暂停';
    this.#commit();
  }
  approve(id: string, choice: string) {
    this.#replace(id, { state: 'submitted', choice });
    this.#commit();
    this.#later(500, () => {
      this.#replace(id, { state: 'resolved' });
      if (choice === '拒绝') { this.#endTurn(); return; }
      this.#runTurn();
    });
  }
  pauseQueue() { this.#queuePaused = '已手动暂停'; this.#commit(); }
  resumeQueue() { this.#queuePaused = null; this.#commit(); this.#next(); }
  removeQueued(id: string) { this.#queue = this.#queue.filter((q) => q.id !== id); this.#commit(); }
  sendQueuedNow(id: string) {
    const q = this.#queue.find((x) => x.id === id);
    if (!q) return;
    this.#queue = this.#queue.filter((x) => x !== q);
    if (this.#running) this.steer(q.text); else this.#deliver(q.text, q.chips);
  }
  /** 立即开始一次长回复（用于测流式性能） */
  startLongReply() {
    if (this.#running) this.stop();
    this.#deliver('请完整说明草稿状态的保存时序，并给出接口和测试示例。', []);
  }

  // ---------- simulation ----------
  #replace(id: string, patch: Partial<Item>) {
    const i = this.#items.findIndex((x) => x.id === id);
    if (i >= 0) this.#items[i] = { ...this.#items[i], ...patch } as Item;
  }
  #deliver(text: string, chips: Chip[]) {
    const it: UserItem = { kind: 'user', id: uid('u'), text, chips, delivery: 'delivering' };
    this.#items.push(it);
    this.#running = true;
    this.#commit();
    this.#later(400, () => { this.#replace(it.id, { delivery: 'received' }); this.#commit(); this.#runTurn(); });
  }
  #runTurn() {
    this.#running = true;
    const tool: ToolItem = { kind: 'tool', id: uid('t'), name: 'Read', target: 'src/core/queue.ts', status: 'running', output: '' };
    this.#items.push(tool);
    this.#commit();
    this.#later(600, () => {
      this.#replace(tool.id, { status: 'done', output: '214 行', dur: '0.2s' });
      const agent: AgentItem = { kind: 'agent', id: uid('m'), text: '', done: false };
      this.#items.push(agent);
      this.#commit();
      let pos = 0;
      const tick = () => {
        pos = Math.min(LONG_REPLY.length, pos + this.streamRate);
        this.#replace(agent.id, { text: LONG_REPLY.slice(0, pos) });
        if (pos < LONG_REPLY.length) { this.#commit(true); this.#later(16, tick); return; }
        this.#replace(agent.id, { done: true });
        this.#endTurn();
      };
      this.#later(16, tick);
    });
  }
  #endTurn() {
    this.#running = false;
    this.#commit();
    this.#later(200, () => this.#next());
  }
  #next() {
    if (this.#running || this.#queuePaused || !this.#queue.length) return;
    const q = this.#queue.shift()!;
    this.#deliver(q.text, q.chips);
  }
}

// ---------- sample content ----------
const LONG_REPLY = `草稿状态的保存分三步：编辑时合并、发送前确认、转入队列时转移保留关系。

## 1. 编辑时合并

输入停顿 400ms 后合并写入一次；\`DraftStore\` 只记录**当前编辑值**与**已保存版本**，不把每次按键都称为已持久化。

- 呈现包只订阅 \`draft.snapshot\` 与 \`draft.changed\`
- 重挂载时先读快照，再按游标接续增量
- 切换会话不触发提交

\`\`\`ts
export interface DraftSnapshot {
  sessionId: SessionId
  text: string
  attachments: AttachmentRef[]
  references: PinnedReference[]
  savedVersion: number
}

export function confirmBeforeSend(store: DraftStore, id: SessionId): Promise<DraftSnapshot> {
  return store.flush(id).then(() => store.snapshot(id))
}
\`\`\`

## 2. 发送前确认

发送或入队前，必须确认所用正文、附件与引用已经保存。宿主侧的实现大致如下：

\`\`\`go
func (c *Control) Enqueue(ctx context.Context, sid SessionID, d Draft) (InputID, error) {
	if err := c.drafts.Confirm(ctx, sid, d.Version); err != nil {
		return "", fmt.Errorf("draft not saved: %w", err)
	}
	return c.queue.Push(ctx, sid, d.Freeze())
}
\`\`\`

## 3. 转入队列

草稿转为队列项时，附件与固定引用的保留关系**同时转移**，不能出现内容暂时无主、可能被回收的间隙。

\`\`\`bash
pnpm test --filter @aibo/core -- draft queue
\`\`\`

以上改动不影响已发送消息使用的引用版本。如果需要，我可以继续补充“缺口后重新同步”的测试。`;

function rng(seed: number) {
  return () => { seed = (seed * 1664525 + 1013904223) % 4294967296; return seed / 4294967296; };
}
const SENTENCES = [
  '先确认相关文件与当前状态，再决定改动范围。', '队列推进受暂停、信任和有效设置共同约束。', '结果未知不能自动重发，只能明确手动重新发送。',
  '引用在加入输入框时固定版本，来源后续内容不会自动加入。', '目标、回合、队列与后台任务分别表示与控制。', '流式正文在可靠保存后批量发布给界面。',
  '取消请求不代表已经停止，需要等待 Agent 确认。', '子 Agent 的过程记录与父会话关联保存。',
];
function generateHistory(n: number): Item[] {
  const r = rng(42), out: Item[] = [];
  const para = (k: number) => Array.from({ length: k }, () => SENTENCES[Math.floor(r() * SENTENCES.length)]).join('');
  for (let i = 0; out.length < n; i++) {
    out.push({ kind: 'user', id: uid('h'), text: `第 ${i + 1} 轮：${para(1 + Math.floor(r() * 2))}`, chips: [], delivery: 'received' });
    if (out.length < n && r() < 0.6) out.push({ kind: 'tool', id: uid('h'), name: r() < 0.5 ? 'Read' : 'Shell', target: `src/core/module-${i % 17}.ts`, status: 'done', output: para(3), dur: `${(r() * 3).toFixed(1)}s` });
    if (out.length < n) {
      let text = Array.from({ length: 1 + Math.floor(r() * 4) }, () => para(2 + Math.floor(r() * 4))).join('\n\n');
      if (r() < 0.3) text += `\n\n\`\`\`ts\nexport function step${i}(input: Input): Output {\n  return transform(input, { round: ${i} })\n}\n\`\`\``;
      out.push({ kind: 'agent', id: uid('h'), text, done: true });
    }
  }
  return out.slice(0, n);
}
