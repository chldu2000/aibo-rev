/* 渲染：状态 → DOM。所有交互状态集中在 S，呈现只读它。 */
const S = Object.assign(seed(), {
  current: 's1', view: 'session', tab: 'activity', drill: null, theme: 'light', themePref: '浅色', accent: '#4C5BD4',
  settingsPage: 'general', pkgTab: 'agent', lang: '跟随系统', open: new Set(), applying: {},
  drafts: { s1: '' }, chips: { s1: [{ type: 'ref', sid: 's2', at: '14:32', range: '近 3 轮', chars: 9800, stale: true }] },
  pop: null, tool: null,
  recovery: false, saveFail: false, refDefaults: { turns: 3, per: 12000, total: 32000 }, notify: true,
});
S.sessions.forEach((x, i) => { x.last = Date.now() - i * 60000; });
const $ = (q, r = document) => r.querySelector(q);
const esc = (s) => String(s ?? '').replace(/[&<>"']/g, (c) => ({ '&': '&amp;', '<': '&lt;', '>': '&gt;', '"': '&quot;', "'": '&#39;' }[c]));
const icon = (n, c = '') => `<svg class="i ${c}" viewBox="0 0 24 24" aria-hidden="true">${ICONS[n] || ''}</svg>`;
const sess = (id) => S.sessions.find((s) => s.id === id);
const cur = () => sess(S.current);
const ag = (s) => AGENTS[s.agent];
const wsOf = (s) => S.workspaces.find((w) => w.id === s.ws);
const k = (n) => (n >= 1000 ? (n / 1000).toFixed(1).replace(/\.0$/, '') + 'k' : String(n));
const mark = (a) => `<span class="agent-mark" style="background:${a.color}" role="img" aria-label="${esc(a.name)}" title="${esc(a.name)}">${a.short}</span>`;
const pendingReq = (s) => s.items.filter((i) => (i.t === 'approval' || i.t === 'question') && i.state === 'pending');

function status(s) {
  if (s.archived) return 'archived';
  if (s.stopping && s.stopping.turn === 'sent') return 'stopping';
  if (s.stopping && s.stopping.turn === 'unconfirmed') return 'unknown';
  if (pendingReq(s).length) return 'waiting';
  if (s.recover && s.recover.state !== 'done') return 'unknown';
  if (s.failed) return 'failed';
  if (s.running) return 'running';
  if (s.queue.some((q) => q.unknown)) return 'unknown';
  return 'idle';
}
function statusText(s) {
  const st = status(s);
  if (st === 'stopping') return '正在停止…';
  if (st === 'unknown' && s.stopping) return '停止未确认';
  if (st === 'waiting') return pendingReq(s)[0].t === 'approval' ? '等待审批' : '等待你回答';
  if (st === 'unknown' && s.recover) return '退出时有未完成活动';
  if (st === 'failed') return '回合失败';
  if (st === 'running') { const t = Math.floor((Date.now() - s.startedAt) / 1000); return `执行中 · ${Math.floor(t / 60)}:${String(t % 60).padStart(2, '0')}`; }
  if (st === 'unknown') return '投递结果未知';
  if (st === 'archived') return '已归档';
  return '空闲';
}
const ST_LABEL = { running: '执行中', waiting: '等待你处理', failed: '失败', unknown: '结果未知或未确认', stopping: '正在停止', idle: '空闲', archived: '已归档' };

/* ---------- sidebar ---------- */
function renderSidebar() {
  let h = `<div class="sb-top"><div class="traffic"><i></i><i></i><i></i></div>
    <button class="ib" data-a="toggleSidebar" data-tip="收起侧栏 ⌘B" aria-label="收起侧栏">${icon('sidebar')}</button>
    <button class="ib" data-a="newSession" data-tip="新建会话 ⌘N" aria-label="新建会话">${icon('plus')}</button></div>
    <div class="sb-search" data-a="palette" role="button" tabindex="0">${icon('search', 'sm')}<span>搜索与命令</span><span class="kbd">⌘K</span></div>
    <div class="sb-list">`;
  let n = 0;
  for (const w of S.workspaces) {
    h += `<div class="ws-head" title="${esc(w.path)}"><span class="ib" data-a="toggleWs" data-id="${w.id}" style="opacity:1;width:16px">${icon(w.open ? 'down' : 'right', 'sm')}</span>
      <span class="name" data-a="toggleWs" data-id="${w.id}">${esc(w.name)}</span>
      ${w.trusted ? '' : `<span data-tip="此工作区尚未信任：可以阅读历史，不能启动 Agent">${icon('lock', 'sm')}</span>`}
      <button class="ib" data-a="newSession" data-ws="${w.id}" aria-label="在此工作区新建会话">${icon('plus', 'sm')}</button>
      <button class="ib" data-a="wsMenu" data-id="${w.id}" aria-label="工作区菜单">${icon('more', 'sm')}</button></div>`;
    if (!w.open) continue;
    for (const s of S.sessions.filter((x) => x.ws === w.id && !x.archived).sort((a, b) => (b.pinned ? 1 : 0) - (a.pinned ? 1 : 0) || b.last - a.last)) {
      n++;
      const st = status(s), uq = s.queue.some((q) => q.unknown);
      h += `<div class="srow ${s.id === S.current && S.view === 'session' ? 'active' : ''}" data-a="open" data-id="${s.id}" data-n="${n}" tabindex="0" data-testid="session-row" data-session-id="${s.id}" data-status="${st}">
        <span class="dotbox"><span class="dot ${st}" role="img" aria-label="${ST_LABEL[st]}" title="${ST_LABEL[st]}" data-testid="session-status"></span></span>
        ${S.renaming === s.id ? `<input id="renameInput" value="${esc(s.title)}" aria-label="会话名称">` : `<span class="title" data-a2="rename" data-id="${s.id}">${esc(s.title)}</span>`}
        ${s.pinned ? `<span class="pinmark" title="已置顶">${icon('pin', 'sm')}</span>` : ''}
        ${s.queue.length ? `<span class="qbadge ${uq ? 'unknown' : ''}" title="${s.queue.length} 项未发送">${s.queue.length}</span>` : ''}
        ${mark(ag(s))}
        <button class="ib more" data-a="sessMenu" data-id="${s.id}" aria-label="会话菜单" data-testid="session-menu-button">${icon('more', 'sm')}</button></div>`;
    }
  }
  h += `</div><div class="sb-foot">
    <div class="row" data-a="addWs">${icon('folder', 'sm')}添加工作区</div>
    <div class="row" data-a="showArchive">${icon('archive', 'sm')}已归档 <span class="muted" style="margin-left:auto">${S.sessions.filter((s) => s.archived).length}</span></div>
    <div class="row" data-a="settings">${icon('settings', 'sm')}设置 <span class="kbd" style="margin-left:auto">⌘,</span></div>
    <div class="row" data-a="toggleTheme">${icon(S.theme === 'dark' ? 'sun' : 'moon', 'sm')}${S.theme === 'dark' ? '浅色' : '深色'}主题</div></div>`;
  $('#sidebar').innerHTML = h;
  const ri = $('#renameInput'); if (ri) { ri.focus(); ri.select(); }
}

/* ---------- banners ---------- */
function renderBanners() {
  let h = '';
  if (S.saveFail) h += `<div class="banner danger" role="alert" data-testid="banner-save-failure">${icon('alert')}<div class="grow"><b>无法写入会话历史：磁盘空间不足。</b>已暂停新的工作，并请求停止 2 项执行（<span style="color:var(--unknown)">1 项停止未确认</span>）。最近 3 条 Agent 输出尚未保存。</div>
    <button class="btn sm" data-a="saveDetail">查看详情</button><button class="btn sm" data-a="retrySave">重试保存</button><button class="btn sm ghost" data-a="toast" data-t="（原型）打开日志目录">打开日志目录</button></div>`;
  if (S.recovery) h += `<div class="banner info">${icon('refresh')}<div class="grow">已恢复上次的历史、草稿与队列。<b>2 个会话</b>在退出时有未完成活动，继续执行前需要你确认。</div>
    <button class="btn sm" data-a="open" data-id="s1">查看</button><button class="ib" data-a="dismissRecovery" aria-label="关闭">${icon('x', 'sm')}</button></div>`;
  $('#banners').innerHTML = h;
}

/* ---------- session header ---------- */
function renderHead() {
  const s = cur(), a = ag(s), st = status(s);
  $('#shead').innerHTML = `
    ${S.sidebarHidden ? `<button class="ib" data-a="toggleSidebar" aria-label="展开侧栏">${icon('sidebar')}</button>` : ''}
    <span class="t" data-a="renameHead" title="点击改名">${esc(s.title)}</span><span class="ws">${esc(wsOf(s).name)}</span>
    <span class="sp"></span>
    <span class="run-state ${st}" data-a="jumpState">${st === 'stopping' ? '<span class="spin"></span>' : `<span class="dot ${st}"></span>`}${statusText(s)}</span>
    <span class="agent-id">${mark(a)}${esc(a.name)} ${a.verified ? '' : '<span class="tag unknown" data-tip="此 Agent/适配器组合可以使用，但不在正式支持范围内。" data-testid="agent-unverified">未验证</span>'}</span>
    <button class="ib" data-a="headMenu" aria-label="会话菜单" data-testid="session-head-menu">${icon('more')}</button>`;
}

/* ---------- stream ---------- */
function chipHtml(c, i, editable) {
  const x = editable ? `<button class="x" data-a="rmChip" data-i="${i}" aria-label="移除">${icon('x', 'sm')}</button>` : '';
  if (c.type === 'file') return `<span class="chip" data-testid="attachment-chip" title="按路径引用，发送时读取当时内容">${icon(c.label.endsWith('/') ? 'folder' : 'file', 'sm')}<span class="lbl mono">${esc(c.label)}</span>${x}</span>`;
  if (c.type === 'image') return `<span class="chip" data-testid="attachment-chip"><span class="thumb"></span><span class="lbl">${esc(c.label)}</span>${x}</span>`;
  const src = sess(c.sid);
  return `<span class="chip ref" data-testid="reference-chip" ${editable ? `data-a="refMenu" data-i="${i}"` : ''} title="固定版本：${c.at}。来源后续的新内容不会自动加入。">${icon('hash', 'sm')}<span class="lbl">${esc(src ? src.title : c.sid)} · 固定 ${c.at} · ${c.range} · ${k(c.chars)}</span>${c.incomplete ? '<span class="tag attention">未完成</span>' : ''}${c.stale && editable ? '<span class="stale" title="来源有新内容，可更新"></span>' : ''}${editable ? icon('down', 'sm') : ''}${x}</span>`;
}

function itemHtml(s, it, idx) {
  const a = ag(s);
  const id = `${s.id}:${idx}`;
  const isOpen = S.open.has(id);
  switch (it.t) {
    case 'user': {
      const chips = it.chips && it.chips.length ? `<div class="chips">${it.chips.map((c) => chipHtml(c, 0, false)).join('')}</div>` : '';
      let d = '';
      if (it.d === 'queued') d = `<div class="dlv">${icon('clock', 'sm')}已排队</div>`;
      else if (it.d === 'delivering') d = `<div class="dlv"><span class="spin"></span>投递中…</div>`;
      else if (it.d === 'received') d = `<div class="dlv">${icon('check', 'sm')}已接收${it.note ? ` · ${esc(it.note)}` : ''}</div>`;
      else if (it.d === 'unknown') d = `<div class="dlv unknown">${icon('help', 'sm')}<span>结果未知 · 连接在投递过程中断开，无法确认 Agent 是否已接收</span>
        <button class="btn sm" data-a="verify" data-q="${it.qid}">核实</button><button class="btn sm" data-a="resend" data-q="${it.qid}">重新发送（可能重复执行）</button></div>`;
      else if (it.d === 'unknownDone') d = `<div class="dlv unknown">${icon('help', 'sm')}第 1 次尝试结果未知 · 已手动重新发送（见下方）</div>`;
      return `<div class="bubble-wrap"><div class="msg-user ${it.steer ? 'steer' : ''}" data-testid="message-user" data-delivery="${it.d === 'unknownDone' ? 'unknown' : it.d}"${it.d === 'unknownDone' ? ' data-resent="true"' : ''}>${it.steer ? '<span class="tag accent" style="margin-bottom:4px">插话</span>\n' : ''}${chips}${esc(it.text)}</div>${d}</div>`;
    }
    case 'agent': return `<div class="role">${mark(a)}${esc(a.name)}</div><div class="msg-agent" data-testid="message-agent">${it.html}</div>`;
    case 'thinking': return `<div class="fold thinking" data-testid="thinking-card"><div class="fh" data-a="fold" data-id="${id}" role="button" tabindex="0" aria-expanded="${isOpen}">${icon('spark', 'sm')}<span>思考 · ${it.secs} 秒</span><span class="tgt"></span>${icon(isOpen ? 'down' : 'right', 'sm')}</div>${isOpen ? `<div class="fb">${esc(it.text)}</div>` : ''}</div>`;
    case 'group': return `<div class="fold"><div class="fh" data-a="fold" data-id="${id}">${icon('file', 'sm')}<span>${esc(it.label)}</span><span class="tgt"></span>${icon(isOpen ? 'down' : 'right', 'sm')}</div>${isOpen ? `<div class="fb"><div class="subfold mono">${it.items.map((f) => `<div>${esc(f)}</div>`).join('')}</div></div>` : ''}</div>`;
    case 'tool': {
      const sti = it.st === 'running' ? '<span class="spin st-run"></span>' : it.st === 'failed' ? `<span class="st-fail">${icon('x', 'sm')}</span>` : `<span class="st-done">${icon('check', 'sm')}</span>`;
      return `<div class="fold" data-testid="tool-card"><div class="fh" data-a="fold" data-id="${id}" role="button" tabindex="0" aria-expanded="${isOpen}">${icon(it.name === 'Shell' || it.name === 'Bash' ? 'term' : 'tool', 'sm')}<b style="font-weight:500">${esc(it.name)}</b><span class="tgt mono">${esc(it.target)}</span>${sti}<span class="muted small">${it.dur || ''}</span>${icon(isOpen ? 'down' : 'right', 'sm')}</div>${isOpen ? `<div class="fb"><pre class="mono">${esc(it.out || '（无输出）')}</pre></div>` : ''}</div>`;
    }
    case 'approval': {
      if (it.state === 'expired') return `<div class="card expired" data-testid="approval-card" data-state="expired"><div class="ct">${icon('shield')}审批请求已失效</div><div class="cmd">${esc(it.tool)}  ${esc(it.cmd)}</div><div class="note">此请求已失效（连接已重建），不能再作答。需要时请重新发起。</div></div>`;
      const done = it.state !== 'pending';
      return `<div class="card attention" tabindex="0" data-approval="${it.id}" id="req-${it.id}" data-testid="approval-card" data-state="${it.state}" role="group" aria-label="需要审批">
        <div class="ct">${icon('shield')}需要审批</div>
        <div class="cmd"><span class="muted">${esc(it.tool)}</span>  ${esc(it.cmd)}</div>
        <div class="muted small">工作目录 ${esc(it.cwd)}</div>
        <div class="acts">${done
          ? (it.state === 'submitted' ? `<span class="tag unknown"><span class="spin"></span>已提交：${esc(it.choice)}，等待 Agent 确认</span>` : `<span class="tag success">${icon('check', 'sm')}已答复：${esc(it.choice)}</span>`)
          : it.options.map((o, i) => `<button class="btn ${i === 0 ? 'primary' : ''} sm" data-a="approve" data-id="${it.id}" data-o="${i}"><span class="kbd">${i + 1}</span>${esc(o)}</button>`).join('')}</div>
        ${done ? '' : '<div class="note">选项由 Agent 提供，按原样显示。聚焦此卡片后可按数字键选择。</div>'}</div>`;
    }
    case 'question': {
      const done = it.state !== 'pending';
      return `<div class="card attention" id="req-${it.id}" data-testid="question-card" data-state="${done ? 'answered' : 'pending'}" role="radiogroup" aria-label="${esc(it.title)}"><div class="ct">${icon('help')}${esc(it.title)}</div>
        ${it.options.map((o, i) => `<div class="opt ${it.sel === i ? 'sel' : ''}" role="radio" tabindex="0" aria-checked="${it.sel === i}" data-a="${done ? '' : 'qsel'}" data-id="${it.id}" data-o="${i}"><span class="num">${i + 1}</span>${esc(o)}</div>`).join('')}
        ${done ? `<div class="acts"><span class="tag success">${icon('check', 'sm')}已回答</span></div>` : `<textarea class="fld" rows="2" placeholder="${esc(it.extra)}" data-qtext="${it.id}">${esc(it.text || '')}</textarea>
        <div class="acts"><button class="btn primary sm" data-a="qsubmit" data-id="${it.id}" ${it.sel == null ? 'disabled' : ''}>提交回答</button><button class="btn sm ghost" data-a="qskip" data-id="${it.id}">跳过此问题</button>
        <span class="muted small">填写中的回答会随草稿保存，切换会话不会丢失，也不会自动提交。</span></div>`}</div>`;
    }
    case 'subagent': {
      const sa = s.subagents.find((x) => x.id === it.ref);
      return `<div class="fold" data-testid="subagent-card"><div class="fh" data-a="drill" data-type="sub" data-id="${sa.id}" role="button" tabindex="0">${icon('bot', 'sm')}<span>子 Agent · ${esc(sa.name)}</span><span class="tgt"></span>${sa.status === 'running' ? '<span class="spin st-run"></span>' : `<span class="st-done">${icon('check', 'sm')}</span>`}<span class="muted small">查看过程</span>${icon('right', 'sm')}</div></div>`;
    }
    case 'error': return `<div class="card danger"><div class="ct">${icon('alert')}回合失败</div><div>${esc(it.msg)}</div>
      ${S.open.has(id) ? `<pre class="term mono" style="margin-top:8px">${esc(it.detail)}</pre>` : ''}
      <div class="acts">${it.resolved ? `<span class="tag">${esc(it.resolved)}</span>` : `<button class="btn primary sm" data-a="retry">${icon('refresh', 'sm')}重试</button><button class="btn sm" data-a="skipFail">跳过并恢复队列</button>`}<button class="btn sm ghost" data-a="fold" data-id="${id}">${S.open.has(id) ? '收起' : '查看'}原始错误</button></div>
      ${it.resolved ? '' : '<div class="note">后续 1 项排队消息已暂停，等待你决定。</div>'}</div>`;
    case 'gap': return `<div class="sysline gap" data-testid="gap-marker">${icon('alert', 'sm')}${esc(it.text)}</div>`;
    case 'sys': return `<div class="sysline">${esc(it.text)}</div>`;
    case 'stop': {
      const p = s.stopping || it.final; if (!p) return '';
      const row = (k2, v) => `<li><span class="k">${k2}</span>${v}</li>`;
      const tv = { sent: '<span class="spin" style="color:var(--unknown)"></span>停止请求已发送', confirmed: `<span class="st-done">${icon('check', 'sm')}</span>已停止`, unconfirmed: '<span class="tag unknown">未确认</span>Agent 尚未确认当前回合已结束' };
      const gv = { sent: '<span class="spin" style="color:var(--unknown)"></span>暂停请求已发送', confirmed: `<span class="st-done">${icon('check', 'sm')}</span>已暂停`, unsupported: `<span class="hint warn">此 Agent 不支持暂停目标，目标可能继续推进</span>` };
      return `<div class="card unknown" data-testid="stop-summary"><div class="ct">${icon('stop')}停止当前执行</div><ul class="stoplist">
        ${row('当前回合', tv[p.turn])}
        ${row('未发送队列', p.queue ? `已暂停（${p.queue}）` : '<span class="muted">无</span>')}
        ${p.goal ? row('持续目标', gv[p.goal]) : ''}
        ${row('后台任务', '<span class="muted">不受影响，可在输入框上方的“后台任务”中单独取消</span>')}</ul>
        ${p.turn === 'unconfirmed' && s.stopping ? '<div class="acts"><button class="btn sm" data-a="waitStop">继续等待</button></div>' : ''}</div>`;
    }
    case 'recover': {
      const r = s.recover; if (!r) return '';
      if (r.state === 'idle') return `<div class="card unknown" id="recover" data-testid="recovery-card"><div class="ct">${icon('refresh')}应用重启前有未完成的活动</div>
        <ul class="stoplist"><li><span class="k">当前回合</span><span class="tag unknown">未确认结束</span></li><li><span class="k">未发送队列</span>已暂停（${s.queue.length}）</li>${s.goal ? `<li><span class="k">持续目标</span>${s.goal.status === 'active' ? '退出时推进中，现已停止推进' : '已暂停'}</li>` : ''}</ul>
        <div class="acts"><button class="btn primary sm" data-a="reconnect">恢复连接</button><button class="btn sm ghost" data-a="readOnly">仅查看</button></div>
        <div class="note">恢复连接会重新启动 Agent 并加载原生会话，Agent 可能会因此继续执行。队列和目标仍需分别恢复。</div></div>`;
      if (r.state === 'connecting') return `<div class="card unknown" data-testid="recovery-card"><div class="ct"><span class="spin"></span>正在启动 Agent 并加载原生会话…</div></div>`;
      if (r.state === 'connected') return `<div class="card" data-testid="recovery-card"><div class="ct" style="color:var(--success)">${icon('check')}已重新连接 · 原生会话已加载</div>
        <div class="muted small">Agent 报告：上次回合已中断，没有正在执行的工作。</div>
        <div class="acts">${s.queuePaused ? `<button class="btn sm" data-a="resumeQueue">恢复队列（${s.queue.length}）</button>` : ''}${s.goal && s.goal.status !== 'active' ? '<button class="btn sm" data-a="goalResume">恢复目标</button>' : ''}<button class="btn sm ghost" data-a="recoverDone">完成</button></div>
        <div class="note">恢复队列不会恢复目标，恢复目标也不会恢复队列。</div></div>`;
      return '';
    }
  }
  return '';
}

function renderStream(keepScroll) {
  const s = cur(), el = $('#stream');
  $('#composerWrap').hidden = !!S.drill;
  if (S.drill) { el.innerHTML = `<div class="it">${drillHtml(s)}</div>`; return; }
  const atBottom = el.scrollHeight - el.scrollTop - el.clientHeight < 40;
  const prev = el.scrollTop;
  if (!s.items.length) {
    const a = ag(s);
    el.innerHTML = `<div class="empty"><h3>${mark(a)} ${esc(a.name)} · ${esc(s.model)}</h3><div>新会话。输入消息开始工作。</div>
      <div class="hints"><span><span class="kbd">@</span> 文件/目录</span><span><span class="kbd">#</span> 引用会话</span><span><span class="kbd">/</span> 命令与 Skills</span></div></div>`;
    return;
  }
  el.innerHTML = s.items.map((it, i) => `<div class="it" data-idx="${i}">${itemHtml(s, it, i)}</div>`).join('');
  if (keepScroll === 'restore') { el.style.scrollBehavior = 'auto'; el.scrollTop = S.scrollPos?.[s.id] ?? el.scrollHeight; el.style.scrollBehavior = ''; $('#newContent').hidden = true; return; }
  if (atBottom || keepScroll === 'bottom') { el.scrollTop = el.scrollHeight; $('#newContent').hidden = true; }
  else { el.scrollTop = prev; if (keepScroll === 'new') $('#newContent').hidden = false; }
}

/* ---------- composer ---------- */
function blockReason(s) {
  if (S.saveFail) return '关键保存失败，新的工作已暂停。草稿仍可编辑。';
  if (!wsOf(s).trusted) return '工作区未信任，不能启动 Agent 或发送。';
  if (s.archived) return '已归档会话只能阅读。';
  if (Object.keys(S.applying[s.id] || {}).length) return '正在应用执行设置，确认后才能发送，不会按旧值发送。';
  return null;
}
function renderPending() {
  const s = cur(); let h = '';
  const w = wsOf(s);
  if (!w.trusted) h += `<div class="trust" data-testid="trust-block">${icon('lock')}<div class="grow"><b>此工作区尚未信任。</b><div class="hint">信任后 Aibo 才会在这里启动 Agent，并向 Agent 提供文件和引用读取。现在可以阅读历史、编辑草稿。</div></div><button class="btn primary" data-a="trust" data-id="${w.id}">信任此工作区</button></div>`;
  if (s.archived) h += `<div class="trust" data-testid="archived-block">${icon('archive')}<div class="grow">已归档的会话可以阅读、搜索和被引用，不能继续发送。</div><button class="btn" data-a="unarchive" data-id="${s.id}">取消归档后继续</button></div>`;
  const chips = stripChips(s);
  if (S.pop && !chips.some((c) => c.includes(`data-p="${S.pop}"`))) S.pop = null;
  if (S.pop) h += `<div class="pop" role="dialog" data-testid="strip-popover" data-kind="${S.pop}" aria-label="${POP_T[S.pop]}"><div class="pop-h"><b>${POP_T[S.pop]}</b><span class="grow"></span><span class="kbd">Esc</span><button class="ib" data-a="pop" data-p="${S.pop}" aria-label="收起">${icon('x', 'sm')}</button></div><div class="pop-b">${popHtml(s)}</div></div>`;
  if (chips.length) h += `<div class="strip" data-testid="status-strip">${chips.join('')}</div>`;
  $('#pending').innerHTML = h;
}
function renderChips() {
  const s = cur(); const cs = S.chips[s.id] || [];
  $('#chips').innerHTML = cs.map((c, i) => chipHtml(c, i, true)).join('');
}
function refBudget(s) { return (S.chips[s.id] || []).filter((c) => c.type === 'ref').reduce((a, c) => a + c.chars, 0); }
function renderCbar() {
  const s = cur(), a = ag(s), busy = s.running || !!s.stopping;
  const br = blockReason(s);
  const rb = refBudget(s);
  const over = rb > S.refDefaults.total || (S.chips[s.id] || []).some((c) => c.type === 'ref' && c.chars > S.refDefaults.per);
  let h = `<button class="ib" data-a="addImage" data-tip="添加图片（也可直接粘贴）" aria-label="添加图片">${icon('image')}</button>
    <button class="ib" data-a="trig" data-ch="@" data-tip="引用文件或目录" aria-label="引用文件">${icon('at')}</button>
    <button class="ib" data-a="trig" data-ch="#" data-tip="引用其他会话" aria-label="引用会话">${icon('hash')}</button>
    <button class="ib" data-a="trig" data-ch="/" data-tip="命令与 Skills" aria-label="命令">${icon('slash')}</button><span class="grow"></span>`;
  if (rb) h += `<span class="budget ${over ? 'over' : ''}" ${over ? `data-tip="超出部分将按省略标记截断。可改为全文，由 Agent 按需读取。"` : ''}>引用 ${k(rb)} / ${k(S.refDefaults.total)}</span>`;
  if (s.running && !s.stopping) h += `<button class="btn" data-a="stop" data-tip="停止当前执行 ⌘." data-testid="composer-stop">${icon('stop', 'sm')}停止</button>`;
  if (busy) {
    const can = a.steer && !br;
    h += `<span ${can ? '' : `data-tip="${esc(br || a.steerWhy)}"`}><button class="btn ${can ? '' : 'is-disabled'}" data-a="${can ? 'steer' : ''}" data-testid="composer-steer" ${can ? '' : 'aria-disabled="true" aria-describedby="steer-why"'}>${icon('bolt', 'sm')}插话 <span class="kbd">⌘⏎</span></button>${can ? '' : `<span id="steer-why" class="sr-only">${esc(br || a.steerWhy)}</span>`}</span>`;
  }
  const label = busy || (s.recover && s.recover.state !== 'done') || (s.queuePaused && s.queue.length) ? '排队' : '发送';
  h += br ? `<span data-tip="${esc(br)}"><button class="btn primary is-disabled" aria-disabled="true" aria-describedby="primary-why" data-testid="composer-primary">${icon('send', 'sm')}${label} <span class="kbd">⏎</span></button><span id="primary-why" class="sr-only">${esc(br)}</span></span>`
          : `<button class="btn primary" data-a="send" data-testid="composer-primary">${icon('send', 'sm')}${label} <span class="kbd">⏎</span></button>`;
  $('#cbar').innerHTML = h;
}
function renderSettingsBar() {
  const s = cur(), a = ag(s), ap = S.applying[s.id] || {};
  const sel = (key, val, list, why) => {
    if (!list || !list.length) return `<span class="sel disabled" data-tip="${esc(why)}" role="button" tabindex="0" aria-disabled="true" aria-describedby="${key}-why">${key === 'reasoning' ? '推理' : key}：不可用</span><span id="${key}-why" class="sr-only">${esc(why)}</span>`;
    return `<button class="sel ${ap[key] ? 'applying' : ''}" data-a="setMenu" data-k="${key}" ${key === 'model' ? `data-testid="model-select" data-applying="${!!ap[key]}"` : ''} ${ap[key] ? 'data-tip="应用中，等待 Agent 确认"' : ''}>${esc(val)}${icon('down', 'sm')}</button>`;
  };
  let h = sel('model', s.model, a.models) + sel('reasoning', s.reasoning ? `推理 ${s.reasoning}` : '', a.reasoning, a.reasoningWhy);
  h += a.fast ? `<button class="sel ${ap.fast ? 'applying' : ''}" data-a="toggleFast">Fast ${s.fast ? '开' : '关'}</button>` : `<span class="sel disabled" data-tip="${esc(a.fastWhy)}" data-testid="setting-fast" role="button" tabindex="0" aria-disabled="true" aria-describedby="fast-why">Fast</span><span id="fast-why" class="sr-only">${esc(a.fastWhy)}</span>`;
  h += sel('mode', `权限 ${s.mode}`, a.modes);
  h += `<span class="grow"></span>`;
  h += s.ctx ? `<span class="ctxmeter" data-a="details" title="上下文占用 · 点击查看会话详情">上下文 ${Math.round(s.ctx[0] / s.ctx[1] * 100)}%<span class="bar"><i style="width:${s.ctx[0] / s.ctx[1] * 100}%"></i></span>${s.ctx[0]}k/${s.ctx[1]}k</span>`
             : `<span class="ctxmeter" data-tip="此 Agent 未报告上下文占用。">上下文 未报告</span>`;
  $('#settingsBar').innerHTML = h;
}

/* ---------- status strip above composer ---------- */
const GOAL_ST = { active: ['推进中', 'accent'], paused: ['已暂停', 'attention'], pausing: ['暂停请求已发送', 'unknown'], resuming: ['恢复请求已发送', 'unknown'], clearing: ['清除请求已发送', 'unknown'], done: ['已完成', 'success'] };
const POP_T = { goal: '持续目标', sub: '子 Agent', task: '后台任务', queue: '消息队列' };
function stripChips(s) {
  const out = [], pr = pendingReq(s), on = (p) => (S.pop === p ? ' on' : '');
  if (pr.length) out.push(`<button class="schip attention" data-a="jumpReq" data-testid="strip-chip" data-kind="pending">${icon('shield', 'sm')}${pr.length} 个${pr[0].t === 'approval' ? '审批' : '问题'}待处理<span class="kbd">⌘⇧A</span></button>`);
  if (s.goal) { const g = GOAL_ST[s.goal.status]; out.push(`<button class="schip ${g[1] === 'unknown' ? 'unknown' : ''}${on('goal')}" data-a="pop" data-p="goal" data-testid="strip-chip" data-kind="goal">${g[1] === 'unknown' ? '<span class="spin"></span>' : `<span class="dot ${s.goal.status === 'active' ? 'running' : s.goal.status === 'paused' ? 'waiting' : 'idle'}"></span>`}目标 · ${g[0]}</button>`); }
  if (s.subagents.length) { const r = s.subagents.filter((x) => x.status === 'running').length; out.push(`<button class="schip${on('sub')}" data-a="pop" data-p="sub" data-testid="strip-chip" data-kind="sub">${icon('bot', 'sm')}子 Agent<span class="cnt">${s.subagents.length}</span>${r ? `<span class="dot running" title="${r} 个执行中"></span>` : ''}</button>`); }
  if (s.tasks.length) { const r = s.tasks.filter((x) => x.status === 'running').length; out.push(`<button class="schip${on('task')}" data-a="pop" data-p="task" data-testid="strip-chip" data-kind="task">${icon('term', 'sm')}后台任务<span class="cnt">${s.tasks.length}</span>${r ? `<span class="dot running" title="${r} 个运行中"></span>` : ''}</button>`); }
  if (s.queue.length) { const unk = s.queue.some((q) => q.unknown); out.push(`<button class="schip ${unk ? 'unknown' : s.queuePaused ? 'attention' : ''}${on('queue')}" data-a="pop" data-p="queue" data-testid="strip-chip" data-kind="queue">${icon(s.queuePaused ? 'pause' : 'list', 'sm')}队列<span class="cnt">${s.queue.length}</span>${unk ? ' · 结果未知' : s.queuePaused ? ' · 已暂停' : ''}</button>`); }
  return out;
}
function popHtml(s) {
  if (S.pop === 'goal') return goalHtml(s);
  if (S.pop === 'queue') return queueHtml(s);
  if (S.pop === 'sub') return s.subagents.map((x) => `<div class="irow" data-a="drill" data-type="sub" data-id="${x.id}"><span class="dot ${x.status === 'running' ? 'running' : 'idle'}"></span><span class="grow">${esc(x.name)}</span>${x.bg ? '<span class="tag">后台</span>' : ''}<span class="muted small">${x.status === 'running' ? x.since : '已完成'}</span>${icon('right', 'sm')}</div>`).join('');
  return s.tasks.map((t) => `<div class="irow" data-a="drill" data-type="task" data-id="${t.id}">${t.status === 'running' ? '<span class="dot running"></span>' : t.status === 'cancelling' ? '<span class="dot unknown"></span>' : '<span class="dot idle"></span>'}<span class="grow mono">${esc(t.cmd)}</span><span class="muted small">${{ running: '运行中', cancelling: '取消中', cancelled: '已取消' }[t.status]}</span>${icon('right', 'sm')}</div>`).join('') + '<div class="hint" style="margin-top:6px">后台任务不随“停止当前执行”结束，可在详情中单独取消。</div>';
}
function goalHtml(s) {
  const a = ag(s), g = s.goal, gs = GOAL_ST[g.status];
  return `<div style="display:flex;align-items:center;gap:6px"><span class="tag ${gs[1]}">${gs[0]}</span><span class="muted small">开始于 ${g.since} · 由 Agent 原生推进</span></div>
    <div class="gt" style="margin:6px 0 8px">${esc(g.text)}</div>
    <dl class="kv">${g.budget.map(([kk, v, src]) => `<dt>${kk}</dt><dd>${v === '未报告' ? '<span class="muted">未报告</span>' : `${v} <span class="muted">· ${src}</span>`}</dd>`).join('')}</dl>
    <div style="display:flex;gap:6px;margin-top:10px;flex-wrap:wrap">
      ${g.status === 'active' ? (a.goalPause ? `<span data-tip="只暂停目标，不停止当前回合和队列。"><button class="btn sm" data-a="goalPause">${icon('pause', 'sm')}暂停目标</button></span>` : `<span data-tip="${esc(a.goalPauseWhy)}"><button class="btn sm is-disabled">${icon('pause', 'sm')}暂停目标</button></span>`) : ''}
      ${g.status === 'paused' ? `<span data-tip="恢复目标可能立即开始新回合；不会恢复消息队列。"><button class="btn sm" data-a="goalResume">${icon('play', 'sm')}恢复目标</button></span>` : ''}
      <button class="btn sm danger" data-a="goalClear">清除</button></div>`;
}

/* ---------- tool panel area (capability plugins) ---------- */
const TOOLS = [
  { id: 'git', name: 'Git', icon: 'branch', pkg: 'Git 工作流 0.2.0 · 示例能力插件' },
  { id: 'term', name: '终端', icon: 'term', pkg: '终端 0.3.1 · 示例能力插件' },
];
function toolBody(id) {
  if (id === 'git') return `<div class="hint" style="margin-bottom:8px">aibo-rev · <span class="mono">main</span> · 领先 2</div>
    <div class="sec-h">变更 4</div>${[['M', 'src/presentation/shared/draft.ts', '+84 −37'], ['M', 'src/core/queue.ts', '+12 −3'], ['A', 'src/core/delivery-evidence.ts', '+141'], ['M', 'docs/design/ui-interaction.md', '+9 −2']].map(([t, f, d]) => `<div class="irow"><span class="mono" style="color:${t === 'A' ? 'var(--success)' : 'var(--attention)'}">${t}</span><span class="grow mono small">${f}</span><span class="muted small">${d}</span></div>`).join('')}
    <textarea class="fld" rows="2" placeholder="提交消息"></textarea><div style="display:flex;gap:6px;margin-top:6px"><button class="btn sm primary">提交</button><button class="btn sm ghost">让 Agent 审查改动</button></div>`;
  return `<pre class="term mono" style="max-height:none;min-height:320px">~/Workspace/aibo-dev/aibo-rev $ pnpm dev
  ➜  Local:   http://localhost:5173/
  ➜  ready in 412 ms
▍</pre>`;
}
function renderTools() {
  const t = TOOLS.find((x) => x.id === S.tool);
  const panel = t ? `<div class="tpanel" data-testid="tool-panel"><div class="tp-h">${icon(t.icon, 'sm')}<b>${t.name}</b><span class="grow"></span><button class="ib" data-a="tool" data-id="${t.id}" aria-label="收起面板">${icon('x', 'sm')}</button></div>
    <div class="tp-note" data-testid="tool-panel-source">${t.pkg}。此面板内容由插件提供，默认呈现只负责承载位置。</div><div class="tp-b">${toolBody(t.id)}</div></div>` : '';
  $('#inspector').innerHTML = panel + `<nav class="abar" aria-label="插件工具面板" data-testid="tool-bar">${TOOLS.map((x) => `<button class="ib ${S.tool === x.id ? 'on' : ''}" data-a="tool" data-id="${x.id}" data-tip="${x.name}（${x.pkg}）" aria-label="${x.name}">${icon(x.icon)}</button>`).join('')}<span style="flex:1"></span><button class="ib" data-a="settingsPage" data-p="packages" data-sub="cap" data-tip="管理能力包" aria-label="管理能力包">${icon('plus')}</button></nav>`;
}
function renderInspector() {}
function activityHtml(s) {
  const a = ag(s); let h = '<div class="sec"><div class="sec-h">' + icon('target', 'sm') + '持续目标</div>';
  if (!s.goal) {
    h += a.goal ? `<div class="hint" style="margin-bottom:8px">没有进行中的目标。目标由 Agent 原生推进，与单次回合和消息队列分开。</div><button class="btn sm" data-a="setGoal">${icon('plus', 'sm')}设置目标</button>`
                : `<div class="hint" style="margin-bottom:8px">没有进行中的目标。</div><span data-tip="${esc(a.goalWhy)}"><button class="btn sm is-disabled">${icon('plus', 'sm')}设置目标</button></span>`;
  } else {
    const g = s.goal;
    const gs = { active: ['推进中', 'accent'], paused: ['已暂停', 'attention'], pausing: ['暂停请求已发送', 'unknown'], resuming: ['恢复请求已发送', 'unknown'], clearing: ['清除请求已发送', 'unknown'], done: ['已完成', 'success'] }[g.status];
    h += `<div class="goal"><div style="display:flex;align-items:center;gap:6px"><span class="tag ${gs[1]}">${gs[0]}</span><span class="muted small">开始于 ${g.since}</span></div>
      <div class="gt">${esc(g.text)}</div>
      <dl class="kv">${g.budget.map(([kk, v, src]) => `<dt>${kk}</dt><dd>${v === '未报告' ? '<span class="muted">未报告</span>' : `${v} <span class="muted">· ${src}</span>`}</dd>`).join('')}</dl>
      <div class="acts" style="display:flex;gap:6px;margin-top:10px;flex-wrap:wrap">
        ${g.status === 'active' ? (a.goalPause ? `<span data-tip="只暂停目标，不停止当前回合和队列。"><button class="btn sm" data-a="goalPause">${icon('pause', 'sm')}暂停目标</button></span>` : `<span data-tip="${esc(a.goalPauseWhy)}"><button class="btn sm is-disabled">${icon('pause', 'sm')}暂停目标</button></span>`) : ''}
        ${g.status === 'paused' ? `<span data-tip="恢复目标可能立即开始新回合；不会恢复消息队列。"><button class="btn sm" data-a="goalResume">${icon('play', 'sm')}恢复目标</button></span>` : ''}
        <button class="btn sm danger" data-a="goalClear">清除</button></div></div>`;
  }
  h += '</div><div class="sec"><div class="sec-h">' + icon('bot', 'sm') + '子 Agent</div>';
  h += s.subagents.length ? s.subagents.map((x) => `<div class="irow" data-a="drill" data-type="sub" data-id="${x.id}"><span class="dot ${x.status === 'running' ? 'running' : 'idle'}"></span><span class="grow">${esc(x.name)}</span>${x.bg ? '<span class="tag">后台</span>' : ''}<span class="muted small">${x.status === 'running' ? x.since : '已完成'}</span>${icon('right', 'sm')}</div>`).join('') : '<div class="hint">本会话没有子 Agent。</div>';
  h += '</div><div class="sec"><div class="sec-h">' + icon('term', 'sm') + '后台任务</div>';
  h += s.tasks.length ? s.tasks.map((t) => `<div class="irow" data-a="drill" data-type="task" data-id="${t.id}">${t.status === 'running' ? '<span class="dot running"></span>' : t.status === 'cancelling' ? '<span class="dot unknown"></span>' : '<span class="dot idle"></span>'}<span class="grow mono">${esc(t.cmd)}</span><span class="muted small">${{ running: '运行中', cancelling: '取消中', cancelled: '已取消' }[t.status]}</span>${icon('right', 'sm')}</div>`).join('') : '<div class="hint">本会话没有后台任务。</div>';
  return h + '</div>';
}
function drillHtml(s) {
  const d = S.drill;
  if (d.type === 'sub') {
    const sa = s.subagents.find((x) => x.id === d.id);
    return `<div data-testid="detail-view"><div class="crumb"><a data-a="undrill" href="#" role="button" data-testid="detail-back">${icon('left', 'sm')}返回会话</a>${icon('right', 'sm')}子 Agent · 只读</div>
      <div style="display:flex;align-items:center;gap:8px;margin-bottom:10px"><span class="dot ${sa.status === 'running' ? 'running' : 'idle'}"></span><b class="grow">${esc(sa.name)}</b></div>
      <div class="mini">${sa.items.map((it, i) => `<div class="it">${itemHtml(s, it, 'sa' + i)}</div>`).join('')}</div>
      <div class="hint" style="margin-top:8px">父会话：${esc(s.title)} · 按 Esc 返回</div></div>`;
  }
  const t = s.tasks.find((x) => x.id === d.id);
  return `<div data-testid="detail-view"><div class="crumb"><a data-a="undrill" href="#" role="button" data-testid="detail-back">${icon('left', 'sm')}返回会话</a>${icon('right', 'sm')}后台任务</div>
    <div class="mono" style="margin-bottom:6px">${esc(t.cmd)}</div>
    <dl class="kv" style="margin-bottom:10px"><dt>状态</dt><dd>${{ running: '运行中', cancelling: '<span style="color:var(--unknown)">取消请求已发送，等待 Agent 报告结束</span>', cancelled: '已取消（Agent 报告）' }[t.status]}</dd><dt>开始</dt><dd>${t.since}</dd><dt>最后输出</dt><dd>${t.last}</dd></dl>
    <pre class="term mono">${esc(t.out)}${t.status === 'running' ? '▍' : ''}</pre>
    <div style="display:flex;gap:6px;margin-top:10px">${t.status === 'running' ? `<button class="btn sm danger" data-a="cancelTask" data-id="${t.id}">取消任务</button>` : ''}<span class="hint">取消只影响这一个任务，不会停止当前回合。</span></div></div>`;
}
function queueHtml(s) {
  const a = ag(s);
  if (!s.queue.length) return `<div class="hint">没有排队的消息。执行期间按 Enter 发送的消息会在这里排队，按顺序投递。</div>`;
  let h = `<div class="qhead ${s.queuePaused ? 'paused' : ''}">${icon(s.queuePaused ? 'pause' : 'play', 'sm')}<span class="grow">${s.queuePaused ? esc(s.queuePaused) : '按顺序推进：当前回合结束后投递下一条'}</span>
    ${s.queuePaused ? (s.queue.some((q) => q.unknown) ? '<span data-tip="请先处理结果未知的项。">' + '<button class="btn sm is-disabled">恢复</button></span>' : '<button class="btn sm" data-a="resumeQueue">恢复</button>') : '<button class="btn sm" data-a="pauseQueue">暂停</button>'}
    <button class="btn sm ghost" data-a="clearQueue">清空</button></div>`;
  h += s.queue.map((q, i) => q.unknown ? `<div class="qitem unknown" data-testid="queue-item" data-unknown="true"><div class="qm" style="margin:0 0 4px;color:var(--unknown)">${icon('help', 'sm')}投递结果未知</div><div class="qt">${esc(q.text)}</div>
      <div class="qm">${esc(q.why)}</div><div class="qa"><button class="btn sm" data-a="verify" data-q="${q.id}">核实</button><button class="btn sm" data-a="resend" data-q="${q.id}">重新发送（可能重复执行）</button></div></div>`
    : `<div class="qitem" draggable="true" data-testid="queue-item" data-unknown="false"><div class="qt">${esc(q.text)}</div><div class="qm">${icon('grip', 'sm')}#${i + 1} · 入队 ${q.at}</div>
      <div class="qa"><button class="btn sm" data-a="sendNow" data-q="${q.id}" data-tip="${s.running ? (a.steer ? '当前正在执行：将以插话方式发送此条。' : '当前正在执行且不支持插话：需要先停止。') : '立即发送此条，其余保持原状态。'}">立即发送</button>
      <button class="btn sm ghost" data-a="editQ" data-q="${q.id}">编辑</button><button class="btn sm ghost" data-a="delQ" data-q="${q.id}">删除</button></div></div>`).join('');
  return h + `<div class="hint" style="margin-top:8px">恢复队列不会恢复持续目标；暂停队列也不会暂停目标或取消后台任务。</div>`;
}
function contextHtml(s) {
  const a = ag(s);
  const refs = s.id === 's1' ? [{ t: '队列与投递证据', at: '11-04 14:02', g: '有效' }] : [];
  let h = `<div class="sec"><div class="sec-h">${icon('hash', 'sm')}本会话发出的引用</div>`;
  h += refs.length ? refs.map((r) => `<div class="irow"><span class="grow">${esc(r.t)}</span><span class="muted small">固定 ${r.at}</span><span class="tag success">${r.g}</span></div>`).join('') : '<div class="hint">尚未发出引用。</div>';
  h += `</div><div class="sec"><div class="sec-h">${icon('layers', 'sm')}上下文与用量</div><dl class="kv">
    <dt>上下文窗口</dt><dd>${s.ctx ? `${s.ctx[0]}k / ${s.ctx[1]}k（Agent 报告）` : '<span class="muted">未报告</span>'}</dd>
    <dt>本会话 Token</dt><dd>${s.ctx ? '输入 412k · 输出 38k' : '<span class="muted">未报告</span>'}</dd><dt>费用</dt><dd><span class="muted">未报告</span></dd></dl></div>
    <div class="sec"><div class="sec-h">${icon('bot', 'sm')}会话信息</div><dl class="kv">
    <dt>Agent</dt><dd>${esc(a.name)}</dd><dt>适配器</dt><dd>${esc(a.adapter)}</dd>
    <dt>验证状态</dt><dd>${a.verified ? '<span class="tag success">已验证组合</span>' : '<span class="tag unknown">未验证组合</span>'}</dd>
    <dt>原生会话</dt><dd class="mono">ses_${s.id}7f3a…e21 <button class="ib" data-a="toast" data-t="已复制原生会话 ID" aria-label="复制">${icon('copy', 'sm')}</button></dd>
    <dt>会话绑定</dt><dd>${s.recover && s.recover.state !== 'done' && s.recover.state !== 'connected' ? '<span style="color:var(--unknown)">待恢复连接</span>' : '已连接'}</dd><dt>工作区</dt><dd class="mono">${esc(wsOf(s).path)}</dd></dl></div>`;
  return h;
}

/* ---------- other views ---------- */
function renderOther() {
  const el = $('#otherView');
  if (S.view === 'archive') {
    const list = S.sessions.filter((s) => s.archived);
    el.innerHTML = `<div class="arch"><h2><button class="btn sm ghost" data-a="back">${icon('left', 'sm')}返回</button>已归档会话</h2>
      ${list.length ? list.map((s) => `<div class="pkg">${mark(ag(s))}<div class="grow"><div>${esc(s.title)}</div><div class="hint">${esc(wsOf(s).name)} · ${s.updated}</div></div><button class="btn sm" data-a="open" data-id="${s.id}">阅读</button><button class="btn sm" data-a="unarchive" data-id="${s.id}">取消归档</button></div>`).join('') : '<div class="hint">没有已归档的会话。</div>'}
      <div class="hint">归档会话可以阅读、搜索，也可以在搜索结果中带“已归档”标签出现。</div></div>`;
  } else if (S.view === 'settings') el.innerHTML = settingsHtml();
  else if (S.view === 'welcome') el.innerHTML = `<div class="welcome"><h2 style="font-size:20px;margin:0 0 6px">欢迎使用 Aibo</h2><p class="muted" style="margin:0 0 16px">添加一个项目目录作为工作区，然后选择 Agent 开始会话。</p>
      <button class="btn primary" data-a="addWs">${icon('folder', 'sm')}添加工作区</button>
      <div class="sec" style="margin-top:28px"><div class="sec-h">检测到的 Agent</div>${Object.values(AGENTS).map((a) => `<div class="pkg">${mark(a)}<div class="grow"><div>${a.name} ${a.verified ? '' : '<span class="tag unknown">未验证</span>'}</div><div class="hint ${a.ok ? '' : 'warn'}">${a.ok ? `${a.adapter} · 可执行文件与认证检查通过` : a.okWhy}</div></div>${a.ok ? `<span class="st-done">${icon('check')}</span>` : '<button class="btn sm" data-a="settingsPage" data-p="agents">打开 Agent 设置</button>'}</div>`).join('')}</div>
      <button class="btn sm ghost" data-a="back">（原型）返回</button></div>`;
}
function settingsHtml() {
  const P = S.settingsPage;
  const nav = [['general', '通用'], ['appearance', '外观'], ['workspaces', '工作区'], ['agents', 'Agent'], ['packages', '扩展包'], ['refs', '引用'], ['diag', '诊断']];
  const row = (t, d, ctl) => `<div class="srowx"><div class="lab"><div>${t}</div>${d ? `<div class="hint">${d}</div>` : ''}</div>${ctl}</div>`;
  const seg = (key, opts, val) => `<div class="seg">${opts.map((o) => `<button class="${o === val ? 'on' : ''}" data-a="setOpt" data-k="${key}" data-v="${o}">${o}</button>`).join('')}</div>`;
  let pg = '';
  if (P === 'general') pg = `<h2>通用</h2>${row('界面语言', '切换后立即生效。', seg('lang', ['中文', 'English', '跟随系统'], S.lang))}
    ${row('系统通知', '窗口未聚焦时，审批、追问和回合失败发送通知。', seg('notify', ['开', '关'], S.notify ? '开' : '关'))}
    <div class="sec-h" style="margin-top:22px">操作确认</div>
    ${row('归档有活动的会话', '', seg('c1', ['确认', '不确认'], '确认'))}${row('清空队列', '', seg('c2', ['确认', '不确认'], '确认'))}
    ${row('重新发送结果未知的消息', '可能导致 Agent 重复执行，始终需要确认。', '<span class="tag">始终确认</span>')}
    <div class="sec-h" style="margin-top:22px">快捷键</div><div class="hint">⌘K 搜索 · ⌘N 新建会话 · ⌘1–9 切换会话 · ⌘B 侧栏 · ⌥⌘I 工具面板 · ⌘. 停止 · ⌘⏎ 插话 · ⌘⇧A 待处理请求 · ⌘⇧R 重新加载呈现</div><div class="hint" style="margin-top:4px">第一版不支持自定义快捷键。</div>`;
  if (P === 'appearance') pg = `<h2>外观</h2>${row('主题', '', seg('theme', ['浅色', '深色', '跟随系统'], S.themePref))}
    ${row('强调色', '只替换强调色；状态颜色保持不变。', `<div class="swatches">${['#4C5BD4', '#2F7A6A', '#B5543A', '#8A4FB8', '#3D7BC9', '#5A5A60'].map((c) => `<span class="swatch ${S.accent === c ? 'on' : ''}" style="background:${c}" data-a="accent" data-c="${c}" role="button" aria-label="强调色 ${c}"></span>`).join('')}</div>`)}
    ${row('会话正文字号', '', seg('fs', ['13', '14', '15'], '14'))}
    ${row('当前呈现', '默认呈现 1.0.0（内置）。外部呈现包在“扩展包 › 呈现包”中管理。', '<button class="btn sm" data-a="settingsPage" data-p="packages" data-sub="pres">管理呈现包</button>')}`;
  if (P === 'workspaces') pg = `<h2>工作区</h2>${S.workspaces.map((w) => `<div class="pkg">${icon('folder')}<div class="grow"><div>${esc(w.name)}</div><div class="hint mono">${esc(w.path)}</div></div>${w.trusted ? '<span class="tag success">已信任</span>' : '<span class="tag attention">未信任</span>'}<button class="btn sm" data-a="${w.trusted ? 'revokeTrust' : 'trust'}" data-id="${w.id}">${w.trusted ? '撤销信任' : '信任'}</button><button class="btn sm ghost danger" data-a="removeWs" data-id="${w.id}">移除</button></div>`).join('')}`;
  if (P === 'agents') pg = `<h2>Agent</h2>${Object.values(AGENTS).map((a) => `<div class="pkg">${mark(a)}<div class="grow"><div>${a.name} ${a.verified ? '<span class="tag success">已验证</span>' : '<span class="tag unknown">未验证组合</span>'}</div>
    <div class="hint">${a.adapter}</div><div class="hint ${a.ok ? '' : 'warn'}">${a.ok ? '可执行文件 ✓　依赖 ✓　认证 ✓' : a.okWhy}</div></div><button class="btn sm" data-a="toast" data-t="正在重新检测 ${a.name}…">重新检测</button></div>`).join('')}
    <button class="btn sm" data-a="toast" data-t="（原型）打开通用 ACP Agent 接入模板">${icon('plus', 'sm')}接入其他 ACP Agent</button>`;
  if (P === 'packages') {
    const T = S.pkgTab;
    const lists = {
      agent: [['Claude Code 接入', '1.3.0', '5 个会话使用'], ['Codex 接入', '0.9.1', '8 个会话使用'], ['Gemini CLI（通用模板）', '0.1.0', '1 个会话使用']],
      cap: [['会话历史读取工具', '1.0.0', '提供全文读取 MCP 工具'], ['Figma 上下文', '0.4.2', '提供设计稿上下文']],
      pres: [['默认呈现', '1.0.0', '内置 · 当前使用'], ['Compact Pro', '1.2.0', '外部 · Solid'], ['Zen Reader', '0.3.1', '外部 · Vue']],
    }[T];
    pg = `<h2>扩展包</h2><div class="subtabs">${[['agent', 'Agent 包'], ['cap', '能力包'], ['pres', '呈现包']].map(([kk, n]) => `<button class="${T === kk ? 'on' : ''}" data-a="pkgTab" data-t="${kk}">${n}</button>`).join('')}</div>
      <div style="margin-bottom:10px"><button class="btn sm" data-a="toast" data-t="（原型）选择本地包目录">${icon('plus', 'sm')}从本地安装</button></div>
      ${lists.map(([n, v, d], i) => `<div class="pkg">${icon(T === 'pres' ? 'layers' : T === 'cap' ? 'tool' : 'bot')}<div class="grow"><div>${n} <span class="muted small">${v}</span></div><div class="hint">${d}</div></div>
      ${T === 'pres' ? (i === 0 ? '<span class="tag accent">当前</span>' : `<button class="btn sm" data-a="usePres" data-n="${n}">设为当前</button>`) : '<button class="btn sm ghost">停用</button>'}
      ${T === 'agent' ? `<button class="btn sm ghost" data-a="upgradePkg" data-n="${n}">升级…</button>` : ''}${i === 0 && T === 'pres' ? '' : `<button class="btn sm ghost danger" data-a="uninstallPkg" data-n="${n}" data-t="${T}">卸载</button>`}</div>`).join('')}`;
  }
  if (P === 'refs') pg = `<h2>引用</h2><div class="hint" style="margin-bottom:8px">引用在加入输入框时固定版本。下列为默认附带的摘录范围；超出部分按省略标记截断，可由 Agent 按需读取同版本全文。</div>
    ${row('默认摘录轮数', '最近几轮原文及关联工具记录。', `<input class="inp" type="number" value="${S.refDefaults.turns}" data-ref="turns">`)}
    ${row('单条上限（字符）', '', `<input class="inp" type="number" value="${S.refDefaults.per}" data-ref="per">`)}
    ${row('单次输入合计上限（字符）', '', `<input class="inp" type="number" value="${S.refDefaults.total}" data-ref="total">`)}`;
  if (P === 'diag') pg = `<h2>诊断</h2>${row('日志占用', '保留最长 14 天、总量最多 100 MiB，任一达到后清理最旧日志。', '<span>23.4 MiB / 100 MiB</span>')}
    ${row('日志内容', '结构化事件、时间、关联 ID、状态与处理后的错误说明；不包含完整业务正文。', '')}
    ${row('日志目录', '<span class="mono">~/Library/Logs/Aibo</span>', '<button class="btn sm" data-a="toast" data-t="（原型）在 Finder 中打开日志目录">打开</button>')}`;
  return `<div class="set"><nav><button class="btn sm ghost back" data-a="back">${icon('left', 'sm')}返回会话 <span class="kbd">Esc</span></button>
    ${nav.map(([kk, n]) => `<div class="n ${P === kk ? 'on' : ''}" data-a="settingsPage" data-p="${kk}">${n}</div>`).join('')}</nav><div class="pg">${pg}</div></div>`;
}

/* ---------- compose ---------- */
function renderAll(scroll) {
  document.documentElement.dataset.theme = S.theme;
  document.documentElement.style.setProperty('--accent', S.accent && S.theme === 'light' ? S.accent : '');
  const app = $('#app');
  app.classList.toggle('no-sidebar', !!S.sidebarHidden);
  app.classList.toggle('no-inspector', S.view === 'welcome');
  renderSidebar(); renderBanners();
  const inSession = S.view === 'session';
  $('#sessionView').hidden = !inSession; $('#otherView').hidden = inSession;
  if (inSession) { renderHead(); renderStream(scroll); renderPending(); renderChips(); renderCbar(); renderSettingsBar(); }
  else renderOther();
  renderTools();
}
function renderLive(scroll) {
  renderSidebar(); renderBanners();
  if (S.view !== 'session') return;
  renderHead(); renderStream(scroll); renderPending(); renderCbar(); renderSettingsBar(); renderInspector();
}
