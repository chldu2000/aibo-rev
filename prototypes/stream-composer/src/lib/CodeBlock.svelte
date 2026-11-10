<script lang="ts">
  import { highlight, highlightCached } from '../core/highlight';

  let { code, lang, closed }: { code: string; lang: string; closed: boolean } = $props();

  // 已缓存时同步取得，避免虚拟列表重新挂载时闪烁；未完成的围栏只显示纯文本
  let loaded = $state<{ key: string; html: string } | null>(null);
  const key = $derived(lang + '\u0000' + code);
  const html = $derived(closed ? (highlightCached(code, lang) ?? (loaded?.key === key ? loaded.html : null)) : null);

  $effect(() => {
    if (!closed || html) return;
    const k = key;
    let alive = true;
    highlight(code, lang).then((h) => { if (alive) loaded = { key: k, html: h }; });
    return () => { alive = false; };
  });

  let copied = $state(false);
  async function copy() {
    await navigator.clipboard?.writeText(code);
    copied = true;
    setTimeout(() => (copied = false), 1200);
  }
</script>

<div class="codeblock">
  <div class="ch"><span>{lang}{closed ? '' : ' · 生成中'}</span><button class="ib" onclick={copy} aria-label="复制代码">{copied ? '已复制' : '复制'}</button></div>
  {#if html}
    {@html html}
  {:else}
    <pre class="plain"><code>{code}</code></pre>
  {/if}
</div>
