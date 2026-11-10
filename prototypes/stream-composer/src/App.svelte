<script lang="ts">
  import { Tooltip } from 'bits-ui';
  import { Session } from './core/session';
  import { SessionStore } from './lib/store.svelte';
  import Stream from './lib/Stream.svelte';
  import Composer from './lib/Composer.svelte';
  import PerfPanel from './lib/PerfPanel.svelte';
  import { preloadHighlighter } from './core/highlight';
  import { onMount } from 'svelte';

  const session = new Session();
  session.seedHistory(10000);
  const store = new SessionStore(session);
  const snap = $derived(store.snap);

  onMount(() => { const t = setTimeout(preloadHighlighter, 1200); return () => clearTimeout(t); });

  let theme = $state<'light' | 'dark'>('light');
  $effect(() => { document.documentElement.dataset.theme = theme; });

  let imeLog = $state<string[]>([]);
  const onIme = (line: string) => { imeLog = [...imeLog.slice(-7), `${new Date().toTimeString().slice(0, 8)} ${line}`]; };

  let stream = $state<ReturnType<typeof Stream>>();
  function onkeydown(e: KeyboardEvent) {
    const mod = e.metaKey || e.ctrlKey;
    if (mod && e.key === '.') { e.preventDefault(); session.stop(); }
    if (mod && e.shiftKey && e.key.toLowerCase() === 'a') {
      e.preventDefault();
      const id = snap.items.find((i) => i.kind === 'approval' && i.state === 'pending')?.id;
      if (id) { stream?.scrollToId(id); requestAnimationFrame(() => document.querySelector<HTMLElement>(`[data-approval="${id}"]`)?.focus()); }
    }
  }
</script>

<svelte:window {onkeydown} />

<Tooltip.Provider delayDuration={250}>
  <main class="app">
    <header class="shead">
      <span class="t">重构共享交互模块的草稿状态</span><span class="ws">aibo-rev</span><span class="sp"></span>
      <span class="run-state" class:running={snap.running}>{snap.running ? '执行中' : '空闲'}</span>
      <span class="agent-id"><span class="agent-mark">CC</span>Claude Code</span>
    </header>
    <Stream bind:this={stream} items={snap.items} expanded={snap.expanded} olderRemaining={snap.olderRemaining} {session} />
    <Composer {session} {snap} {onIme} />
  </main>
  <PerfPanel {session} itemCount={snap.items.length + snap.olderRemaining} {imeLog} bind:theme />
</Tooltip.Provider>
