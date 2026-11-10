// WebKit（接近 macOS WKWebView）上的冒烟与性能检查。用法：node scripts/webkit-check.mjs [url] [webkit 可执行文件]
import { webkit } from 'playwright-core';

const url = process.argv[2] ?? 'http://127.0.0.1:5199/';
const executablePath = process.argv[3];
const browser = await webkit.launch(executablePath ? { executablePath } : {});
const page = await browser.newPage({ viewport: { width: 1440, height: 900 } });
const errors = [];
page.on('pageerror', (e) => errors.push(String(e)));
page.on('console', (m) => { if (m.type() === 'error' && !m.text().includes('favicon')) errors.push(m.text()); });
await page.goto(url);
await page.waitForTimeout(800);

const perf = () => page.evaluate(() => Object.fromEntries([...document.querySelectorAll('.perf dt')].map((d, i) => [d.textContent, document.querySelectorAll('.perf dd')[i].textContent])));
const out = { ua: await page.evaluate(() => navigator.userAgent) };

// 1. 流式长回复
await page.click('.perf .btn:nth-child(3)');
await page.click('[data-approval] .btn.primary');
await page.waitForTimeout(9000);
out.streaming = await perf();
out.shikiBlocks = await page.locator('.codeblock .shiki').count();
out.followedBottom = await page.evaluate(() => { const s = document.querySelector('.stream'); return s.scrollHeight - s.scrollTop - s.clientHeight < 60; });

// 2. 触发符弹层与键盘选择
await page.click('textarea');
await page.keyboard.type('看 @src/co');
await page.waitForTimeout(200);
out.pickerVisible = await page.locator('.picker').isVisible();
out.focusStaysInTextarea = await page.evaluate(() => document.activeElement?.tagName === 'TEXTAREA');
await page.keyboard.press('ArrowDown');
await page.keyboard.press('Enter');
out.fileChip = await page.locator('.composer .chip').last().textContent();

// 3. Enter 发送（空闲时直接投递）
await page.keyboard.type('的暂停逻辑');
await page.keyboard.press('Enter');
await page.waitForTimeout(300);
out.sentBubble = await page.locator('.bubble').last().textContent();

// 4. 引用菜单与模型选择（上一步发送已带走所有 chip，先加一个 # 引用）
await page.keyboard.type('#修复');
await page.waitForTimeout(150);
await page.keyboard.press('Enter');
await page.click('.chip-trigger');
await page.waitForTimeout(150);
out.refMenuVisible = await page.locator('.menu').isVisible();
await page.keyboard.press('Escape');
await page.locator('.settings-bar .sel').first().click();
await page.waitForTimeout(150);
await page.keyboard.press('ArrowDown');
await page.keyboard.press('Enter');
await page.waitForTimeout(1000);
out.model = (await page.locator('.settings-bar .sel').first().textContent())?.trim();

// 5. 向前加载历史的滚动锚定
out.prepend = await page.evaluate(async () => {
  const wait = (ms) => new Promise((r) => setTimeout(r, ms));
  const sc = document.querySelector('.stream');
  const top = () => sc.getBoundingClientRect().top;
  const vis = () => { const r = [...sc.querySelectorAll('.row')].find((x) => x.getBoundingClientRect().bottom > top() + 1); return r ? { id: r.dataset.id, y: r.getBoundingClientRect().top - top() } : null; };
  const res = [];
  for (let i = 0; i < 3; i++) { sc.scrollTop = 150; await wait(40); const a = vis(); await wait(700); const b = vis(); res.push(a && b && a.id === b.id ? Math.round(b.y - a.y) : `changed ${a?.id}→${b?.id}`); }
  return res;
});

// 6. 载入全部 10,000 条后滚动到中部
await page.click('.perf .btn:nth-child(1)');
await page.waitForTimeout(500);
await page.evaluate(() => { const s = document.querySelector('.stream'); s.scrollTop = s.scrollHeight / 2; });
await page.waitForTimeout(800);
out.allHistory = await perf();
out.slowPublishes = await page.evaluate(() => window.__slow ?? []);
out.errors = errors;
console.log(JSON.stringify(out, null, 2));
await browser.close();
