/// <reference lib="webworker" />
/** Shiki 在 Worker 中运行：首次编译语法（JSC 上可达上百毫秒）不阻塞主线程的流式渲染 */
import { createHighlighterCore, type HighlighterCore } from 'shiki/core';
import { createJavaScriptRegexEngine } from 'shiki/engine/javascript';

let h: Promise<HighlighterCore> | null = null;
const get = () => (h ??= createHighlighterCore({
  themes: [import('@shikijs/themes/github-light'), import('@shikijs/themes/github-dark')],
  langs: [import('@shikijs/langs/typescript'), import('@shikijs/langs/go'), import('@shikijs/langs/bash'), import('@shikijs/langs/json')],
  engine: createJavaScriptRegexEngine(),
}));

self.onmessage = async (e: MessageEvent<{ id: number; code: string; lang: string }>) => {
  const { id, code, lang } = e.data;
  try {
    const hl = await get();
    const l = hl.getLoadedLanguages().includes(lang) ? lang : 'text';
    const html = hl.codeToHtml(code, { lang: l, themes: { light: 'github-light', dark: 'github-dark' }, defaultColor: false });
    self.postMessage({ id, html });
  } catch (err) {
    self.postMessage({ id, error: String(err) });
  }
};
