/**
 * 流式 Markdown：按空行与代码围栏切块，已渲染块按文本缓存。
 * 流式追加时只有最后一块的文本变化，前面的块命中缓存、DOM 不更新。
 * 原型限制：用空行切块会把“松散列表”拆成多段，正式实现需改为按 token 切块。
 */
import MarkdownIt from 'markdown-it';

export type Block =
  | { kind: 'md'; html: string }
  | { kind: 'code'; lang: string; code: string; closed: boolean };

const md = new MarkdownIt({ html: false, linkify: true });
const cache = new Map<string, string>();
let renders = 0;
export const markdownStats = () => ({ renders, cached: cache.size });

function renderMd(text: string): string {
  let html = cache.get(text);
  if (html === undefined) {
    html = md.render(text);
    renders++;
    if (cache.size > 5000) cache.clear();
    cache.set(text, html);
  }
  return html;
}

export function toBlocks(src: string): Block[] {
  const blocks: Block[] = [];
  let buf: string[] = [];
  let code: string[] | null = null;
  let lang = '';
  const flush = () => {
    const t = buf.join('\n').trim();
    if (t) blocks.push({ kind: 'md', html: renderMd(t) });
    buf = [];
  };
  for (const line of src.split('\n')) {
    if (code) {
      if (/^```\s*$/.test(line)) { blocks.push({ kind: 'code', lang, code: code.join('\n'), closed: true }); code = null; }
      else code.push(line);
      continue;
    }
    const fence = line.match(/^```([\w-]*)\s*$/);
    if (fence) { flush(); code = []; lang = fence[1] || 'text'; }
    else if (line.trim() === '') flush();
    else buf.push(line);
  }
  if (code) blocks.push({ kind: 'code', lang, code: code.join('\n'), closed: false });
  flush();
  return blocks;
}
