<script lang="ts">
  import { onMount } from 'svelte';
  import type { Session } from '../core/session';
  import { markdownStats } from '../core/markdown';

  let { session, itemCount, imeLog, theme = $bindable() }: { session: Session; itemCount: number; imeLog: string[]; theme: 'light' | 'dark' } = $props();

  let open = $state(true);
  // svelte-ignore state_referenced_locally
  let interval = $state(session.publishInterval);
  // svelte-ignore state_referenced_locally
  let rate = $state(session.streamRate);
  let m = $state({ fps: 0, longFrames: 0, maxGap: 0, latAvg: 0, latMax: 0, rows: 0, nodes: 0, heap: '不可用', mdRenders: 0 });

  $effect(() => { session.publishInterval = interval; });
  $effect(() => { session.streamRate = rate; });

  let longReset = () => {};
  onMount(() => {
    let raf = 0, last = performance.now(), frames = 0, windowStart = last, seenVersion = 0;
    let longFrames = 0, maxGap = 0;
    const lats: number[] = [];
    longReset = () => { longFrames = 0; lats.length = 0; };
    const loop = (t: number) => {
      const gap = t - last;
      last = t; frames++;
      if (gap > 50) longFrames++;
      if (gap > maxGap) maxGap = gap;
      const snap = session.getSnapshot();
      if (snap.version !== seenVersion) {
        seenVersion = snap.version;
        const lat = performance.now() - snap.publishedAt;
        lats.push(lat);
        if (lat > 50) {
          const last = snap.items[snap.items.length - 1];
          // 诊断用：记录慢发布时的上下文，供自动化脚本读取
          const w = window as unknown as { __slow?: unknown[] };
          (w.__slow ??= []).push({ lat: Math.round(lat), version: snap.version, items: snap.items.length, last: last?.kind, len: last && last.kind === 'agent' ? last.text.length : null, running: snap.running, at: Math.round(performance.now()) });
        }
        if (lats.length > 120) lats.shift();
      }
      if (t - windowStart > 500) {
        const mem = (performance as unknown as { memory?: { usedJSHeapSize: number } }).memory;
        const s = markdownStats();
        m = {
          fps: Math.round((frames * 1000) / (t - windowStart)),
          longFrames, maxGap: Math.round(maxGap),
          latAvg: lats.length ? Math.round(lats.reduce((a, b) => a + b, 0) / lats.length) : 0,
          latMax: lats.length ? Math.round(Math.max(...lats)) : 0,
          rows: document.querySelectorAll('.stream .row').length,
          nodes: document.querySelector('.stream')?.getElementsByTagName('*').length ?? 0,
          heap: mem ? `${Math.round(mem.usedJSHeapSize / 1048576)} MB` : '不可用（非 Chromium）',
          mdRenders: s.renders,
        };
        frames = 0; windowStart = t; maxGap = 0;
      }
      raf = requestAnimationFrame(loop);
    };
    raf = requestAnimationFrame(loop);
    return () => cancelAnimationFrame(raf);
  });
</script>

<aside class="perf" class:closed={!open} aria-label="验证面板">
  <button class="perf-h" onclick={() => (open = !open)}>验证面板（非产品界面）{open ? '▾' : '▸'}</button>
  {#if open}
    <div class="perf-b">
      <div class="row2">
        <button class="btn sm" onclick={() => session.loadAllHistory()}>载入全部历史</button>
        <button class="btn sm" onclick={() => session.startLongReply()}>开始长回复</button>
        <button class="btn sm" onclick={() => { longReset(); }}>清零</button>
        <button class="btn sm" onclick={() => (theme = theme === 'dark' ? 'light' : 'dark')}>{theme === 'dark' ? '浅色' : '深色'}</button>
      </div>
      <label>发布间隔
        <select bind:value={interval}><option value={0}>每次立即</option><option value={16}>16ms</option><option value={50}>50ms</option><option value={100}>100ms</option></select>
      </label>
      <label>流速
        <select bind:value={rate}><option value={2}>2 字/16ms</option><option value={6}>6 字/16ms</option><option value={20}>20 字/16ms</option><option value={80}>80 字/16ms</option></select>
      </label>
      <dl>
        <dt>条目总数</dt><dd>{itemCount}</dd>
        <dt>已挂载行</dt><dd>{m.rows}</dd>
        <dt>会话流 DOM 节点</dt><dd>{m.nodes}</dd>
        <dt>帧率</dt><dd>{m.fps} fps</dd>
        <dt>&gt;50ms 的帧</dt><dd>累计 {m.longFrames}（0.5s 内最长 {m.maxGap}ms）</dd>
        <dt>发布→下一帧</dt><dd>平均 {m.latAvg}ms · 最大 {m.latMax}ms</dd>
        <dt>Markdown 块渲染</dt><dd>{m.mdRenders} 次</dd>
        <dt>JS 堆</dt><dd>{m.heap}</dd>
      </dl>
      <div class="ime">
        <div class="mh">IME 事件（最近 8 条）</div>
        {#each imeLog as l, i (i)}<div class="mono">{l}</div>{:else}<div class="hint">在输入框中用中文输入法输入，再按 Enter 确认候选词。</div>{/each}
      </div>
    </div>
  {/if}
</aside>
