<script lang="ts">
  import { tick, untrack } from 'svelte';
  import { VList, type VListHandle } from 'virtua/svelte';
  import type { Item, Session } from '../core/session';
  import ItemView from './Item.svelte';

  let { items, expanded, olderRemaining, session }: {
    items: readonly Item[];
    expanded: ReadonlySet<string>;
    olderRemaining: number;
    session: Session;
  } = $props();

  let list = $state<VListHandle>();
  /** 是否跟随底部：用户向上滚动后停止跟随 */
  let follow = true;
  let shift = $state(false);
  let unseen = $state(false);
  let loading = false;

  const atBottom = () => !!list && list.getScrollOffset() + list.getViewportSize() >= list.getScrollSize() - 48;
  const toEnd = () => { if (items.length) list?.scrollToIndex(items.length - 1, { align: 'end' }); };

  function onscroll() {
    if (!list) return;
    // 任何滚动（滚轮、拖动滚动条、键盘、程序定位）都按“是否仍在底部”重新决定是否跟随；
    // 流式增长本身不触发 scroll 事件，因此不会误判
    follow = atBottom();
    if (follow) unseen = false;
    if (list.getScrollOffset() < 400 && olderRemaining > 0 && !loading) loadOlder();
  }
  // 用户主动向上滚动时立即停止跟随，避免与流式自动滚动互相争夺
  function onwheel(e: WheelEvent) { if (e.deltaY < 0) follow = false; }
  function onkeydown(e: KeyboardEvent) { if (['ArrowUp', 'PageUp', 'Home'].includes(e.key)) follow = false; }

  async function loadOlder() {
    loading = true;
    shift = true;
    session.loadOlder();
    await tick();
    requestAnimationFrame(() => { shift = false; loading = false; });
  }

  // 末尾条目变化（新条目或流式增长）时：跟随则滚到底，否则提示有新内容
  let lastId: string | undefined;
  let lastLen = 0;
  $effect(() => {
    const last = items[items.length - 1];
    const id = last?.id;
    const len = last && last.kind === 'agent' ? last.text.length : 0;
    untrack(() => {
      const changed = id !== lastId || len !== lastLen;
      lastId = id; lastLen = len;
      if (!changed) return;
      if (follow) requestAnimationFrame(toEnd);
      else unseen = true;
    });
  });

  // 推迟到下一帧，避免在 ResizeObserver 回调里同步改滚动位置引发 RO 循环警告
  function onresize() { if (follow) requestAnimationFrame(toEnd); }
  function jumpToEnd() { follow = true; unseen = false; toEnd(); }

  /** 搜索命中等场景：定位到指定条目 */
  export function scrollToId(id: string) {
    const i = items.findIndex((x) => x.id === id);
    if (i < 0) return;
    follow = false;
    list?.scrollToIndex(i, { align: 'center' });
  }
</script>

<!-- svelte-ignore a11y_no_static_element_interactions -->
<div class="stream-wrap" {onwheel} {onkeydown}>
  <VList bind:this={list} data={items} getKey={(it) => it.id} {shift} {onscroll} {onresize} class="stream" role="log" aria-live="polite" tabindex={0}>
    {#snippet children(item)}
      <div class="row" data-id={item.id}>
        {#if item === items[0] && olderRemaining > 0}<div class="older">正在加载更早的记录…（还有 {olderRemaining} 条）</div>{/if}
        <ItemView {item} expanded={expanded.has(item.id)} {session} />
      </div>
    {/snippet}
  </VList>
  {#if unseen}<button class="new-content" onclick={jumpToEnd}>↓ 有新内容</button>{/if}
</div>
