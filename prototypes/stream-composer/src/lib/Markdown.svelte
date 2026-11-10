<script lang="ts">
  import { toBlocks } from '../core/markdown';
  import CodeBlock from './CodeBlock.svelte';

  let { text, done }: { text: string; done: boolean } = $props();
  const blocks = $derived(toBlocks(text));
</script>

<div class="md">
  {#each blocks as b, i (i)}
    {#if b.kind === 'md'}
      {@html b.html}
    {:else}
      <CodeBlock code={b.code} lang={b.lang} closed={b.closed} />
    {/if}
  {/each}
  {#if !done}<span class="caret" aria-hidden="true"></span>{/if}
</div>
