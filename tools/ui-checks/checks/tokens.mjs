import { existsSync, readFileSync } from 'node:fs';
import { abs, config, read, walk, Report } from '../lib/common.mjs';
import { extractStyles } from '../lib/css.mjs';
import { renderTokensCss } from '../build-tokens.mjs';

const tokens = () => JSON.parse(read(config.tokens));

function parseColor(v) {
  const s = v.trim().toLowerCase();
  let m = s.match(/^#([0-9a-f]{3}|[0-9a-f]{6})$/);
  if (m) { const h = m[1].length === 3 ? m[1].split('').map((c) => c + c).join('') : m[1]; return [0, 2, 4].map((i) => parseInt(h.slice(i, i + 2), 16)); }
  m = s.match(/^rgba?\(([^)]+)\)$/);
  if (m) { const p = m[1].split(',').map((x) => parseFloat(x)); if (p.length === 3 || p[3] === 1) return p.slice(0, 3); }
  return null; // 带透明度或非颜色值不参与对比度计算
}
const lum = (rgb) => { const c = rgb.map((v) => v / 255).map((v) => (v <= 0.03928 ? v / 12.92 : ((v + 0.055) / 1.055) ** 2.4)); return 0.2126 * c[0] + 0.7152 * c[1] + 0.0722 * c[2]; };
export const ratio = (a, b) => { const x = lum(a), y = lum(b); return (Math.max(x, y) + 0.05) / (Math.min(x, y) + 0.05); };

export function tokensComplete() {
  const r = new Report('tokens-complete');
  const { light, dark } = tokens().color;
  for (const k of Object.keys(light)) if (!(k in dark)) r.error(`深色缺少令牌 ${k}`);
  for (const k of Object.keys(dark)) if (!(k in light)) r.error(`浅色缺少令牌 ${k}`);
  r.note(`${Object.keys(light).length} 个颜色令牌`);
  return r;
}

export function contrast() {
  const r = new Report('contrast');
  const t = tokens();
  let n = 0;
  for (const theme of ['light', 'dark']) {
    for (const [fg, bg, min] of t.contrast.pairs) {
      const a = parseColor(t.color[theme][fg] ?? ''), b = parseColor(t.color[theme][bg] ?? '');
      if (!a || !b) { r.error(`${theme}: ${fg}/${bg} 不是不透明颜色，无法计算`); continue; }
      const v = ratio(a, b); n++;
      if (v < min) r.error(`${theme}: ${fg} 在 ${bg} 上为 ${v.toFixed(2)}:1，要求 ≥ ${min}:1`);
    }
  }
  r.note(`${n} 组对比度`);
  return r;
}

const norm = (v) => v.replace(/\s+/g, '').toLowerCase();

export function tokenSync() {
  const r = new Report('token-sync');
  const t = tokens();
  // 1. 生成文件是否最新
  const expected = renderTokensCss(t);
  if (!existsSync(abs(config.tokensCss))) r.error(`${config.tokensCss} 不存在，运行 node tools/ui-checks/build-tokens.mjs`);
  else if (readFileSync(abs(config.tokensCss), 'utf8') !== expected) r.error(`${config.tokensCss} 与 tokens.json 不一致，运行 node tools/ui-checks/build-tokens.mjs`);
  // 2. 样式中自行声明的颜色令牌必须与 tokens.json 相同
  let checked = 0;
  for (const file of walk(config.styleRoots, ['.css', '.svelte', '.html'])) {
    if (file === config.tokensCss) continue;
    for (const d of extractStyles(file, read(file)).decls) {
      if (!d.prop.startsWith('--')) continue;
      const name = d.prop.slice(2);
      if (!(name in t.color.light)) continue;
      const theme = d.selectors.some((s) => /dark/.test(s)) ? 'dark' : 'light';
      checked++;
      if (norm(d.value) !== norm(t.color[theme][name])) r.error(`--${name}（${theme}）为 ${d.value}，令牌为 ${t.color[theme][name]}`, `${file}:${d.line}`);
    }
  }
  r.note(`核对了 ${checked} 处令牌声明`);
  return r;
}
