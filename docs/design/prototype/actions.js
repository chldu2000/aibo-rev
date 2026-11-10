/* 交互：用户意图 → 状态变化 → 重新渲染。Agent 行为为模拟。 */
const ta = $('#ta');
const now = () => new Date().toTimeString().slice(0, 5);
const later = (s, ms, fn) => { const t = setTimeout(() => { s.timers = (s.timers || []).filter((x) => x !== t); fn(); renderLive(s.id === S.current ? 'new' : undefined); }, ms); (s.timers = s.timers || []).push(t); };
const clearTimers = (s) => { (s.timers || []).forEach(clearTimeout); s.timers = []; };
const push = (s, it) => { s.items.push(it); return it; };

/* ---------- overlay helpers ---------- */
function closeOverlay() { $('#overlay').innerHTML = ''; S.palette = null; S.menuFns = null; S.modalFns = null; }
function showModal({ title, body, actions }) {
  S.modalFns = actions.map((a) => a.fn);
  $('#overlay').innerHTML = `<div class="scrim" data-a="scrim"><div class="modal" role="dialog" aria-modal="true" aria-label="${esc(title)}"><h2>${title}</h2><div class="mb">${body}</div>
    <div class="mf">${actions.map((a, i) => `<button class="btn ${a.cls || ''}" data-a="modalAct" data-i="${i}" ${a.dis ? 'disabled' : ''}>${a.label}</button>`).join('')}</div></div></div>`;
  const b = $('#overlay .mf .btn.primary, #overlay .mf .btn.solid') || $('#overlay .mf .btn'); b && b.focus();
}
function showMenu(anchor, items) {
  const r = anchor.getBoundingClientRect();
  S.menuFns = items.map((x) => x.fn);
  $('#overlay').innerHTML = `<div class="scrim" data-a="scrim" style="background:transparent"></div><div class="menu" role="menu">${items.map((x, i) => x === 'sep' ? '<div class="sep"></div>' : x.h ? `<div class="mh">${x.h}</div>`
    : `<div class="mi ${x.dis ? 'dis' : ''}" role="menuitem" tabindex="0" data-a="${x.dis ? '' : 'menuAct'}" data-i="${i}" ${x.tip ? `title="${esc(x.tip)}"` : ''}>${x.label}${x.d ? `<span class="d">${x.d}</span>` : ''}</div>`).join('')}</div>`;
  const m = $('#overlay .menu');
  const top = r.bottom + 4 + m.offsetHeight > innerHeight ? r.top - m.offsetHeight - 4 : r.bottom + 4;
  m.style.top = top + 'px'; m.style.left = Math.min(r.left, innerWidth - m.offsetWidth - 8) + 'px';
}
function toast(text, actions = [], ms = 5000) {
  const el = document.createElement('div'); el.className = 'toast';
  el.innerHTML = `<div class="grow">${text}${actions.length ? `<div class="acts">${actions.map((a, i) => `<button class="btn sm" data-i="${i}">${a.label}</button>`).join('')}</div>` : ''}</div><button class="ib" aria-label="关闭">${icon('x', 'sm')}</button>`;
  el.addEventListener('click', (e) => { const b = e.target.closest('button'); if (!b) return; if (b.dataset.i != null) actions[b.dataset.i].fn(); el.remove(); });
  $('#toasts').appendChild(el); if (ms) setTimeout(() => el.remove(), ms);
}

/* ---------- simulated agent ---------- */
function runTurn(s, reply) {
  s.running = true; s.failed = false; s.startedAt = Date.now();
  later(s, 700, () => push(s, { t: 'thinking', secs: 3, text: '先确认相关文件和当前状态，再决定改动范围。' }));
  const tool = { t: 'tool', name: 'Read', target: 'src/core/queue.ts', st: 'running', dur: '' };
  later(s, 1400, () => push(s, tool));
  later(s, 2600, () => { tool.st = 'done'; tool.dur = '0.2s'; tool.out = '214 行'; push(s, { t: 'agent', html: `<p>${esc(reply || '已处理。我检查了相关代码，改动集中在队列状态迁移一处，没有影响其他模块。')}</p>` }); endTurn(s); });
}
function endTurn(s) {
  s.running = false;
  if (s.ctx) s.ctx[0] = Math.min(s.ctx[1], s.ctx[0] + 3);
  if (s.id !== S.current && S.notify) toast(`「${esc(s.title)}」回复完成`, [{ label: '查看', fn: () => A.open(null, s.id) }], 3500);
  later(s, 300, () => afterTurn(s));
}
function afterTurn(s) {
  if (s.running || s.stopping || s.queuePaused || !s.queue.length || s.queue[0].unknown) return;
  if (!wsOf(s).trusted || S.saveFail) return;
  const q = s.queue.shift(); deliver(s, q.text, q.chips);
}
function deliver(s, text, chips, note) {
  s.last = Date.now();
  const it = push(s, { t: 'user', text, chips, d: 'delivering', note });
  later(s, 600, () => { it.d = 'received'; runTurn(s); });
}

/* ---------- stop ---------- */
function stopSession(s) {
  if (!s.running || s.stopping) return;
  const a = ag(s);
  clearTimers(s);
  s.items.forEach((i) => { if ((i.t === 'approval' || i.t === 'question') && i.state === 'pending') i.state = 'expired'; if (i.t === 'tool' && i.st === 'running') { i.st = 'failed'; i.out = '已随回合停止'; } });
  s.stopping = { turn: 'sent', queue: s.queue.length, goal: s.goal && s.goal.status === 'active' ? (a.goalPause ? 'sent' : 'unsupported') : null };
  if (s.queue.length) s.queuePaused = '已停止当前执行，队列已暂停';
  s.stopItem = push(s, { t: 'stop' });
  renderLive('bottom');
  if (s.stopping.goal === 'sent') later(s, 1100, () => { s.stopping.goal = 'confirmed'; s.goal.status = 'paused'; });
  later(s, 1900, () => {
    if (s.agent === 'claude') { s.stopping.turn = 'unconfirmed'; return; }
    s.stopping.turn = 'confirmed'; finishStop(s);
  });
}
function finishStop(s) {
  s.stopItem.final = { ...s.stopping }; s.stopping = null; s.running = false;
  s.subagents.forEach((x) => { if (x.status === 'running') x.status = 'done'; });
  if (s.afterStop) { const q = s.afterStop; s.afterStop = null; later(s, 200, () => deliver(s, q.text, q.chips, '停止后发送的选中项')); }
}

/* ---------- picker ---------- */
function detectTrig() {
  const pos = ta.selectionStart, v = ta.value.slice(0, pos);
  const m = v.match(/(^|\s)([@#/])([^\s@#/]*)$/);
  if (!m) return closePicker();
  S.trig = { ch: m[2], start: pos - m[3].length - 1, q: m[3].toLowerCase() }; S.pickIdx = 0; renderPicker();
}
function pickerItems() {
  const { ch, q } = S.trig;
  if (ch === '@') return FILES.filter((f) => f.toLowerCase().includes(q)).map((f) => ({ ic: f.endsWith('/') ? 'folder' : 'file', label: f, d: f.endsWith('/') ? '目录' : '文件', v: f }));
  if (ch === '#') return S.sessions.filter((s) => s.id !== S.current && !s.archived && s.ws === cur().ws && s.title.toLowerCase().includes(q)).map((s) => ({ ic: 'hash', label: s.title, d: `${ag(s).name} · ${s.running ? '执行中' : s.updated}`, v: s.id }));
  const a = ag(cur());
  return COMMANDS.filter((c) => c.n.includes(q)).map((c) => ({ ic: 'slash', label: c.n, d: c.n === '/goal' && !a.goal ? `不可用：${a.goalWhy}` : `${c.d} · ${c.src}`, v: c.n }));
}
function renderPicker() {
  const items = pickerItems(), el = $('#picker');
  const head = { '@': '文件与目录 · 遵循忽略规则 · 按路径引用', '#': '引用同工作区的会话 · 加入时固定版本', '/': `${ag(cur()).name} 命令与 Skills` }[S.trig.ch];
  el.innerHTML = `<div class="ph"><span>${head}</span><span>↑↓ 选择 · ⏎ 确认 · Esc 关闭</span></div>` + (items.length ? items.map((x, i) => `<div class="pi ${i === S.pickIdx ? 'on' : ''}" data-a="pick" data-i="${i}">${icon(x.ic, 'sm')}<span class="${S.trig.ch === '#' ? '' : 'mono'}">${esc(x.label)}</span><span class="d">${esc(x.d)}</span></div>`).join('') : '<div class="pi muted">没有匹配项</div>');
  el.hidden = false;
}
function closePicker() { S.trig = null; $('#picker').hidden = true; }
function pick(i) {
  const x = pickerItems()[i]; if (!x) return;
  const { ch, start } = S.trig, end = ta.selectionStart;
  const s = cur(); S.chips[s.id] = S.chips[s.id] || [];
  let ins = '';
  if (ch === '@') S.chips[s.id].push({ type: 'file', label: x.v });
  else if (ch === '#') { const src = sess(x.v); S.chips[s.id].push({ type: 'ref', sid: src.id, at: now(), range: `近 ${S.refDefaults.turns} 轮`, chars: 1800 * src.items.length + 700, incomplete: src.running, stale: false }); }
  else if (x.v === '/goal') { ta.value = ta.value.slice(0, start) + ta.value.slice(end); closePicker(); if (ag(s).goal) A.setGoal(); else toast(ag(s).goalWhy); return; }
  else ins = x.v + ' ';
  ta.value = ta.value.slice(0, start) + ins + ta.value.slice(end);
  ta.selectionStart = ta.selectionEnd = start + ins.length;
  S.drafts[s.id] = ta.value; closePicker(); renderChips(); renderCbar(); renderInspector(); ta.focus();
}

/* ---------- palette ---------- */
const PCMDS = [
  ['新建会话', '⌘N', () => A.newSession()], ['打开设置', '⌘,', () => A.settings()], ['切换深色/浅色主题', '', () => A.toggleTheme()],
  ['收起/展开侧栏', '⌘B', () => A.toggleSidebar()], ['收起/展开工具面板', '⌥⌘I', () => A.toggleInspector()], ['查看已归档会话', '', () => A.showArchive()],
  ['重新加载呈现', '⌘⇧R', () => toast('（原型）正在重新加载呈现… 草稿与阅读位置将从共享状态恢复。')], ['切回默认呈现', '', () => toast('当前已是默认呈现。')],
];
const PSETS = [['界面语言', 'general'], ['系统通知', 'general'], ['主题', 'appearance'], ['强调色', 'appearance'], ['工作区信任', 'workspaces'], ['Agent 检测', 'agents'], ['呈现包', 'packages'], ['引用摘录范围', 'refs'], ['诊断日志', 'diag']];
function hl(t, q) { if (!q) return esc(t); const i = t.toLowerCase().indexOf(q); if (i < 0) return esc(t); return esc(t.slice(0, i)) + '<mark>' + esc(t.slice(i, i + q.length)) + '</mark>' + esc(t.slice(i + q.length)); }
function snippet(t, q) { const i = t.toLowerCase().indexOf(q); const st = Math.max(0, i - 24); return (st ? '…' : '') + t.slice(st, i + q.length + 60); }
function paletteResults() {
  let raw = S.palette.q, scope = S.palette.scope;
  const pre = { '>': 'cmd', '#': 'sess', '?': 'msg', '@': 'file' }[raw[0]];
  if (pre) { scope = pre; raw = raw.slice(1); }
  const q = raw.trim().toLowerCase(), G = [];
  const want = (x) => scope === 'all' || scope === x;
  const tag = (s) => s.archived ? ' <span class="tag">已归档</span>' : '';
  if (!q) {
    G.push(['最近会话', S.sessions.filter((s) => !s.archived).slice(0, 4).map((s) => ({ ic: 'hash', t: esc(s.title), sub: `${esc(wsOf(s).name)} · ${ag(s).name}`, fn: () => A.open(null, s.id) }))]);
    G.push(['命令', PCMDS.slice(0, 4).map(([t, kb, fn]) => ({ ic: 'right', t, sub: '', kb, fn }))]);
    return G;
  }
  if (want('cmd')) G.push(['命令', PCMDS.filter(([t]) => t.toLowerCase().includes(q)).map(([t, kb, fn]) => ({ ic: 'right', t: hl(t, q), kb, fn }))]);
  if (want('sess')) G.push(['会话', S.sessions.filter((s) => s.title.toLowerCase().includes(q)).map((s) => ({ ic: 'hash', t: hl(s.title, q) + tag(s), sub: `${esc(wsOf(s).name)} · ${ag(s).name}`, fn: () => A.open(null, s.id) }))]);
  if (want('msg')) {
    const hits = [];
    S.sessions.forEach((s) => s.items.forEach((it, i) => {
      const txt = it.t === 'user' ? it.text : it.t === 'agent' ? it.html.replace(/<[^>]+>/g, '') : it.t === 'tool' ? `${it.name} ${it.target} ${it.out || ''}` : '';
      if (txt && txt.toLowerCase().includes(q)) hits.push({ ic: it.t === 'user' ? 'quote' : it.t === 'tool' ? 'tool' : 'bot', t: hl(snippet(txt, q), q), sub: `${esc(s.title)} · ${it.t === 'user' ? '你' : ag(s).name}` + tag(s), fn: () => jumpTo(s.id, i) });
    }));
    G.push(['消息', hits.slice(0, 8)]);
  }
  if (want('file')) G.push(['文件 · 选择后作为 @ 引用插入输入框', FILES.filter((f) => f.toLowerCase().includes(q)).map((f) => ({ ic: f.endsWith('/') ? 'folder' : 'file', t: `<span class="mono">${hl(f, q)}</span>`, fn: () => { (S.chips[S.current] = S.chips[S.current] || []).push({ type: 'file', label: f }); renderChips(); ta.focus(); } }))]);
  if (scope === 'all') {
    G.push(['命令与 Skills', COMMANDS.filter((c) => (c.n + c.d).toLowerCase().includes(q)).map((c) => ({ ic: 'slash', t: hl(c.n, q), sub: `${c.d} · ${c.src}`, fn: () => { ta.value = c.n + ' ' + ta.value; ta.focus(); } }))]);
    G.push(['设置', PSETS.filter(([t]) => t.toLowerCase().includes(q)).map(([t, p]) => ({ ic: 'settings', t: hl(t, q), fn: () => { A.settings(); S.settingsPage = p; renderAll(); } }))]);
  }
  return G.filter((g) => g[1].length);
}
function renderPalette() {
  const P = S.palette, G = paletteResults(); P.flat = G.flatMap((g) => g[1]); P.idx = Math.min(P.idx, Math.max(0, P.flat.length - 1));
  let n = 0;
  const scopes = [['all', '全部'], ['cmd', '> 命令'], ['sess', '# 会话'], ['msg', '? 消息'], ['file', '@ 文件']];
  const html = `<div class="scrim" data-a="scrim"><div class="palette" role="dialog" aria-label="搜索与命令">
    <div class="pin">${icon('search')}<input id="pq" value="${esc(P.q)}" placeholder="搜索会话、消息、文件、命令与设置…" autocomplete="off"><span class="kbd">Esc</span></div>
    <div class="scopes">${scopes.map(([kk, n2]) => `<button class="scope ${P.scope === kk ? 'on' : ''}" data-a="pscope" data-s="${kk}">${n2}</button>`).join('')}<span style="flex:1"></span><button class="scope on">全部工作区</button><button class="scope on">含已归档</button></div>
    <div class="res">${G.length ? G.map(([gname, rs]) => `<div class="grp">${gname}</div>` + rs.map((r) => `<div class="r ${n === P.idx ? 'on' : ''}" data-a="prun" data-i="${n++}">${icon(r.ic, 'sm')}<div class="rt"><div>${r.t}</div>${r.sub ? `<div class="sub">${r.sub}</div>` : ''}</div>${r.kb ? `<span class="kbd">${r.kb}</span>` : ''}</div>`).join('')).join('') : '<div class="grp">没有结果。可以切换到“全部”范围，或在文件搜索中包含隐藏文件。</div>'}</div>
    <div class="foot"><span>↑↓ 选择</span><span>⏎ 打开</span><span>前缀 &gt; # ? @ 限定范围</span><span style="margin-left:auto">已跳过 12 个二进制文件</span></div></div></div>`;
  const had = document.activeElement && document.activeElement.id === 'pq', pos = had ? $('#pq').selectionStart : P.q.length;
  $('#overlay').innerHTML = html;
  const inp = $('#pq'); inp.focus(); inp.selectionStart = inp.selectionEnd = pos;
  const on = $('#overlay .r.on'); on && on.scrollIntoView({ block: 'nearest' });
}
function jumpTo(sid, idx) {
  A.open(null, sid);
  requestAnimationFrame(() => { const el = $(`#stream .it[data-idx="${idx}"]`); if (!el) return; el.scrollIntoView({ block: 'center' }); el.classList.add('flash'); setTimeout(() => el.classList.remove('flash'), 2100); });
}

/* ---------- actions ---------- */
const A = {
  open(el, id) {
    id = id || el.dataset.id; const prev = cur();
    if (prev) { S.drafts[prev.id] = ta.value; (S.scrollPos = S.scrollPos || {})[prev.id] = $('#stream').scrollTop; }
    S.current = id; S.view = 'session'; S.drill = null; S.pop = null; closePicker(); closeOverlay();
    ta.value = S.drafts[id] || ''; renderAll('restore');
  },
  toggleWs(el) { const w = S.workspaces.find((x) => x.id === el.dataset.id); w.open = !w.open; renderSidebar(); },
  toggleSidebar() { S.sidebarHidden = !S.sidebarHidden; renderAll(); },
  toggleInspector() { S.tool = S.tool ? null : (S.lastTool || 'git'); if (S.tool) S.lastTool = S.tool; renderTools(); },
  tool(el) { S.tool = S.tool === el.dataset.id ? null : el.dataset.id; if (S.tool) S.lastTool = S.tool; renderTools(); },
  pop(el) { S.pop = S.pop === el.dataset.p ? null : el.dataset.p; renderPending(); },
  pin(el) { const s = sess(el.dataset.id); s.pinned = !s.pinned; renderSidebar(); toast(s.pinned ? '已置顶' : '已取消置顶', [], 1500); },
  details() {
    const s = cur();
    showModal({ title: `会话详情 · ${esc(s.title)}`, body: `<div class="hint" style="margin-bottom:10px">这些信息默认不在工作台中展示，主要用于排查问题。</div>${contextHtml(s)}`, actions: [{ label: '关闭' }] });
  },
  headMenu(el) {
    const s = cur();
    showMenu(el, [{ label: s.pinned ? '取消置顶' : '置顶', fn: () => A.pin({ dataset: { id: s.id } }) }, { label: '改名', d: 'F2', fn: () => { S.renaming = s.id; renderSidebar(); } }, { label: '会话详情…', fn: A.details }, 'sep', { label: '归档', fn: () => archive(s) }]);
  },
  toggleTheme() { S.theme = S.theme === 'dark' ? 'light' : 'dark'; S.themePref = S.theme === 'dark' ? '深色' : '浅色'; renderAll(); },
  tabQueue() { S.pop = 'queue'; renderPending(); },
  drill(el) { (S.scrollPos = S.scrollPos || {})[S.current] = $('#stream').scrollTop; S.drill = { type: el.dataset.type, id: el.dataset.id }; S.pop = null; renderAll(); $('#stream').scrollTop = 0; },
  undrill() { S.drill = null; renderAll('restore'); },
  fold(el) { const id = el.dataset.id; S.open.has(id) ? S.open.delete(id) : S.open.add(id); renderStream(); },
  scrollBottom() { const s = $('#stream'); s.scrollTop = s.scrollHeight; $('#newContent').hidden = true; },
  copy() { toast('已复制'); },
  toast(el) { toast(el.dataset.t); },
  scrim(el, _, e) { if (e.target === el) closeOverlay(); },
  modalAct(el) { const fn = S.modalFns[el.dataset.i]; closeOverlay(); fn && fn(); },
  menuAct(el) { const fn = S.menuFns[el.dataset.i]; closeOverlay(); fn && fn(); },
  jumpReq() { const r = pendingReq(cur())[0]; if (!r) return; const el = $('#req-' + r.id); el.scrollIntoView({ block: 'center' }); el.focus(); el.classList.add('flash'); },
  jumpState() { const s = cur(), st = status(s); if (st === 'waiting') A.jumpReq(); else if (s.recover) $('#recover')?.scrollIntoView({ block: 'center' }); else if (s.queue.length && st === 'unknown') A.tabQueue(); else A.scrollBottom(); },

  /* composer */
  send() {
    const s = cur(), text = ta.value.trim(), chips = S.chips[s.id] || [];
    if (!text && !chips.length) { ta.focus(); return; }
    const br = blockReason(s); if (br) return toast(br);
    const queueIt = s.running || s.stopping || (s.recover && s.recover.state !== 'done') || (s.queuePaused && s.queue.length);
    if (queueIt) { s.queue.push({ id: 'q' + Date.now(), text: text || '（仅图片）', chips, at: now() }); toast(`已加入队列（第 ${s.queue.length} 条）${s.queuePaused ? '。队列当前已暂停。' : '，当前回合结束后投递。'}`, [], 2500); }
    else deliver(s, text, chips);
    ta.value = ''; S.drafts[s.id] = ''; S.chips[s.id] = []; renderChips(); renderLive('bottom');
  },
  steer() {
    const s = cur(), text = ta.value.trim(); if (!text) return ta.focus();
    const it = push(s, { t: 'user', text, steer: true, d: 'delivering' });
    later(s, 600, () => { it.d = 'received'; it.note = '已插入当前回合'; });
    ta.value = ''; S.drafts[s.id] = ''; renderLive('bottom');
  },
  stop() { stopSession(cur()); },
  waitStop() { const s = cur(); toast('继续等待 Agent 确认停止…', [], 1500); later(s, 1200, () => { s.stopping.turn = 'confirmed'; finishStop(s); }); },
  rmChip(el, _, e) { e.stopPropagation(); S.chips[S.current].splice(+el.dataset.i, 1); renderChips(); renderCbar(); renderInspector(); },
  addImage() { (S.chips[S.current] = S.chips[S.current] || []).push({ type: 'image', label: `截图 ${now()}.png` }); renderChips(); ta.focus(); },
  trig(el) { const c = el.dataset.ch, p = ta.selectionStart, pre = p && !/\s/.test(ta.value[p - 1]) ? ' ' : ''; ta.value = ta.value.slice(0, p) + pre + c + ta.value.slice(p); ta.focus(); ta.selectionStart = ta.selectionEnd = p + pre.length + 1; detectTrig(); },
  pick(el) { pick(+el.dataset.i); },
  refMenu(el) {
    const s = cur(), a = ag(s), i = +el.dataset.i, c = S.chips[s.id][i];
    showMenu(el, [{ h: `固定于 ${c.at} · 来源后续内容不会自动加入` },
      { label: `${c.range.startsWith('近') ? '✓ ' : '　'}近 ${S.refDefaults.turns} 轮原文`, fn: () => { c.range = `近 ${S.refDefaults.turns} 轮`; c.chars = Math.min(c.chars, 9800); renderChips(); renderCbar(); } },
      a.fullRead ? { label: `${c.range === '全文' ? '✓ ' : '　'}全文`, d: '由 Agent 按需读取', fn: () => { c.range = '全文'; c.chars = 41200; renderChips(); renderCbar(); } } : { label: '　全文', d: '不可用', dis: true, tip: '接收 Agent 不支持全文读取工具，仅能附带摘录。' },
      'sep',
      { label: c.stale ? '更新到最新版本' : '已是最新版本', d: c.stale ? '新增 2 轮' : '', dis: !c.stale, fn: () => { c.stale = false; c.at = now(); c.chars += 2100; renderChips(); renderCbar(); toast('引用已更新为新的固定版本。已发送的消息仍使用原版本。'); } },
      { label: '预览摘录', fn: () => showModal({ title: `引用摘录 · ${esc(sess(c.sid).title)}`, body: `<p class="hint">固定版本 ${c.at} · ${c.range} · ${k(c.chars)} 字符</p><pre class="term mono">[用户] 给投递尝试加上独立 ID……\n[工具] Shell  rg "DeliveryAttempt" -n src\n[Codex] 已拆出 DeliveryAttempt，状态分为……\n── 已省略 4 轮，可按需读取同版本全文 ──</pre>`, actions: [{ label: '关闭', fn: null }] }) },
      'sep', { label: '移除', fn: () => { S.chips[s.id].splice(i, 1); renderChips(); renderCbar(); } }]);
  },
  setMenu(el) {
    const s = cur(), a = ag(s), key = el.dataset.k;
    const list = { model: a.models, reasoning: a.reasoning, mode: a.modes }[key];
    showMenu(el, [{ h: '在下一次投递时生效，不影响正在进行的回合' }, ...list.map((v) => ({ label: (s[key] === v ? '✓ ' : '　') + v, fn: () => applySetting(s, key, v) }))]);
  },
  toggleFast() { const s = cur(); applySetting(s, 'fast', !s.fast); },

  /* approval & question */
  approve(el) {
    const s = cur(), it = s.items.find((x) => x.id === el.dataset.id); if (!it || it.state !== 'pending') return;
    it.state = 'submitted'; it.choice = it.options[+el.dataset.o]; renderLive();
    later(s, 800, () => {
      it.state = 'resolved';
      if (it.choice === '拒绝') { push(s, { t: 'agent', html: '<p>好的，我不运行测试。改动已保留，你可以稍后手动验证。</p>' }); endTurn(s); return; }
      const t = push(s, { t: 'tool', name: 'Bash', target: it.cmd, st: 'running' });
      later(s, 1600, () => { t.st = 'done'; t.dur = '14.2s'; t.out = 'Test Files  18 passed (18)\n     Tests  142 passed (142)'; s.subagents.forEach((x) => { x.status = 'done'; });
        push(s, { t: 'agent', html: '<p>core 包 142 个测试全部通过。草稿转入队列时的保留关系也有覆盖。</p>' }); endTurn(s); });
    });
  },
  qsel(el) { const it = cur().items.find((x) => x.id === el.dataset.id); it.sel = +el.dataset.o; renderStream(); },
  qsubmit(el) { const s = cur(), it = s.items.find((x) => x.id === el.dataset.id); it.state = 'answered'; renderLive(); runTurn(s, `好的，按“${it.options[it.sel]}”处理。我会先在 WM_DPICHANGED 中重设 WebView 边界并补一个双屏拖动的回归用例。`); },
  qskip(el) { const s = cur(), it = s.items.find((x) => x.id === el.dataset.id); it.state = 'answered'; it.sel = null; renderLive(); runTurn(s, '已跳过这个问题，我按影响最小的方案继续。'); },
  retry() { const s = cur(), e = s.items.find((x) => x.t === 'error' && !x.resolved); e.resolved = '已重试'; s.failed = false; renderLive(); runTurn(s, '已重新对照 10 号决议整理完毕，共 14 项，见下表草稿。'); },
  skipFail() { const s = cur(), e = s.items.find((x) => x.t === 'error' && !x.resolved); e.resolved = '已跳过'; s.failed = false; s.queuePaused = null; renderLive(); afterTurn(s); renderLive('bottom'); },

  /* queue */
  resumeQueue(el, _, e) { e && e.stopPropagation(); const s = cur(); s.queuePaused = null; renderLive(); afterTurn(s); renderLive('bottom'); },
  pauseQueue() { cur().queuePaused = '已手动暂停'; renderLive(); },
  clearQueue() { const s = cur(); const n = s.queue.filter((q) => !q.unknown).length; showModal({ title: '清空队列？', body: `<p>将删除 ${n} 条尚未开始投递的消息。结果未知的项不受影响。</p>`, actions: [{ label: '取消' }, { label: '清空', cls: 'danger solid', fn: () => { s.queue = s.queue.filter((q) => q.unknown); renderLive(); } }] }); },
  delQ(el) { const s = cur(); s.queue = s.queue.filter((q) => q.id !== el.dataset.q); renderLive(); },
  editQ(el) { const s = cur(), q = s.queue.find((x) => x.id === el.dataset.q); s.queue = s.queue.filter((x) => x !== q); ta.value = q.text + (ta.value ? '\n' + ta.value : ''); S.drafts[s.id] = ta.value; renderLive(); ta.focus(); toast('已取回到输入框，原队列项已移除。', [], 2500); },
  sendNow(el) {
    const s = cur(), a = ag(s), q = s.queue.find((x) => x.id === el.dataset.q);
    const take = () => { s.queue = s.queue.filter((x) => x !== q); };
    if (!s.running && !s.stopping) { take(); deliver(s, q.text, q.chips, '立即发送的选中项；其余队列保持原状态'); return renderLive('bottom'); }
    if (a.steer) { take(); const it = push(s, { t: 'user', text: q.text, steer: true, d: 'delivering' }); later(s, 600, () => { it.d = 'received'; }); return renderLive('bottom'); }
    showModal({ title: '停止当前执行后发送此条？', body: `<p>${esc(a.name)} 不支持插话。发送这条需要先停止当前回合。</p><ul><li>停止得到确认后才会发送此条</li><li>其余 ${s.queue.length - 1} 条队列保持暂停</li></ul>`,
      actions: [{ label: '取消' }, { label: '停止后发送', cls: 'primary', fn: () => { take(); s.afterStop = q; stopSession(s); } }] });
  },
  verify(el) { const s = cur(); toast('正在向 Agent 核实这次投递…', [], 1500); later(s, 1400, () => toast('Agent 未能提供可核实的接收证据，结果仍为未知。可以选择重新发送（可能重复执行）。')); },
  resend(el) {
    const s = cur(), qid = el.dataset.q, q = s.queue.find((x) => x.id === qid);
    showModal({ title: '重新发送这条消息？', body: `<p>上一次投递的结果未知，Agent <b>可能已经收到并执行过</b>。重新发送可能导致重复执行。</p><p class="hint">原记录会保留，新的投递尝试会关联到它。其余队列仍保持暂停。</p>`,
      actions: [{ label: '取消' }, { label: '重新发送（可能重复执行）', cls: 'primary', fn: () => {
        s.queue = s.queue.filter((x) => x !== q);
        const orig = s.items.find((x) => x.qid === qid); if (orig) orig.d = 'unknownDone';
        if (s.queue.length) s.queuePaused = '已手动重发结果未知的项；其余队列仍暂停';
        deliver(s, q.text, q.chips, '第 2 次尝试 · 关联第 1 次（结果未知）'); renderLive('bottom');
      } }] });
  },

  /* goal & tasks */
  setGoal() {
    const s = cur();
    showModal({ title: '设置持续目标', body: `<p class="hint">目标由 Agent 原生推进，可能立即开始新的回合。它与消息队列分开控制。</p><textarea id="goalText" class="fld" rows="3">让队列相关测试全部通过</textarea>`,
      actions: [{ label: '取消' }, { label: '设置并开始', cls: 'primary', fn: () => { s.goal = { text: $('#goalText')?.value || '目标', status: 'active', since: now(), budget: [['Token', '0 / 200k', 'Agent 报告'], ['时长', '未报告', '']] }; if (!s.running) runTurn(s, '已根据目标开始工作。'); renderLive(); } }] });
    setTimeout(() => { const g = $('#goalText'); g && g.focus(); });
  },
  goalPause() { const s = cur(); s.goal.status = 'pausing'; renderLive(); later(s, 1200, () => { s.goal.status = 'paused'; }); },
  goalResume() { const s = cur(); s.goal.status = 'resuming'; renderLive(); later(s, 1000, () => { s.goal.status = 'active'; if (!s.running) runTurn(s, '目标已恢复，继续处理剩余的失败用例。'); }); },
  goalClear() { const s = cur(); showModal({ title: '清除目标？', body: '<p>只清除持续目标，不删除已有历史和引用，也不停止当前回合或队列。</p>', actions: [{ label: '取消' }, { label: '清除', cls: 'danger solid', fn: () => { s.goal.status = 'clearing'; renderLive(); later(s, 1000, () => { s.goal = null; }); } }] }); },
  cancelTask(el) { const s = cur(), t = s.tasks.find((x) => x.id === el.dataset.id); showModal({ title: '取消后台任务？', body: `<p class="mono">${esc(t.cmd)}</p><p class="hint">只取消这一个任务。实际影响范围以 Agent 报告为准。</p>`, actions: [{ label: '保留' }, { label: '取消任务', cls: 'danger solid', fn: () => { t.status = 'cancelling'; renderLive(); later(s, 1500, () => { t.status = 'cancelled'; t.out += '\n^C 已被用户取消\n'; }); } }] }); },

  /* sessions & workspaces */
  newSession(el) {
    const wsId = el?.dataset?.ws || cur().ws; let pickA = 'codex';
    const body = () => `<label class="hint">工作区</label><select id="nsWs" class="fld" style="margin-bottom:10px">${S.workspaces.map((w) => `<option value="${w.id}" ${w.id === wsId ? 'selected' : ''}>${esc(w.name)}${w.trusted ? '' : '（未信任）'}</option>`).join('')}</select>
      ${Object.entries(AGENTS).map(([id, a]) => `<div class="agent-opt ${id === pickA ? 'sel' : ''}" data-a="nsPick" data-id="${id}">${mark(a)}<div class="grow"><div>${a.name} ${a.verified ? '<span class="tag success">已验证</span>' : '<span class="tag unknown">未验证组合</span>'}</div><div class="hint ${a.ok ? '' : 'warn'}">${a.ok ? `${a.adapter} · 可执行文件与认证检查通过` : a.okWhy}</div></div></div>`).join('')}`;
    S.nsPick = (id) => { pickA = id; $('#overlay .mb').innerHTML = body(); const ok = AGENTS[id].ok, b = $('#overlay .mf .primary'); b.disabled = !ok; b.title = ok ? '' : AGENTS[id].okWhy; };
    showModal({ title: '新建会话', body: body(), actions: [{ label: '取消' }, { label: '创建 ⏎', cls: 'primary', fn: () => {
      const a = AGENTS[pickA], w = $('#nsWs')?.value || wsId, id = 'n' + Date.now();
      S.sessions.unshift({ id, ws: w, title: '新会话', agent: pickA, model: a.models[0], reasoning: a.reasoning[1] || a.reasoning[0] || '', fast: false, mode: a.modes[1] || a.modes[0], ctx: a.reasoning.length ? [0, 200] : null, running: false, updated: '刚刚', queue: [], queuePaused: null, goal: null, subagents: [], tasks: [], items: [] });
      A.open(null, id); ta.focus();
    } }] });
    S.wsPick = wsId;
  },
  nsPick(el) { S.nsPick(el.dataset.id); },
  sessMenu(el, _, e) {
    e.stopPropagation(); const s = sess(el.dataset.id);
    showMenu(el, [{ label: s.pinned ? '取消置顶' : '置顶', fn: () => A.pin({ dataset: { id: s.id } }) }, { label: '改名', d: 'F2', fn: () => { S.renaming = s.id; renderSidebar(); } }, { label: '归档', fn: () => archive(s) }, { label: '复制会话信息', fn: () => toast('已复制会话信息') }, 'sep', { label: '关闭会话', d: '默认界面不提供', dis: true, tip: '默认工作台用会话列表切换；离开会话不会结束执行。' }]);
  },
  wsMenu(el, _, e) {
    e.stopPropagation(); const w = S.workspaces.find((x) => x.id === el.dataset.id);
    showMenu(el, [{ label: '新建会话', fn: () => A.newSession({ dataset: { ws: w.id } }) }, { label: w.trusted ? '撤销信任…' : '信任此工作区', fn: () => (w.trusted ? A.revokeTrust : A.trust)({ dataset: { id: w.id } }) }, { label: '在 Finder 中显示', fn: () => toast(esc(w.path)) }, 'sep', { label: '移除工作区…', fn: () => A.removeWs({ dataset: { id: w.id } }) }]);
  },
  trust(el) { const w = S.workspaces.find((x) => x.id === el.dataset.id); w.trusted = true; toast(`已信任 ${esc(w.name)}。之前暂停的队列不会自动恢复。`); renderAll(); },
  revokeTrust(el) {
    const w = S.workspaces.find((x) => x.id === el.dataset.id), act = S.sessions.filter((s) => s.ws === w.id && (s.running || (s.goal && s.goal.status === 'active')));
    showModal({ title: `撤销对 ${esc(w.name)} 的信任？`, body: `<p>将立即阻止新的投递，以及 Agent 对文件和引用的访问，并请求停止正在运行的活动。历史和引用内容会保留。</p>${act.length ? `<ul>${act.map((s) => `<li>${esc(s.title)}</li>`).join('')}</ul>` : ''}<p class="hint">重新信任后，暂停的队列不会自动恢复。</p>`,
      actions: [{ label: '取消' }, { label: '撤销信任', cls: 'danger solid', fn: () => { w.trusted = false; act.forEach(stopSession); S.sessions.filter((s) => s.ws === w.id && s.queue.length).forEach((s) => { s.queuePaused = '工作区信任已撤销，队列已暂停'; }); renderAll(); } }] });
  },
  removeWs(el) { const w = S.workspaces.find((x) => x.id === el.dataset.id); showModal({ title: `移除工作区 ${esc(w.name)}？`, body: '<p>只移除目录入口。业务历史、项目文件和暂停的队列都会保留，可在搜索中按“已移除”来源找到历史。</p>', actions: [{ label: '取消' }, { label: '移除', cls: 'danger solid', fn: () => toast('（原型）未实际移除示例工作区。') }] }); },
  addWs() { toast('（原型）打开系统文件夹选择器 → 信任确认 → 新建会话'); },
  unarchive(el) { const s = sess(el.dataset.id); s.archived = false; toast(`已取消归档「${esc(s.title)}」`); A.open(null, s.id); },
  showArchive() { S.view = 'archive'; renderAll(); },
  settings() { S.view = 'settings'; renderAll(); },
  settingsPage(el) { S.view = 'settings'; S.settingsPage = el.dataset.p; if (el.dataset.sub) S.pkgTab = el.dataset.sub; renderAll(); },
  pkgTab(el) { S.pkgTab = el.dataset.t; renderAll(); },
  back() { S.view = 'session'; renderAll('restore'); },
  setOpt(el) {
    const { k: key, v } = el.dataset;
    if (key === 'theme') { S.themePref = v; S.theme = v === '深色' ? 'dark' : v === '浅色' ? 'light' : (matchMedia('(prefers-color-scheme: dark)').matches ? 'dark' : 'light'); }
    if (key === 'lang') { S.lang = v; if (v === 'English') toast('（原型仅提供中文文案）'); }
    if (key === 'notify') S.notify = v === '开';
    renderAll();
  },
  accent(el) { S.accent = el.dataset.c; renderAll(); },
  usePres(el) { toast(`（原型）切换到「${esc(el.dataset.n)}」。若加载失败会自动切回默认呈现。`, [{ label: '模拟加载失败', fn: () => SC.fallback() }]); },
  uninstallPkg(el) { showModal({ title: `卸载 ${esc(el.dataset.n)}？`, body: el.dataset.t === 'agent' ? '<p>使用该 Agent 的 5 个会话仍可阅读、搜索和引用，但不能继续发送。有活动时需等其结束后才生效。</p>' : '<p>包的私有状态会保留，重新安装后可以恢复。</p>', actions: [{ label: '取消' }, { label: '卸载', cls: 'danger solid', fn: () => toast('（原型）未实际卸载') }] }); },
  upgradePkg(el) { showModal({ title: `升级 ${esc(el.dataset.n)}`, body: '<p>新版本：1.4.0</p><p>受影响的会话：5 个，其中 1 个正在执行。升级会在相关活动结束后生效，并先迁移宿主副本。</p>', actions: [{ label: '取消' }, { label: '活动结束后升级', cls: 'primary', fn: () => toast('已安排升级，相关活动结束后生效。') }] }); },
  palette() { S.palette = { q: '', scope: 'all', idx: 0 }; renderPalette(); },
  pscope(el) { S.palette.scope = el.dataset.s; S.palette.idx = 0; renderPalette(); },
  prun(el) { const r = S.palette.flat[+el.dataset.i]; closeOverlay(); r && r.fn(); },

  /* recovery */
  reconnect() { const s = cur(); s.recover.state = 'connecting'; renderLive(); later(s, 1500, () => { s.recover.state = 'connected'; }); },
  readOnly() { toast('保持只读。可以随时从会话底部的恢复卡继续。'); },
  recoverDone() { const s = cur(); s.recover.state = 'done'; renderLive(); },
  dismissRecovery() { S.recovery = false; renderBanners(); },
  saveDetail() { showModal({ title: '关键保存失败', body: '<p><b>未能保存的内容</b></p><ul><li>重构共享交互模块的草稿状态：最近 3 条 Agent 输出</li></ul><p><b>停止状态</b></p><ul><li>重构共享交互模块的草稿状态 · <span style="color:var(--unknown)">停止未确认</span></li><li>队列与投递证据 · 后台任务 cargo test · 已停止</li></ul><p class="hint">这两类事实分开记录：内容未保存不代表执行已停止。</p>', actions: [{ label: '关闭' }] }); },
  retrySave() { S.saveFail = false; toast('保存已恢复。暂停的队列和执行不会自动恢复。'); renderAll(); },
};
function applySetting(s, key, v) {
  (S.applying[s.id] = S.applying[s.id] || {})[key] = true; renderLive();
  later(s, 1100, () => { s[key] = v; delete S.applying[s.id][key]; });
}
function archive(s) {
  const busy = s.running || (s.goal && s.goal.status === 'active') || s.tasks.some((t) => t.status === 'running');
  const doIt = () => { s.archived = true; if (s.queue.length) s.queuePaused = '会话已归档，队列已暂停'; if (S.current === s.id) A.open(null, S.sessions.find((x) => !x.archived).id); else renderAll(); toast(`已归档「${esc(s.title)}」`, [{ label: '撤销', fn: () => { s.archived = false; renderAll(); } }]); };
  if (!busy) return doIt();
  showModal({ title: '会话仍有活动', body: `<p>「${esc(s.title)}」需要先停止以下活动才能归档：</p><ul>${s.running ? '<li>当前回合正在执行</li>' : ''}${s.goal && s.goal.status === 'active' ? '<li>持续目标推进中</li>' : ''}${s.tasks.filter((t) => t.status === 'running').map((t) => `<li>后台任务 <span class="mono">${esc(t.cmd)}</span></li>`).join('')}</ul><p class="hint">未发送的队列会保留并暂停。</p>`,
    actions: [{ label: '取消' }, { label: '停止后归档', cls: 'primary', fn: () => { clearTimers(s); s.running = false; s.stopping = null; if (s.goal) s.goal.status = 'paused'; s.tasks.forEach((t) => { if (t.status === 'running') t.status = 'cancelled'; }); toast('（原型）已确认停止'); doIt(); } }] });
}

/* ---------- scenarios ---------- */
const SC = {
  restart() {
    closeOverlay();
    ['s1', 's2'].forEach((id) => {
      const s = sess(id); clearTimers(s); s.running = false; s.stopping = null;
      s.items.forEach((i) => { if ((i.t === 'approval' || i.t === 'question') && i.state !== 'resolved' && i.state !== 'answered') i.state = 'expired'; if (i.t === 'tool' && i.st === 'running') { i.st = 'failed'; i.out = '应用退出时中断'; } });
      if (s.queue.length) s.queuePaused = '应用重启，队列已暂停';
      if (s.goal) s.goal.status = 'paused';
      s.subagents.forEach((x) => { x.status = 'done'; }); s.tasks.forEach((t) => { if (t.status === 'running') t.status = 'cancelled'; });
      s.items = s.items.filter((i) => i.t !== 'recover');
      push(s, { t: 'sys', text: `应用于 ${now()} 重启 · 此前回合未确认结束` }); push(s, { t: 'recover' }); s.recover = { state: 'idle' };
    });
    S.recovery = true; S.view = 'session'; renderAll('bottom'); toast('（原型）已模拟应用重启。草稿、队列和历史均已恢复，没有自动执行任何操作。');
  },
  saveFail() { S.saveFail = true; renderAll(); },
  fallback() { toast('外部呈现「Compact Pro 1.2」加载失败，已切回默认呈现。草稿和阅读位置已恢复。', [{ label: '查看原因', fn: () => showModal({ title: '呈现加载失败', body: '<pre class="term mono">Compact Pro 1.2.0\nbundle: index.js\nTypeError: presentation API v3 required, host offers v2\n→ 判定为不兼容，自动回退。</pre>', actions: [{ label: '关闭' }] }) }, { label: '再次尝试该呈现', fn: () => toast('（原型）再次加载 Compact Pro…') }], 0); },
  hang() {
    showModal({ title: '工作台无响应', body: `<p class="hint" style="margin-bottom:8px">这是宿主的原生恢复窗口（MyGO 原生 UI），不依赖当前 WebView。也可从应用菜单“工作台”、Dock 菜单或 Windows 通知区域菜单打开。</p>
      <dl class="kv"><dt>当前呈现</dt><dd>Compact Pro 1.2.0</dd><dt>状态</dt><dd style="color:var(--unknown)">15 秒未响应</dd><dt>运行中活动</dt><dd>3 项，不受影响</dd></dl>`,
      actions: [{ label: '继续等待' }, { label: '打开诊断', fn: () => toast('（原型）打开诊断') }, { label: '重新加载呈现', fn: () => toast('已重新加载呈现，草稿与阅读位置已恢复。') }, { label: '切回默认呈现', cls: 'primary', fn: () => toast('已切回默认呈现。执行未中断。') }] });
  },
  quit() {
    const acts = [];
    S.sessions.forEach((s) => { if (s.running) acts.push([s, '当前回合执行中']); if (s.goal && s.goal.status === 'active') acts.push([s, '持续目标推进中']); s.tasks.filter((t) => t.status === 'running').forEach((t) => acts.push([s, `后台任务 ${t.cmd}`])); if (s.queue.some((q) => q.unknown)) acts.push([s, '1 项投递结果未知（不会自动重发）']); });
    if (!acts.length) return toast('（原型）没有活动，直接退出。');
    showModal({ title: '退出 Aibo？', body: `<p>以下活动会被停止，恢复信息会被保存：</p><ul>${acts.map(([s, t]) => `<li>${esc(s.title)} · ${esc(t)}</li>`).join('')}</ul>`, actions: [{ label: '取消' }, { label: '停止并退出', cls: 'primary', fn: () => SC.quitting(acts) }] });
  },
  quitting(acts) {
    const st = acts.map(() => 'sent');
    const draw = (done) => showModal({ title: '正在停止活动…', body: `<ul class="stoplist">${acts.map(([s, t], i) => `<li>${st[i] === 'sent' ? '<span class="spin" style="color:var(--unknown)"></span>' : st[i] === 'ok' ? `<span class="st-done">${icon('check', 'sm')}</span>` : '<span class="tag unknown">未确认</span>'}${esc(s.title)} · ${esc(t)}</li>`).join('')}</ul>${done ? '<p class="hint">强制退出后，这些活动可能仍在运行，下次启动时会标记为待核实。</p>' : ''}`,
      actions: done ? [{ label: '继续等待', fn: () => { st.forEach((v, i) => { if (v === 'no') st[i] = 'ok'; }); draw(false); setTimeout(() => showModal({ title: '（原型）Aibo 已退出', body: '<p>所有活动已确认停止。</p>', actions: [{ label: '模拟重新启动', cls: 'primary', fn: SC.restart }] }), 900); } }, { label: '强制退出', cls: 'danger solid', fn: () => showModal({ title: '（原型）Aibo 已强制退出', body: '<p>未确认停止的活动在下次启动时标记为待核实。</p>', actions: [{ label: '模拟重新启动', cls: 'primary', fn: SC.restart }] }) }] : [] });
    draw(false);
    setTimeout(() => { acts.forEach(([s], i) => { st[i] = s.agent === 'claude' && s.running ? 'no' : 'ok'; }); draw(st.includes('no')); if (!st.includes('no')) setTimeout(() => showModal({ title: '（原型）Aibo 已退出', body: '<p>所有活动已确认停止。</p>', actions: [{ label: '模拟重新启动', cls: 'primary', fn: SC.restart }] }), 700); }, 1500);
  },
  closeWin() { const n = S.sessions.filter((s) => s.running || (s.goal && s.goal.status === 'active')).length; toast(`Aibo 仍在后台运行 ${n} 项活动。${document.documentElement.dataset.platform === 'mac' ? '可从 Dock 图标重新打开，从菜单栏退出。' : '可从通知区域图标重新打开或退出。'}<br><span class="hint">（只在第一次关闭最后一个窗口时提示）</span>`, [], 7000); },
  welcome() { S.view = 'welcome'; renderAll(); },
  platform() { const r = document.documentElement; r.dataset.platform = r.dataset.platform === 'mac' ? 'win' : 'mac'; renderScenario(); },
};
function renderScenario() {
  const el = $('#scenario');
  el.innerHTML = `<span class="lbl">原型场景（非产品界面）</span>
    <button data-sc="restart">重启恢复</button><button data-sc="saveFail">关键保存失败</button><button data-sc="fallback">呈现回退</button><button data-sc="hang">呈现无响应</button>
    <button data-sc="quit">退出应用</button><button data-sc="closeWin">关闭最后窗口</button><button data-sc="welcome">首次启动</button><button data-sc="platform">平台：${document.documentElement.dataset.platform === 'mac' ? 'macOS' : 'Windows'}</button>`;
}
$('#scenario').addEventListener('click', (e) => { const b = e.target.closest('[data-sc]'); if (!b) return; SC[b.dataset.sc](); });

/* ---------- events ---------- */
document.addEventListener('click', (e) => {
  const el = e.target.closest('[data-a]'); if (!el || !el.dataset.a) return;
  const fn = A[el.dataset.a]; if (fn) fn(el, el.dataset, e);
});
document.addEventListener('dblclick', (e) => { const t = e.target.closest('[data-a2="rename"]'); if (t) { S.renaming = t.dataset.id; renderSidebar(); } });
document.addEventListener('keydown', (e) => {
  const ri = e.target.id === 'renameInput';
  if (ri) { if (e.key === 'Enter' || e.key === 'Escape') { if (e.key === 'Enter') sess(S.renaming).title = e.target.value.trim() || sess(S.renaming).title; S.renaming = null; renderAll(); } return; }
  const mod = e.metaKey || e.ctrlKey;
  if (S.palette && e.target.id === 'pq') {
    if (e.key === 'ArrowDown' || e.key === 'ArrowUp') { e.preventDefault(); S.palette.idx = Math.max(0, Math.min(S.palette.flat.length - 1, S.palette.idx + (e.key === 'ArrowDown' ? 1 : -1))); renderPalette(); return; }
    if (e.key === 'Enter') { e.preventDefault(); const r = S.palette.flat[S.palette.idx]; closeOverlay(); r && r.fn(); return; }
  }
  if (e.key === 'Escape') { if ($('#overlay').innerHTML) return closeOverlay(); if (S.trig) return closePicker(); if (S.pop) { S.pop = null; return renderPending(); } if (S.drill) return A.undrill(); if (S.view !== 'session') return A.back(); }
  if (e.target === ta && S.trig) {
    const n = pickerItems().length;
    if (e.key === 'ArrowDown' || e.key === 'ArrowUp') { e.preventDefault(); S.pickIdx = (S.pickIdx + (e.key === 'ArrowDown' ? 1 : n - 1)) % Math.max(n, 1); return renderPicker(); }
    if ((e.key === 'Enter' || e.key === 'Tab') && !e.isComposing) { e.preventDefault(); return pick(S.pickIdx); }
  }
  if (e.target === ta && e.key === 'Enter' && !e.shiftKey && !e.isComposing && e.keyCode !== 229) {
    e.preventDefault();
    if (mod) { const s = cur(); if ((s.running || s.stopping) && ag(s).steer && !blockReason(s)) A.steer(); else if (s.running) toast(blockReason(s) || ag(s).steerWhy); else A.send(); }
    else A.send(); return;
  }
  if (e.target === ta && e.key === 'Backspace' && ta.selectionStart === 0 && ta.selectionEnd === 0 && (S.chips[S.current] || []).length) { S.chips[S.current].pop(); renderChips(); renderCbar(); return; }
  const ap = document.activeElement?.closest?.('[data-approval]');
  if (ap && /^[1-9]$/.test(e.key)) { const b = ap.querySelectorAll('[data-a="approve"]')[+e.key - 1]; b && b.click(); return; }
  if (e.target.matches?.('.srow') && e.key === 'Enter') return A.open(e.target);
  if (e.key === 'F2' && S.view === 'session' && e.target !== ta) { S.renaming = S.current; renderSidebar(); return; }
  if (!mod) return;
  const key = e.key.toLowerCase();
  if (key === 'k') { e.preventDefault(); A.palette(); }
  else if (key === 'n') { e.preventDefault(); A.newSession(); }
  else if (key === 'b') { e.preventDefault(); A.toggleSidebar(); }
  else if (e.code === 'KeyI' && e.altKey) { e.preventDefault(); A.toggleInspector(); }
  else if (key === '.') { e.preventDefault(); A.stop(); }
  else if (key === ',') { e.preventDefault(); A.settings(); }
  else if (key === 'l') { e.preventDefault(); ta.focus(); }
  else if (e.shiftKey && key === 'a') { e.preventDefault(); A.jumpReq(); }
  else if (e.shiftKey && key === 'r') { e.preventDefault(); toast('（原生菜单）正在重新加载呈现… 运行不受影响。'); }
  else if (/^[1-9]$/.test(e.key)) { const r = document.querySelector(`.srow[data-n="${e.key}"]`); if (r) { e.preventDefault(); A.open(r); } }
});
document.addEventListener('focusout', (e) => { if (e.target.id === 'renameInput' && S.renaming) { sess(S.renaming).title = e.target.value.trim() || sess(S.renaming).title; S.renaming = null; setTimeout(renderAll); } });
document.addEventListener('input', (e) => {
  if (e.target === ta) { S.drafts[S.current] = ta.value; detectTrig(); ta.style.height = 'auto'; ta.style.height = Math.min(240, ta.scrollHeight) + 'px'; }
  else if (e.target.id === 'pq') { S.palette.q = e.target.value; S.palette.idx = 0; renderPalette(); }
  else if (e.target.dataset.qtext) { cur().items.find((x) => x.id === e.target.dataset.qtext).text = e.target.value; }
  else if (e.target.dataset.ref) { S.refDefaults[e.target.dataset.ref] = +e.target.value || S.refDefaults[e.target.dataset.ref]; }
});
ta.addEventListener('click', detectTrig);
ta.addEventListener('blur', () => setTimeout(() => { if (document.activeElement !== ta && !$('#picker:hover')) closePicker(); }, 150));
ta.addEventListener('paste', (e) => { const img = [...(e.clipboardData?.items || [])].find((i) => i.type.startsWith('image/')); if (img) { e.preventDefault(); A.addImage(); } });
$('#stream').addEventListener('scroll', (e) => { const el = e.target; if (el.scrollHeight - el.scrollTop - el.clientHeight < 40) $('#newContent').hidden = true; });
setInterval(() => { if (S.view === 'session' && cur().running) renderHead(); }, 1000);
addEventListener('resize', () => { if (innerWidth < 1180 && S.tool) { S.tool = null; renderTools(); } });

renderScenario();
ta.value = S.drafts[S.current] || '';
renderAll('bottom');
requestAnimationFrame(() => A.scrollBottom());
