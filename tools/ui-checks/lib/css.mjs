/**
 * 足够用于规则检查的 CSS 扫描器：提取每条声明的属性、值、行号及所在选择器链。
 * 来源：.css 文件全文、.svelte/.html 的 <style> 块、模板中的 style="..." 属性。
 */
import { lineIndex } from './common.mjs';

export function scanCss(text, offset = 0, lineOf) {
  const clean = text.replace(/\/\*[\s\S]*?\*\//g, (m) => m.replace(/[^\n]/g, ' '));
  const decls = [];
  const stack = [];
  const keyframes = [];
  let seg = 0;
  const push = (from, to) => {
    const raw = clean.slice(from, to);
    const i = raw.indexOf(':');
    if (i < 0 || !raw.trim()) return;
    const prop = raw.slice(0, i).trim().toLowerCase();
    const value = raw.slice(i + 1).trim();
    if (!prop || /[\s{}]/.test(prop)) return;
    const lead = raw.length - raw.trimStart().length;
    decls.push({ prop, value, line: lineOf(offset + from + lead), selectors: stack.map((s) => s.sel) });
  };
  for (let i = 0; i < clean.length; i++) {
    const ch = clean[i];
    if (ch === '{') {
      const sel = clean.slice(seg, i).trim();
      stack.push({ sel, line: lineOf(offset + seg) });
      if (/^@keyframes\b/.test(sel)) keyframes.push(sel);
      seg = i + 1;
    } else if (ch === ';') {
      if (stack.length) push(seg, i);
      seg = i + 1;
    } else if (ch === '}') {
      if (stack.length) push(seg, i);
      stack.pop();
      seg = i + 1;
    }
  }
  return { decls, keyframes, hasReducedMotion: /prefers-reduced-motion/.test(clean) };
}

/** 从任意源文件提取样式声明 */
export function extractStyles(file, text) {
  const lineOf = lineIndex(text);
  const results = { decls: [], keyframes: [], hasReducedMotion: false };
  const add = (r) => { results.decls.push(...r.decls); results.keyframes.push(...r.keyframes); results.hasReducedMotion ||= r.hasReducedMotion; };
  if (file.endsWith('.css')) add(scanCss(text, 0, lineOf));
  else {
    for (const m of text.matchAll(/<style[^>]*>([\s\S]*?)<\/style>/g)) add(scanCss(m[1], m.index + m[0].indexOf(m[1]), lineOf));
    // style="a: b; c: d"（含模板插值时只检查字面部分）
    for (const m of text.matchAll(/style=(["'])([^"'`]*?)\1/g)) {
      const body = m[2];
      const r = scanCss(`x{${body}}`, m.index + m[0].indexOf(body) - 2, lineOf);
      r.decls.forEach((d) => { d.selectors = ['[inline style]']; });
      add(r);
    }
  }
  return results;
}
