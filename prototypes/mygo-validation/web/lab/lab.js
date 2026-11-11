// 一次性实验：Channel、iframe 面板、粘贴拖放与 IME 记录。
const logEl = document.getElementById('log');
const show = (event, data) => {
  logEl.textContent = `${new Date().toISOString().slice(11, 23)} ${event} ${JSON.stringify(data)}\n` + logEl.textContent;
  try { mygo.call('Lab.Log', event, data); } catch (e) {}
};
const wait = (ms) => new Promise((r) => setTimeout(r, ms));

// ---- iframe 面板 ----
const frames = {};
const readyInfo = {};
const waiters = [];
window.addEventListener('message', async (e) => {
  const m = e.data || {};
  const key = Object.keys(frames).find((k) => frames[k].contentWindow === e.source);
  if (m.type === 'relay') {
    // 父页面代替面板调用 Go；产品中由宿主 SDK 完成并附带面板的受限凭据
    const value = await mygo.call('Lab.Echo', m.value);
    e.source.postMessage({ type: 'relay-reply', id: m.id, value }, '*');
    return;
  }
  if (m.type === 'ready') readyInfo[key] = { origin: m.origin, hasMygo: m.hasMygo, callResult: m.callResult, eventOrigin: e.origin };
  waiters.forEach((w) => w(key, m));
});
const waitFor = (key, type, ms = 10000) => new Promise((resolve) => {
  const t = setTimeout(() => resolve(null), ms);
  const w = (k, m) => { if (k === key && m.type === type) { clearTimeout(t); waiters.splice(waiters.indexOf(w), 1); resolve(m); } };
  waiters.push(w);
});
function addFrame(key, src) {
  const f = document.createElement('iframe');
  f.title = key;
  f.src = src;
  document.getElementById('frames').append(f);
  frames[key] = f;
  return waitFor(key, 'ready');
}

// 父页面在 dur 内的最大定时器间隔，用来判断 iframe 是否与父页面同线程
async function maxGap(dur) {
  let last = performance.now(), max = 0;
  const id = setInterval(() => { const n = performance.now(); max = Math.max(max, n - last); last = n; }, 20);
  await wait(dur);
  clearInterval(id);
  return Math.round(max);
}

// ---- Channel ----
async function stream(n, size, intervalMs, opts = {}) {
  let count = 0, latSum = 0, latMax = 0, outOfOrder = 0, prev = -1;
  const t0 = performance.now();
  const ch = mygo.channel((m) => {
    count++;
    const lat = Date.now() - m.sentAt / 1e6;
    latSum += lat; latMax = Math.max(latMax, lat);
    if (m.seq !== prev + 1) outOfOrder++;
    prev = m.seq;
    if (opts.busyMs) { const t = performance.now(); while (performance.now() - t < opts.busyMs) {} }
  });
  const call = mygo.call('Lab.Stream', n, size, intervalMs, ch);
  if (opts.closeAfterMs) { await wait(opts.closeAfterMs); ch.close(); }
  let host;
  try { host = await call; } catch (err) { host = { error: String(err && err.message) }; }
  return { received: count, pageMs: Math.round(performance.now() - t0), latAvgMs: +(latSum / Math.max(count, 1)).toFixed(2), latMaxMs: Math.round(latMax), outOfOrder, host };
}

async function echoRoundTrip(n) {
  const times = [];
  for (let i = 0; i < n; i++) { const s = performance.now(); await mygo.call('Lab.Echo', i); times.push(performance.now() - s); }
  times.sort((a, b) => a - b);
  return { avg: +(times.reduce((a, b) => a + b, 0) / n).toFixed(3), p95: +times[Math.floor(n * 0.95)].toFixed(3), max: +times[n - 1].toFixed(3) };
}

window.runLab = async function () {
  const r = { origin: location.origin };
  r.echo = await echoRoundTrip(300);
  r.streamFast = await stream(20000, 200, 0);
  r.streamTokens = await stream(300, 40, 16);
  r.streamSlowConsumer = await stream(2000, 4000, 0, { busyMs: 2 });
  r.streamCancel = await stream(1000000, 100, 1, { closeAfterMs: 400 });

  // 面板：同源 iframe 与两个包各自的 origin
  r.frames = {};
  for (const [key, src] of [['same-origin', '/lab/panel.html'], ['pkg-a', 'aibopkg://pkg-a/panel.html'], ['pkg-b', 'aibopkg://pkg-b/panel.html']]) {
    const m = await addFrame(key, src);
    r.frames[key] = m ? readyInfo[key] : 'no ready message';
  }
  frames['pkg-a'].contentWindow.postMessage({ type: 'relay', n: 200 }, '*');
  r.relay = await waitFor('pkg-a', 'relay-done', 20000);

  // 面板忙等 3 秒时父页面是否仍能运行
  r.hang = {};
  for (const key of ['same-origin', 'pkg-a']) {
    frames[key].contentWindow.postMessage({ type: 'hang', ms: 3000 }, '*');
    r.hang[key] = { parentMaxGapMs: await maxGap(3500) };
    await waitFor(key, 'hang-done', 5000);
  }
  show('lab-result', r);

  // 最后：pkg-b 面板死循环，再移除该 iframe，父页面能否继续
  frames['pkg-b'].contentWindow.postMessage({ type: 'loop' }, '*');
  const during = await maxGap(1500);
  frames['pkg-b'].remove();
  const after = await maxGap(1000);
  r.infiniteLoop = { parentMaxGapDuringMs: during, parentMaxGapAfterRemoveMs: after };
  show('lab-loop-result', r.infiniteLoop);
  return r;
};

document.getElementById('run').onclick = async () => show('manual-run', await window.runLab());

// ---- 手动：IME ----
const ime = document.getElementById('ime');
for (const t of ['compositionstart', 'compositionupdate', 'compositionend']) ime.addEventListener(t, (e) => show(t, { data: e.data }));
ime.addEventListener('keydown', (e) => { if (e.key === 'Enter') show('enter', { isComposing: e.isComposing, keyCode: e.keyCode, value: ime.value }); });

// ---- 手动：粘贴与拖放 ----
const describe = (files) => [...files].map((f) => ({ name: f.name, type: f.type, size: f.size }));
document.getElementById('paste').addEventListener('paste', (e) => {
  const items = [...e.clipboardData.items].map((i) => ({ kind: i.kind, type: i.type }));
  show('paste', { items, files: describe(e.clipboardData.files) });
  const img = [...e.clipboardData.files].find((f) => f.type.startsWith('image/'));
  if (img) { const el = new Image(); el.src = URL.createObjectURL(img); el.style.maxHeight = '80px'; e.currentTarget.replaceChildren(el); }
});
const drop = document.getElementById('drop');
drop.addEventListener('dragover', (e) => e.preventDefault());
drop.addEventListener('drop', (e) => { e.preventDefault(); show('drop', { types: [...e.dataTransfer.types], files: describe(e.dataTransfer.files) }); });
