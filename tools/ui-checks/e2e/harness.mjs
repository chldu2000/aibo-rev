/** e2e 夹具：启动静态服务（如目标需要）、浏览器，并按夹具打开页面。 */
import { createServer } from 'node:http';
import { readFile } from 'node:fs/promises';
import { createRequire } from 'node:module';
import { extname, join, normalize } from 'node:path';
import { webkit, chromium } from 'playwright-core';
import { abs, config } from '../lib/common.mjs';

const args = JSON.parse(process.env.UI_CHECKS_ARGS ?? '[]');
const opt = (name, dflt) => { const i = args.indexOf(`--${name}`); return i >= 0 ? args[i + 1] : dflt; };
export const targetName = opt('target', config.defaultTarget);
export const target = config.targets[targetName];
if (!target) throw new Error(`ui-checks.config.json 中没有目标 ${targetName}`);
const browserName = opt('browser', 'webkit');

const MIME = { '.html': 'text/html; charset=utf-8', '.css': 'text/css; charset=utf-8', '.js': 'text/javascript; charset=utf-8', '.mjs': 'text/javascript; charset=utf-8', '.json': 'application/json', '.svg': 'image/svg+xml', '.png': 'image/png' };
let server, baseUrl, browser;

export async function start() {
  if (target.serve) {
    const root = abs(target.serve);
    server = createServer(async (req, res) => {
      const path = normalize(decodeURIComponent(new URL(req.url, 'http://x').pathname)).replace(/^(\.\.[/\\])+/, '');
      try {
        const body = await readFile(join(root, path));
        res.writeHead(200, { 'content-type': MIME[extname(path)] ?? 'application/octet-stream', 'cache-control': 'no-store' });
        res.end(body);
      } catch { res.writeHead(404); res.end(); }
    });
    await new Promise((r) => server.listen(0, '127.0.0.1', r));
    baseUrl = `http://127.0.0.1:${server.address().port}${target.path}`;
  } else baseUrl = target.url;
  try {
    browser = await (browserName === 'chromium' ? chromium : webkit).launch();
  } catch (e) {
    if (/Executable doesn't exist/.test(String(e))) {
      const pkg = browserName === 'chromium' ? 'chromium-headless-shell' : 'webkit';
      throw new Error(`未安装与 playwright-core 匹配的 ${browserName}。运行：cd tools/ui-checks && npx playwright-core install ${pkg}`);
    }
    throw e;
  }
  return { browserName, targetName, baseUrl };
}
export async function stop() { await browser?.close(); server?.close(); }

/** 打开夹具页面；目标不支持该夹具时跳过当前测试并返回 null */
export async function open(t, fixture, { viewport = { width: 1440, height: 900 } } = {}) {
  if (!target.fixtures.includes(fixture)) { t.skip(`目标 ${targetName} 未提供夹具 ${fixture}`); return null; }
  const ctx = await browser.newContext({ viewport });
  const page = await ctx.newPage();
  page.errors = [];
  page.on('pageerror', (e) => page.errors.push(String(e)));
  // 测试期间出现的页面脚本错误视为失败
  t.after(async () => { const errs = [...page.errors]; await ctx.close(); if (errs.length) throw new Error(`页面脚本错误：${errs.join('；')}`); });
  await page.goto(`${baseUrl}?fixture=${fixture}`);
  await page.waitForSelector(tid('composer-input'), { state: 'attached', timeout: 5000 });
  return page;
}
export async function openRaw(t, query) {
  const ctx = await browser.newContext();
  const page = await ctx.newPage();
  t.after(() => ctx.close());
  await page.goto(`${baseUrl}${query}`);
  return page;
}

export const tid = (id, extra = '') => `[data-testid="${id}"]${extra}`;

const require = createRequire(import.meta.url);
let axeSource;
export async function axe(page) {
  axeSource ??= await readFile(require.resolve('axe-core/axe.min.js'), 'utf8');
  await page.addScriptTag({ content: axeSource });
  return page.evaluate(async () => {
    const r = await window.axe.run({ exclude: [['[data-ui-check="ignore"]']] }, { resultTypes: ['violations'] });
    return r.violations.filter((v) => v.impact === 'serious' || v.impact === 'critical')
      .map((v) => ({ id: v.id, impact: v.impact, help: v.help, nodes: v.nodes.slice(0, 3).map((n) => n.target.join(' ')) }));
  });
}

/** 运行时样式检查：左侧信号条与小于 12px 的可见文字（忽略 data-ui-check="ignore" 区域与图形徽记） */
export function runtimeStyleIssues(page) {
  return page.evaluate(() => {
    const desc = (el) => el.dataset.testid ? `[data-testid=${el.dataset.testid}]` : `${el.tagName.toLowerCase()}${el.id ? '#' + el.id : ''}${el.className && typeof el.className === 'string' ? '.' + el.className.trim().split(/\s+/).join('.') : ''}`;
    const transparent = (c) => c === 'transparent' || /rgba\(.+,\s*0\)$/.test(c);
    const out = { leftAccent: [], smallText: [] };
    for (const el of document.querySelectorAll('body *')) {
      if (el.closest('[data-ui-check="ignore"]')) continue;
      const cs = getComputedStyle(el);
      if (cs.display === 'none' || cs.visibility === 'hidden' || !el.getClientRects().length) continue;
      const bl = parseFloat(cs.borderLeftWidth), br = parseFloat(cs.borderRightWidth);
      if (bl > 1 && bl > br && cs.borderLeftStyle !== 'none' && !transparent(cs.borderLeftColor)) out.leftAccent.push(`${desc(el)} border-left ${cs.borderLeftWidth}`);
      for (const ps of ['::before', '::after']) {
        const p = getComputedStyle(el, ps);
        if (p.content === 'none' || p.content === 'normal') continue;
        const w = parseFloat(p.width), h = parseFloat(p.height);
        if (p.position === 'absolute' && parseFloat(p.left) === 0 && w > 0 && w <= 6 && h >= 10 && !transparent(p.backgroundColor)) out.leftAccent.push(`${desc(el)}${ps} ${p.width}×${p.height}`);
      }
      const hasText = [...el.childNodes].some((n) => n.nodeType === 3 && n.textContent.trim());
      if (hasText && parseFloat(cs.fontSize) < 12 && !el.closest('[role="img"], [aria-hidden="true"], .sr-only')) out.smallText.push(`${desc(el)} ${cs.fontSize}`);
    }
    return out;
  });
}
