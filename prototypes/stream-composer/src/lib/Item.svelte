<script lang="ts">
  import type { Item, Session } from '../core/session';
  import Markdown from './Markdown.svelte';

  let { item, expanded, session }: { item: Item; expanded: boolean; session: Session } = $props();

  const DELIVERY = { queued: '已排队', delivering: '投递中…', received: '已接收', unknown: '结果未知' } as const;

  function onApprovalKey(e: KeyboardEvent) {
    if (item.kind !== 'approval' || item.state !== 'pending') return;
    const n = Number(e.key);
    if (n >= 1 && n <= item.options.length) session.approve(item.id, item.options[n - 1]);
  }
</script>

{#if item.kind === 'user'}
  <div class="bubble-wrap">
    <div class="bubble" class:steer={item.steer}>
      {#if item.steer}<span class="tag accent">插话</span>{/if}
      {#if item.chips.length}<div class="chips">{#each item.chips as c (c.id)}<span class="chip">{c.type === 'ref' ? '#' : c.type === 'file' ? '@' : '🖼'} {c.label}</span>{/each}</div>{/if}
      {item.text}
    </div>
    <div class="dlv" class:unknown={item.delivery === 'unknown'}>{DELIVERY[item.delivery]}</div>
  </div>
{:else if item.kind === 'agent'}
  <div class="role"><span class="agent-mark">CC</span>Claude Code</div>
  <Markdown text={item.text} done={item.done} />
{:else if item.kind === 'tool'}
  <div class="fold">
    <button class="fh" onclick={() => session.toggleExpanded(item.id)} aria-expanded={expanded}>
      <b>{item.name}</b><span class="tgt mono">{item.target}</span>
      <span class="st {item.status}">{item.status === 'running' ? '执行中' : item.status === 'done' ? '✓' : '✕'}</span>
      <span class="muted small">{item.dur ?? ''}</span><span aria-hidden="true">{expanded ? '▾' : '▸'}</span>
    </button>
    {#if expanded}<pre class="fb mono">{item.output || '（无输出）'}</pre>{/if}
  </div>
{:else if item.kind === 'approval'}
  <!-- 审批卡可聚焦后按数字键选择选项（设计文档 §7.1） -->
  <!-- svelte-ignore a11y_no_noninteractive_tabindex, a11y_no_noninteractive_element_interactions -->
  <div class="card attention" tabindex="0" role="group" aria-label="需要审批" onkeydown={onApprovalKey} data-approval={item.id}>
    <div class="ct">需要审批</div>
    <div class="cmd mono"><span class="muted">{item.tool}</span> {item.cmd}</div>
    <div class="acts">
      {#if item.state === 'pending'}
        {#each item.options as o, i (o)}
          <button class="btn sm" class:primary={i === 0} onclick={() => session.approve(item.id, o)}><span class="kbd">{i + 1}</span>{o}</button>
        {/each}
      {:else if item.state === 'submitted'}
        <span class="tag unknown">已提交：{item.choice}，等待 Agent 确认</span>
      {:else}
        <span class="tag">已答复：{item.choice}</span>
      {/if}
    </div>
  </div>
{/if}
