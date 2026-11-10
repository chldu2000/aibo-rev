import { config, read, walk, allowed, Report } from '../lib/common.mjs';
import { extractStyles } from '../lib/css.mjs';

const EXTS = ['.css', '.svelte', '.html', '.js', '.ts'];
const files = () => walk(config.styleRoots, EXTS).filter((f) => f !== config.tokensCss);
const px = (v) => { const m = v.match(/(-?\d*\.?\d+)px/); return m ? parseFloat(m[1]) : null; };
const pseudo = (sels) => sels.some((s) => /::?(before|after)\b/.test(s));

/** 对每个文件的样式声明执行 fn(decl, file, lines, report) */
function eachDecl(name, fn) {
  const r = new Report(name);
  let n = 0;
  for (const file of files()) {
    const text = read(file);
    const lines = text.split('\n');
    const { decls } = extractStyles(file, text);
    n += decls.length;
    fn(decls, file, lines, r);
  }
  r.note(`扫描 ${files().length} 个文件、${n} 条声明`);
  return r;
}

export const noLeftAccent = () => eachDecl('no-left-accent', (decls, file, lines, r) => {
  for (const d of decls) {
    if (allowed(lines, d.line, 'left-accent')) continue;
    const where = `${file}:${d.line}`;
    if (/^border-(left|inline-start)(-width)?$/.test(d.prop)) {
      const w = px(d.value);
      if (w !== null && w > 1) r.error(`${d.prop}: ${d.value} — 左侧信号条（UI-VIS-05）`, where);
    }
    if (d.prop === 'box-shadow') {
      const m = d.value.match(/inset\s+(-?\d*\.?\d+)px\s+0/);
      if (m && Math.abs(parseFloat(m[1])) > 1) r.error(`box-shadow: ${d.value} — inset 竖条（UI-VIS-05）`, where);
    }
  }
  // 绝对定位、贴左、窄而有背景的伪元素竖条
  const groups = new Map();
  for (const d of decls) if (pseudo(d.selectors)) { const k = d.selectors.join(' > '); (groups.get(k) ?? groups.set(k, []).get(k)).push(d); }
  for (const [sel, ds] of groups) {
    const get = (p) => ds.find((d) => d.prop === p)?.value ?? '';
    const w = px(get('width'));
    const bg = get('background') || get('background-color');
    if (get('position') === 'absolute' && /^0(px)?$/.test(get('left')) && w !== null && w <= 6 && bg && !/transparent|none/.test(bg) && (get('height') || get('top'))) {
      const line = ds[0].line;
      if (!allowed(lines, line, 'left-accent')) r.error(`${sel} — 贴左的窄竖条伪元素（UI-VIS-05）`, `${file}:${line}`);
    }
  }
});

const COLOR = /#[0-9a-fA-F]{3,8}\b|\b(rgba?|hsla?|oklch|lab|lch)\(/;
export const noHardcodedColor = () => eachDecl('no-hardcoded-color', (decls, file, lines, r) => {
  for (const d of decls) {
    if (d.prop.startsWith('--')) continue; // 令牌声明由 token-sync 核对
    if (!COLOR.test(d.value)) continue;
    if (allowed(lines, d.line, 'color')) continue;
    r.error(`${d.prop}: ${d.value} — 硬写颜色，应使用令牌（UI-VIS-01）`, `${file}:${d.line}`);
  }
});

export const minFontSize = () => eachDecl('min-font-size', (decls, file, lines, r) => {
  for (const d of decls) {
    if (d.prop !== 'font-size' && d.prop !== 'font') continue;
    const v = d.prop === 'font' ? d.value.match(/(\d*\.?\d+)px/)?.[0] ?? '' : d.value;
    const n = px(v);
    if (n === null || n >= 12) continue;
    if (allowed(lines, d.line, 'font-size')) continue;
    r.error(`${d.prop}: ${d.value} — 小于 12px（UI-VIS-04）`, `${file}:${d.line}`);
  }
});

export const focusOutline = () => eachDecl('focus-outline', (decls, file, lines, r) => {
  for (const d of decls) {
    if (d.prop !== 'outline' || !/^(none|0)(\s|$)/.test(d.value)) continue;
    if (allowed(lines, d.line, 'outline')) continue;
    r.error(`outline: ${d.value} — 去掉焦点环但未说明替代样式（UI-VIS-06）`, `${file}:${d.line}`);
  }
});

export function reducedMotion() {
  const r = new Report('reduced-motion');
  for (const file of files()) {
    const s = extractStyles(file, read(file));
    if (s.keyframes.length && !s.hasReducedMotion) r.error(`定义了 ${s.keyframes.join(', ')}，但没有 prefers-reduced-motion 规则（UI-VIS-08）`, file);
  }
  return r;
}
