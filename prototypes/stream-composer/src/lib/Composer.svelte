<script lang="ts">
  import { DropdownMenu, Popover, Select, Tooltip } from 'bits-ui';
  import type { Chip, Session, Snapshot } from '../core/session';

  let { session, snap, onIme }: { session: Session; snap: Snapshot; onIme: (line: string) => void } = $props();

  const FILES = ['src/core/session-control.ts', 'src/core/queue.ts', 'src/core/delivery-evidence.ts', 'src/core/', 'src/presentation/shared/draft.ts', 'src/presentation/default/', 'docs/design/ui-interaction.md', 'package.json', 'go.mod'];
  const SESSIONS = ['队列与投递证据', '修复 Windows 窗口缩放', '整理数据生命周期盘点'];
  const COMMANDS = [['/goal', '设置持续目标 · Aibo'], ['/review', '审查当前改动 · Agent'], ['/compact', '压缩上下文 · Agent'], ['/skill:write-tests', '为改动补充测试 · Skill']];
  const MODELS = ['Claude Opus 5.5', 'Claude Sonnet 5.5', 'Claude Haiku 5.5'];
  const MODES = ['询问', '自动编辑', '计划', '完全访问'];
  const STEER_WHY = '此 Agent/适配器组合的插话存在回退歧义，尚未启用。可以排队，或停止后发送。';

  let ta = $state<HTMLTextAreaElement>();
  let composerEl = $state<HTMLElement>();
  // 草稿属于共通呈现状态：本地编辑值随输入写回 Session（Session 负责合并保存）
  // svelte-ignore state_referenced_locally
  let text = $state(session.getSnapshot().draft);
  let trig = $state<{ ch: string; start: number; q: string } | null>(null);
  let pickIdx = $state(0);
  let composing = false;
  let queueOpen = $state(false);

  type Pick = { label: string; desc: string; value: string };
  const picks = $derived.by<Pick[]>(() => {
    if (!trig) return [];
    const q = trig.q;
    if (trig.ch === '@') return FILES.filter((f) => f.toLowerCase().includes(q)).map((f) => ({ label: f, desc: f.endsWith('/') ? '目录' : '文件', value: f }));
    if (trig.ch === '#') return SESSIONS.filter((s) => s.toLowerCase().includes(q)).map((s) => ({ label: s, desc: '加入时固定版本', value: s }));
    return COMMANDS.filter(([c]) => c.includes(q)).map(([c, d]) => ({ label: c, desc: d, value: c }));
  });
  const pending = $derived(snap.items.filter((i) => i.kind === 'approval' && i.state === 'pending'));
  const refChars = $derived(snap.chips.reduce((a, c) => a + (c.type === 'ref' ? (c.chars ?? 0) : 0), 0));
  const busy = $derived(snap.running);
  const applying = $derived(snap.settings.applying.length > 0);
  const k = (n: number) => (n >= 1000 ? (n / 1000).toFixed(1).replace(/\.0$/, '') + 'k' : String(n));

  function detect() {
    if (!ta) return;
    const pos = ta.selectionStart;
    // 查询部分允许包含 /，以便输入路径（如 @src/co）
    const m = ta.value.slice(0, pos).match(/(^|\s)([@#/])(\S*)$/);
    if (!m) { trig = null; return; }
    trig = { ch: m[2], start: pos - m[3].length - 1, q: m[3].toLowerCase() };
    pickIdx = 0;
  }
  function autosize() {
    if (!ta) return;
    ta.style.height = 'auto';
    ta.style.height = Math.min(240, ta.scrollHeight) + 'px';
  }
  function oninput() { session.setDraft(text); if (!composing) detect(); autosize(); }

  function choose(i: number) {
    const p = picks[i];
    if (!p || !trig || !ta) return;
    const end = ta.selectionStart;
    let insert = '';
    if (trig.ch === '@') session.addChip({ type: 'file', label: p.value });
    else if (trig.ch === '#') session.addChip({ type: 'ref', label: p.value, range: '近 3 轮', at: new Date().toTimeString().slice(0, 5), chars: 6400 });
    else insert = p.value + ' ';
    const start = trig.start;
    text = text.slice(0, start) + insert + text.slice(end);
    trig = null;
    session.setDraft(text);
    queueMicrotask(() => { ta?.focus(); ta?.setSelectionRange(start + insert.length, start + insert.length); });
  }

  function send() {
    const t = text.trim();
    if (!t && !snap.chips.length) return;
    if (applying) return;
    session.send(t);
    text = '';
    queueMicrotask(autosize);
  }

  function onkeydown(e: KeyboardEvent) {
    if (e.key === 'Enter' || e.key === 'Process') onIme(`keydown key=${e.key} isComposing=${e.isComposing} keyCode=${e.keyCode}`);
    if (trig && picks.length) {
      if (e.key === 'ArrowDown' || e.key === 'ArrowUp') { e.preventDefault(); pickIdx = (pickIdx + (e.key === 'ArrowDown' ? 1 : picks.length - 1)) % picks.length; return; }
      if ((e.key === 'Enter' || e.key === 'Tab') && !e.isComposing && e.keyCode !== 229) { e.preventDefault(); choose(pickIdx); return; }
      if (e.key === 'Escape') { e.preventDefault(); trig = null; return; }
    }
    // 组字期间（isComposing 或 Safari 的 keyCode 229）按 Enter 只确认候选词
    if (e.key === 'Enter' && !e.shiftKey && !e.isComposing && e.keyCode !== 229 && !composing) {
      e.preventDefault();
      if ((e.metaKey || e.ctrlKey) && busy) return; // 插话在此组合中禁用，见按钮说明
      send();
      return;
    }
    if (e.key === 'Backspace' && ta && ta.selectionStart === 0 && ta.selectionEnd === 0 && snap.chips.length) session.popChip();
  }
  function onpaste(e: ClipboardEvent) {
    const img = [...(e.clipboardData?.items ?? [])].find((i) => i.type.startsWith('image/'));
    if (img) { e.preventDefault(); session.addChip({ type: 'image', label: `粘贴的图片 ${new Date().toTimeString().slice(0, 8)}.png` }); }
  }

  function jumpApproval() {
    const el = document.querySelector<HTMLElement>(`[data-approval="${pending[0]?.id}"]`);
    el?.focus();
    el?.scrollIntoView({ block: 'center' });
  }
  function chipLabel(c: Chip) {
    return c.type === 'ref' ? `${c.label} · 固定 ${c.at} · ${c.range} · ${k(c.chars ?? 0)}` : c.label;
  }
</script>

<div class="composer-wrap">
  {#if pending.length || snap.queue.length}
    <div class="strip">
      {#if pending.length}
        <button class="schip attention" onclick={jumpApproval}>{pending.length} 个审批待处理 <span class="kbd">⌘⇧A</span></button>
      {/if}
      {#if snap.queue.length}
        <Popover.Root bind:open={queueOpen}>
          <Popover.Trigger class={['schip', snap.queuePaused && 'attention', queueOpen && 'on']}>
            队列 <span class="cnt">{snap.queue.length}</span>{snap.queuePaused ? ' · 已暂停' : ''}
          </Popover.Trigger>
          <Popover.Portal>
            <Popover.Content class="pop" side="top" align="start" sideOffset={8} customAnchor={composerEl} style="width: {composerEl?.offsetWidth ?? 600}px">
              <div class="pop-h">
                <b>消息队列</b><span class="grow"></span>
                {#if snap.queuePaused}<button class="btn sm" onclick={() => session.resumeQueue()}>恢复</button>{:else}<button class="btn sm" onclick={() => session.pauseQueue()}>暂停</button>{/if}
                <Popover.Close class="ib" aria-label="收起">✕</Popover.Close>
              </div>
              <div class="pop-b">
                {#if snap.queuePaused}<div class="qhead">{snap.queuePaused}</div>{/if}
                {#each snap.queue as q, i (q.id)}
                  <div class="qitem">
                    <div class="qt">{q.text}</div>
                    <div class="qm">#{i + 1} · 入队 {q.at}</div>
                    <div class="qa">
                      <button class="btn sm" onclick={() => session.sendQueuedNow(q.id)}>立即发送</button>
                      <button class="btn sm ghost" onclick={() => session.removeQueued(q.id)}>删除</button>
                    </div>
                  </div>
                {/each}
                <div class="hint">恢复队列不会恢复持续目标；暂停队列也不会暂停目标或取消后台任务。</div>
              </div>
            </Popover.Content>
          </Popover.Portal>
        </Popover.Root>
      {/if}
    </div>
  {/if}

  <div class="composer" bind:this={composerEl}>
    {#if snap.chips.length}
      <div class="chips">
        {#each snap.chips as c (c.id)}
          <span class="chip">
            {#if c.type === 'ref'}
              <DropdownMenu.Root>
                <DropdownMenu.Trigger class="chip-trigger" title="固定版本：来源后续的新内容不会自动加入">
                  # {chipLabel(c)}{#if c.stale}<span class="stale" title="来源有新内容，可更新"></span>{/if} ▾
                </DropdownMenu.Trigger>
                <DropdownMenu.Portal>
                  <DropdownMenu.Content class="menu" align="start" sideOffset={6}>
                    <DropdownMenu.Group>
                      <DropdownMenu.GroupHeading class="mh">固定于 {c.at} · 不会自动刷新</DropdownMenu.GroupHeading>
                    </DropdownMenu.Group>
                    <DropdownMenu.RadioGroup value={c.range} onValueChange={(v) => session.updateChip(c.id, { range: v as Chip['range'], chars: v === '全文' ? 41200 : 6400 })}>
                      <DropdownMenu.RadioItem value="近 3 轮" class="mi">{#snippet children({ checked })}<span class="check">{checked ? '✓' : ''}</span>近 3 轮原文{/snippet}</DropdownMenu.RadioItem>
                      <DropdownMenu.RadioItem value="全文" class="mi">{#snippet children({ checked })}<span class="check">{checked ? '✓' : ''}</span>全文 <span class="d">由 Agent 按需读取</span>{/snippet}</DropdownMenu.RadioItem>
                    </DropdownMenu.RadioGroup>
                    <DropdownMenu.Separator class="sep" />
                    <DropdownMenu.Item class="mi" disabled={!c.stale} onSelect={() => session.updateChip(c.id, { stale: false, at: new Date().toTimeString().slice(0, 5), chars: (c.chars ?? 0) + 2100 })}>
                      {c.stale ? '更新到最新版本' : '已是最新版本'}{#if c.stale}<span class="d">新增 2 轮</span>{/if}
                    </DropdownMenu.Item>
                    <DropdownMenu.Item class="mi" onSelect={() => session.removeChip(c.id)}>移除</DropdownMenu.Item>
                  </DropdownMenu.Content>
                </DropdownMenu.Portal>
              </DropdownMenu.Root>
            {:else}
              <span class="lbl mono" title={c.type === 'file' ? '按路径引用，发送时读取当时内容' : ''}>{c.type === 'file' ? '@' : '🖼'} {c.label}</span>
            {/if}
            <button class="x" onclick={() => session.removeChip(c.id)} aria-label="移除 {c.label}">✕</button>
          </span>
        {/each}
      </div>
    {/if}

    <textarea
      bind:this={ta}
      bind:value={text}
      rows="3"
      placeholder="输入消息。@ 文件/目录　# 引用会话　/ 命令与 Skills"
      aria-label="消息输入"
      {oninput}
      {onkeydown}
      {onpaste}
      onclick={detect}
      oncompositionstart={() => { composing = true; onIme('compositionstart'); }}
      oncompositionupdate={(e) => onIme(`compositionupdate data="${e.data}"`)}
      oncompositionend={(e) => { composing = false; onIme(`compositionend data="${e.data}"`); detect(); }}
    ></textarea>

    <div class="cbar">
      <button class="ib" onclick={() => session.addChip({ type: 'image', label: `截图 ${new Date().toTimeString().slice(0, 5)}.png` })} aria-label="添加图片">🖼</button>
      <span class="grow"></span>
      {#if refChars}<span class="budget" class:over={refChars > 32000}>引用 {k(refChars)} / 32k</span>{/if}
      {#if busy}
        <button class="btn" onclick={() => session.stop()}>■ 停止 <span class="kbd">⌘.</span></button>
        <Tooltip.Root>
          <Tooltip.Trigger class="btn is-disabled" aria-disabled="true">插话 <span class="kbd">⌘⏎</span></Tooltip.Trigger>
          <Tooltip.Portal><Tooltip.Content class="tip" sideOffset={6}>{STEER_WHY}</Tooltip.Content></Tooltip.Portal>
        </Tooltip.Root>
      {/if}
      {#if applying}
        <Tooltip.Root>
          <Tooltip.Trigger class="btn primary is-disabled" aria-disabled="true">{busy ? '排队' : '发送'}</Tooltip.Trigger>
          <Tooltip.Portal><Tooltip.Content class="tip" sideOffset={6}>正在应用执行设置，确认后才能发送，不会按旧值发送。</Tooltip.Content></Tooltip.Portal>
        </Tooltip.Root>
      {:else}
        <button class="btn primary" onclick={send}>{busy || (snap.queuePaused && snap.queue.length) ? '排队' : '发送'} <span class="kbd">⏎</span></button>
      {/if}
    </div>
  </div>

  <div class="settings-bar">
    <Select.Root type="single" value={snap.settings.model} items={MODELS.map((m) => ({ value: m, label: m }))} onValueChange={(v) => session.applySetting('model', v)}>
      <Select.Trigger class={['sel', snap.settings.applying.includes('model') && 'applying']}>{snap.settings.model} ▾</Select.Trigger>
      <Select.Portal>
        <Select.Content class="menu" sideOffset={6}>
          <div class="mh">在下一次投递时生效，不影响正在进行的回合</div>
          {#each MODELS as m (m)}
            <Select.Item value={m} label={m} class="mi">{#snippet children({ selected })}<span class="check">{selected ? '✓' : ''}</span>{m}{/snippet}</Select.Item>
          {/each}
        </Select.Content>
      </Select.Portal>
    </Select.Root>
    <Select.Root type="single" value={snap.settings.mode} items={MODES.map((m) => ({ value: m, label: m }))} onValueChange={(v) => session.applySetting('mode', v)}>
      <Select.Trigger class={['sel', snap.settings.applying.includes('mode') && 'applying']}>权限 {snap.settings.mode} ▾</Select.Trigger>
      <Select.Portal>
        <Select.Content class="menu" sideOffset={6}>
          {#each MODES as m (m)}
            <Select.Item value={m} label={m} class="mi">{#snippet children({ selected })}<span class="check">{selected ? '✓' : ''}</span>{m}{/snippet}</Select.Item>
          {/each}
        </Select.Content>
      </Select.Portal>
    </Select.Root>
    <Tooltip.Root>
      <Tooltip.Trigger class="sel disabled" aria-disabled="true">Fast</Tooltip.Trigger>
      <Tooltip.Portal><Tooltip.Content class="tip" sideOffset={6}>此 Agent 未提供服务档位选项。</Tooltip.Content></Tooltip.Portal>
    </Tooltip.Root>
    <span class="grow"></span>
    <span class="ctxmeter">上下文 38% · 76k/200k</span>
  </div>
</div>

<Popover.Root open={!!trig && picks.length > 0} onOpenChange={(o) => { if (!o) trig = null; }}>
  <Popover.Portal>
    <Popover.Content
      class="picker"
      customAnchor={composerEl}
      side="top"
      align="start"
      sideOffset={8}
      trapFocus={false}
      onOpenAutoFocus={(e) => e.preventDefault()}
      onCloseAutoFocus={(e) => e.preventDefault()}
      onInteractOutside={(e) => { if (e.target === ta) e.preventDefault(); }}
    >
      <div class="ph"><span>{trig?.ch === '@' ? '文件与目录 · 按路径引用' : trig?.ch === '#' ? '引用会话 · 加入时固定版本' : '命令与 Skills'}</span><span>↑↓ 选择 · ⏎ 确认 · Esc 关闭</span></div>
      <div role="listbox" aria-label="候选项">
        {#each picks as p, i (p.value)}
          <div role="option" tabindex="-1" aria-selected={i === pickIdx} class={['pi', i === pickIdx && 'on']} onmousedown={(e) => e.preventDefault()} onclick={() => choose(i)} onkeydown={() => {}}>
            <span class={trig?.ch === '#' ? '' : 'mono'}>{p.label}</span><span class="d">{p.desc}</span>
          </div>
        {/each}
      </div>
    </Popover.Content>
  </Popover.Portal>
</Popover.Root>
