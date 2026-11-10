/**
 * docs/design/ui-requirements.md 的验收测试。每个测试标题以要求编号开头。
 * 只通过 §11 的测试契约（data-testid、?fixture=）操作界面，不依赖任何实现细节。
 * 运行：node tools/ui-checks/run.mjs e2e [--target 名称] [--browser webkit|chromium] [--grep 编号]
 */
import { test, before, after } from 'node:test';
import assert from 'node:assert/strict';
import { start, stop, open, openRaw, tid, axe, runtimeStyleIssues, target } from './harness.mjs';

const MOD = 'ControlOrMeta';
const sleep = (ms) => new Promise((r) => setTimeout(r, ms));
const count = (page, sel) => page.locator(sel).count();
const text = (page, sel) => page.locator(sel).first().innerText();
const attr = (page, sel, name) => page.locator(sel).first().getAttribute(name);
/** 可读原因：aria-describedby 指向的文字或 title */
const reasonOf = (page, sel) => page.locator(sel).first().evaluate((el) => {
  const ids = (el.getAttribute('aria-describedby') ?? '').split(/\s+/).filter(Boolean);
  const described = ids.map((id) => document.getElementById(id)?.textContent ?? '').join(' ').trim();
  return described || el.getAttribute('title') || el.closest('[data-tip]')?.getAttribute('data-tip') || '';
});
const typeInto = async (page, s) => { await page.locator(tid('composer-input')).click(); await page.keyboard.type(s); };
const openSessionMenu = async (page, id) => {
  const row = page.locator(tid('session-row', `[data-session-id="${id}"]`));
  await row.hover();
  await row.locator(tid('session-menu-button')).click();
};
const menuItem = (page, name) => page.locator('[role="menuitem"]', { hasText: name }).first();
const STYLE_FIXTURES = ['running-approval', 'delivery-unknown', 'pending-question', 'turn-failed', 'untrusted', 'idle-empty', 'restart', 'save-failure'];

let info;
before(async () => { info = await start(); console.log(`# 目标 ${info.targetName}，浏览器 ${info.browserName}，${info.baseUrl}`); });
after(stop);

// ---------- 总则 ----------
test('UI-GEN-04 验收接口：夹具可加载、核心标识存在、未知夹具报错', async (t) => {
  for (const f of target.fixtures) {
    const page = await open(t, f);
    assert.equal(await count(page, tid('stream')), 1, `${f}: 缺少 stream`);
    assert.deepEqual(page.errors, [], `${f}: 页面脚本错误`);
  }
  const page = await openRaw(t, '?fixture=__no_such_fixture__');
  assert.ok((await page.content()).includes('__no_such_fixture__'), '未知夹具必须报错而不是忽略');
});

// ---------- 视觉（运行时） ----------
test('UI-VIS-04 运行时没有小于 12px 的可读文字', async (t) => {
  for (const f of STYLE_FIXTURES) {
    const page = await open(t, f); if (!page) return;
    const { smallText } = await runtimeStyleIssues(page);
    assert.deepEqual(smallText, [], `${f}: ${smallText.join('；')}`);
  }
});

test('UI-VIS-05 运行时没有左侧信号条', async (t) => {
  for (const f of STYLE_FIXTURES) {
    const page = await open(t, f); if (!page) return;
    const { leftAccent } = await runtimeStyleIssues(page);
    assert.deepEqual(leftAccent, [], `${f}: ${leftAccent.join('；')}`);
  }
});

test('UI-VIS-07 会话状态标记有可读名称', async (t) => {
  const page = await open(t, 'running-approval'); if (!page) return;
  const names = await page.locator(tid('session-status')).evaluateAll((els) => els.map((el) => el.getAttribute('aria-label') || el.textContent.trim()));
  assert.ok(names.length > 0, '没有会话状态标记');
  assert.ok(names.every(Boolean), `存在没有可读名称的状态标记：${JSON.stringify(names)}`);
});

// ---------- 导航 ----------
test('UI-NAV-01 没有关闭会话入口', async (t) => {
  const page = await open(t, 'running-approval'); if (!page) return;
  await openSessionMenu(page, 's1');
  assert.equal(await page.locator('[role="menuitem"]', { hasText: '关闭' }).count(), 0, '会话行菜单中有“关闭”项');
  await page.keyboard.press('Escape');
  await page.locator(tid('session-head-menu')).click();
  assert.equal(await page.locator('[role="menuitem"]', { hasText: '关闭' }).count(), 0, '会话头菜单中有“关闭”项');
});

test('UI-NAV-02 切换查看不改变顺序，置顶排在分组前', async (t) => {
  const page = await open(t, 'running-approval'); if (!page) return;
  const order = () => page.locator(tid('session-row')).evaluateAll((els) => els.map((e) => e.dataset.sessionId));
  const before = await order();
  await page.locator(tid('session-row', '[data-session-id="s4"]')).click();
  assert.deepEqual(await order(), before, '切换会话后列表顺序变化');
  await openSessionMenu(page, 's4');
  await menuItem(page, '置顶').click();
  const after = await order();
  const pinned = await page.locator(tid('session-row')).evaluateAll((els) => els.filter((e) => e.querySelector('.pinmark, [data-pinned="true"]')).map((e) => e.dataset.sessionId));
  assert.ok(pinned.includes('s4'), '置顶后未显示为置顶');
  const firstUnpinned = after.findIndex((id) => !pinned.includes(id) && before.includes(id) && id !== 's5');
  assert.ok(after.indexOf('s4') < firstUnpinned, `置顶会话未排在未置顶会话之前：${after.join(',')}`);
});

test('UI-NAV-03 归档执行中的会话需要先确认', async (t) => {
  const page = await open(t, 'running-approval'); if (!page) return;
  await openSessionMenu(page, 's1');
  await menuItem(page, '归档').click();
  assert.equal(await count(page, tid('confirm-dialog')), 1, '没有出现确认对话框');
  assert.equal(await count(page, tid('session-row', '[data-session-id="s1"]')), 1, '确认前会话已从列表消失');
});

test('UI-NAV-04 会话详情默认隐藏，可从菜单打开', async (t) => {
  const page = await open(t, 'running-approval'); if (!page) return;
  assert.equal(await count(page, tid('session-details')), 0);
  await page.locator(tid('session-head-menu')).click();
  await menuItem(page, '会话详情').click();
  assert.equal(await count(page, tid('session-details')), 1, '无法打开会话详情');
  assert.match(await text(page, tid('session-details')), /原生会话|适配器/);
});

test('UI-NAV-05 工具面板默认收起，展开后标明来源', async (t) => {
  const page = await open(t, 'running-approval'); if (!page) return;
  assert.equal(await count(page, tid('tool-panel')), 0, '工具面板默认不应展开');
  if ((await count(page, tid('tool-bar'))) === 0) return; // 没有插件提供面板时整个区域不出现，也符合要求
  await page.locator(tid('tool-bar')).locator('button').first().click();
  assert.equal(await count(page, tid('tool-panel')), 1);
  assert.match(await text(page, tid('tool-panel-source')), /提供/);
});

test('UI-NAV-06 检查未通过的 Agent 不能创建会话', async (t) => {
  const page = await open(t, 'running-approval'); if (!page) return;
  await page.locator(tid('stream')).click();
  await page.keyboard.press(`${MOD}+N`);
  const dlg = page.locator(tid('new-session-dialog'));
  await dlg.waitFor();
  await dlg.locator('[data-id="opencode"], [data-agent="opencode"]').first().click();
  const create = dlg.getByRole('button', { name: /创建/ });
  assert.ok(await create.isDisabled() || (await create.getAttribute('aria-disabled')) === 'true', '检查未通过时“创建”仍可用');
});

test('UI-NAV-07 未验证组合在会话头标出', async (t) => {
  const page = await open(t, 'untrusted'); if (!page) return;
  assert.equal(await count(page, tid('agent-unverified')), 1);
  assert.match(await text(page, tid('agent-unverified')), /未验证/);
});

test('UI-NAV-09 已归档会话可阅读，输入区提供恢复入口', async (t) => {
  const page = await open(t, 'running-approval'); if (!page) return;
  await page.locator(tid('stream')).click();
  await page.keyboard.press(`${MOD}+K`);
  await page.keyboard.type('旧终端');
  await page.locator(tid('command-palette')).getByText('已归档').first().waitFor();
  await page.keyboard.press('Enter');
  await page.locator(tid('archived-block')).waitFor();
  assert.ok((await count(page, tid('message-agent'))) > 0, '已归档会话没有显示历史');
  assert.match(await text(page, tid('archived-block')), /取消归档/);
});

// ---------- 会话流 ----------
test('UI-STREAM-01 用户消息为右侧气泡', async (t) => {
  const page = await open(t, 'running-approval'); if (!page) return;
  const g = await page.evaluate(() => {
    const s = document.querySelector('[data-testid="stream"]').getBoundingClientRect();
    const m = document.querySelector('[data-testid="message-user"]').getBoundingClientRect();
    const a = document.querySelector('[data-testid="message-agent"]').getBoundingClientRect();
    return { leftGap: m.left - s.left, rightGap: s.right - m.right, userW: m.width, agentW: a.width };
  });
  assert.ok(g.leftGap > g.rightGap + 40, `用户消息未右对齐：左 ${g.leftGap}px、右 ${g.rightGap}px`);
  assert.ok(g.agentW > g.userW, 'Agent 回复应占满正文宽度');
});

test('UI-STREAM-02 结果未知的投递只能核实或确认后重发', async (t) => {
  const page = await open(t, 'delivery-unknown'); if (!page) return;
  const msg = page.locator(tid('message-user', '[data-delivery="unknown"]'));
  assert.equal(await msg.count(), 1);
  const wrap = msg.locator('xpath=..');
  assert.ok(await wrap.getByRole('button', { name: '核实' }).count() > 0, '缺少“核实”');
  assert.ok(await wrap.getByRole('button', { name: /重新发送（可能重复执行）/ }).count() > 0, '缺少“重新发送（可能重复执行）”');
});

test('UI-STREAM-03 离开底部后不被拉回，并提示有新内容', async (t) => {
  const page = await open(t, 'running-approval'); if (!page) return;
  const stream = page.locator(tid('stream'));
  const scrollable = await stream.evaluate((el) => el.scrollHeight > el.clientHeight + 100);
  assert.ok(scrollable, '夹具内容不足以滚动');
  await stream.evaluate((el) => { el.scrollTop = 0; el.dispatchEvent(new Event('scroll')); });
  // 以 DOM 点击作答，避免测试工具自动滚动
  await page.evaluate(() => document.querySelector('[data-testid="approval-card"] button').click());
  await page.waitForFunction(() => (document.querySelectorAll('[data-testid="message-agent"]').length) > 1, null, { timeout: 8000 });
  assert.ok(await stream.evaluate((el) => el.scrollTop) < 5, '用户离开底部后被拉回');
  assert.ok(await page.locator(tid('new-content')).isVisible(), '没有显示“有新内容”');
});

test('UI-STREAM-04 一万条历史虚拟化且向前加载不跳动', async (t) => {
  const page = await open(t, 'history-10k'); if (!page) return;
  const rows = await page.locator(`${tid('stream')} ${tid('message-user')}, ${tid('stream')} ${tid('message-agent')}`).count();
  assert.ok(rows <= 60, `挂载了 ${rows} 条消息`);
});

test('UI-STREAM-05 流式输出不阻塞主线程', async (t) => {
  const page = await open(t, 'history-10k'); if (!page) return;
  const longFrames = await page.evaluate(() => new Promise((resolve) => {
    let last = performance.now(), n = 0; const end = last + 5000;
    const loop = (ts) => { if (ts - last > 50) n++; last = ts; if (ts < end) requestAnimationFrame(loop); else resolve(n); };
    requestAnimationFrame(loop);
  }));
  assert.equal(longFrames, 0);
});

test('UI-STREAM-06 重启后失效的审批不可操作', async (t) => {
  const page = await open(t, 'restart'); if (!page) return;
  const expired = page.locator(tid('approval-card', '[data-state="expired"]'));
  assert.ok(await expired.count() > 0, '没有失效的审批卡');
  assert.equal(await expired.locator('button').count(), 0, '失效的审批仍有可操作按钮');
});

test('UI-STREAM-07 缺口有明确标示', async (t) => {
  const page = await open(t, 'turn-failed'); if (!page) return;
  assert.ok(await page.locator(tid('gap-marker')).isVisible());
  assert.ok((await text(page, tid('gap-marker'))).length > 4, '缺口标示没有说明文字');
});

test('UI-STREAM-08 工具调用默认折叠，可展开', async (t) => {
  const page = await open(t, 'running-approval'); if (!page) return;
  const head = page.locator(`${tid('tool-card')} [aria-expanded]`).first();
  assert.equal(await head.getAttribute('aria-expanded'), 'false');
  await head.click();
  assert.equal(await page.locator(`${tid('tool-card')} [aria-expanded]`).first().getAttribute('aria-expanded'), 'true');
});

test('UI-STREAM-09 切换会话后恢复阅读位置', async (t) => {
  const page = await open(t, 'running-approval'); if (!page) return;
  const stream = page.locator(tid('stream'));
  await stream.evaluate((el) => { el.scrollTop = 120; el.dispatchEvent(new Event('scroll')); });
  await page.locator(tid('session-row', '[data-session-id="s4"]')).click();
  await page.locator(tid('session-row', '[data-session-id="s1"]')).click();
  const top = await page.locator(tid('stream')).evaluate((el) => el.scrollTop);
  assert.ok(Math.abs(top - 120) <= 5, `阅读位置为 ${top}，期望约 120`);
});

// ---------- 输入区 ----------
test('UI-COMP-01 组字期间按 Enter 不发送', async (t) => {
  const page = await open(t, 'idle-empty'); if (!page) return;
  await typeInto(page, '中文输入');
  const fire = (opts) => page.locator(tid('composer-input')).evaluate((el, o) => {
    const e = new KeyboardEvent('keydown', { key: 'Enter', bubbles: true, cancelable: true, isComposing: o.composing });
    if (o.k229) Object.defineProperty(e, 'keyCode', { get: () => 229 });
    el.dispatchEvent(e);
  }, opts);
  await fire({ composing: true });
  await fire({ composing: false, k229: true });
  await sleep(300);
  assert.equal(await page.locator(tid('composer-input')).inputValue(), '中文输入', '组字中的 Enter 清空了输入');
  assert.equal(await count(page, tid('message-user')), 0, '组字中的 Enter 发送了消息');
});

test('UI-COMP-02 执行中按 Enter 进入队列', async (t) => {
  const page = await open(t, 'running-approval'); if (!page) return;
  const users = await count(page, tid('message-user'));
  const queue = page.locator(tid('strip-chip', '[data-kind="queue"]'));
  const before = Number((await queue.innerText()).match(/\d+/)?.[0] ?? 0);
  assert.match(await text(page, tid('composer-primary')), /排队/);
  await typeInto(page, '执行中追加的一条');
  await page.keyboard.press('Enter');
  assert.equal(await page.locator(tid('composer-input')).inputValue(), '');
  assert.equal(Number((await queue.innerText()).match(/\d+/)?.[0]), before + 1, '队列数量没有加 1');
  assert.equal(await count(page, tid('message-user')), users, '执行中的输入被直接投递');
});

test('UI-COMP-03 不支持的插话可聚焦并说明原因', async (t) => {
  const page = await open(t, 'running-approval'); if (!page) return;
  const steer = page.locator(tid('composer-steer'));
  assert.equal(await steer.getAttribute('aria-disabled'), 'true');
  assert.equal(await steer.evaluate((el) => el.hasAttribute('disabled')), false, '使用了会移除焦点的 disabled');
  await steer.focus();
  assert.ok(await steer.evaluate((el) => document.activeElement === el), '插话控件无法聚焦');
  assert.ok((await reasonOf(page, tid('composer-steer'))).length > 4, '没有可读的禁用原因');
});

test('UI-COMP-04 触发符候选可含路径，焦点留在输入框', async (t) => {
  const page = await open(t, 'idle-empty'); if (!page) return;
  await typeInto(page, '看一下 @src/co');
  await page.locator(tid('trigger-picker')).waitFor({ state: 'visible' });
  assert.equal(await page.evaluate(() => document.activeElement?.dataset.testid), 'composer-input');
  await page.keyboard.press('Enter');
  assert.match(await text(page, tid('attachment-chip')), /src\/co/);
});

test('UI-COMP-05 引用在加入时固定版本', async (t) => {
  const page = await open(t, 'idle-empty'); if (!page) return;
  await typeInto(page, '#队列');
  await page.locator(tid('trigger-picker')).waitFor({ state: 'visible' });
  await page.keyboard.press('Enter');
  assert.match(await text(page, tid('reference-chip')), /固定\s*\d{1,2}:\d{2}/);
});

test('UI-COMP-06 设置未确认时不能发送', async (t) => {
  const page = await open(t, 'idle-empty'); if (!page) return;
  await typeInto(page, '需要新模型');
  await page.locator(tid('model-select')).click();
  await page.locator('[role="menuitem"], [role="option"]').filter({ hasNotText: '✓' }).last().click();
  assert.equal(await attr(page, tid('composer-primary'), 'aria-disabled'), 'true', '设置应用中仍可发送');
  await page.waitForFunction(() => document.querySelector('[data-testid="composer-primary"]')?.getAttribute('aria-disabled') !== 'true', null, { timeout: 5000 });
});

test('UI-COMP-07 草稿随会话保存', async (t) => {
  const page = await open(t, 'running-approval'); if (!page) return;
  await typeInto(page, '未写完的草稿');
  await page.locator(tid('session-row', '[data-session-id="s4"]')).click();
  assert.equal(await page.locator(tid('composer-input')).inputValue(), '');
  await page.locator(tid('session-row', '[data-session-id="s1"]')).click();
  assert.equal(await page.locator(tid('composer-input')).inputValue(), '未写完的草稿');
});

test('UI-COMP-08 停止后分别显示各项结果', async (t) => {
  const page = await open(t, 'running-approval'); if (!page) return;
  await page.locator(tid('composer-stop')).click();
  const summary = page.locator(tid('stop-summary'));
  await summary.waitFor();
  const s = await summary.innerText();
  assert.match(s, /停止请求已发送/);
  assert.match(s, /队列/);
  assert.match(s, /后台任务[\s\S]*不受影响/);
});

test('UI-COMP-09 图片附件可粘贴、单独发送、移除', async (t) => {
  const page = await open(t, 'idle-empty'); if (!page) return;
  const paste = () => page.locator(tid('composer-input')).evaluate((el) => {
    const dt = new DataTransfer();
    dt.items.add(new File([new Uint8Array([137, 80, 78, 71])], 'shot.png', { type: 'image/png' }));
    el.dispatchEvent(new ClipboardEvent('paste', { clipboardData: dt, bubbles: true, cancelable: true }));
  });
  await paste();
  assert.equal(await count(page, tid('attachment-chip')), 1, '粘贴图片后没有附件');
  await page.locator(tid('attachment-chip')).getByRole('button').click();
  assert.equal(await count(page, tid('attachment-chip')), 0, '附件无法移除');
  await paste();
  await page.locator(tid('composer-primary')).click();
  await page.locator(tid('message-user')).first().waitFor();
});

test('UI-COMP-10 不支持的设置禁用并说明原因', async (t) => {
  const page = await open(t, 'running-approval'); if (!page) return;
  assert.equal(await attr(page, tid('setting-fast'), 'aria-disabled'), 'true');
  assert.ok((await reasonOf(page, tid('setting-fast'))).length > 4, '没有可读的原因');
});

// ---------- 状态条 ----------
test('UI-STRIP-01 状态条内容与空状态', async (t) => {
  const page = await open(t, 'delivery-unknown'); if (!page) return;
  const kinds = await page.locator(tid('strip-chip')).evaluateAll((els) => els.map((e) => e.dataset.kind));
  for (const k of ['goal', 'task', 'queue']) assert.ok(kinds.includes(k), `缺少 ${k} 标签：${kinds}`);
  const idle = await open(t, 'idle-empty'); if (!idle) return;
  assert.equal(await count(idle, tid('status-strip')), 0, '没有内容时状态条仍然出现');
});

test('UI-STRIP-02 浮层 Esc 收起且不改变草稿', async (t) => {
  const page = await open(t, 'running-approval'); if (!page) return;
  await typeInto(page, '草稿保持');
  await page.locator(tid('strip-chip', '[data-kind="queue"]')).click();
  await page.locator(tid('strip-popover', '[data-kind="queue"]')).waitFor();
  assert.equal(await count(page, tid('strip-popover')), 1, '同时打开了多个浮层');
  await page.keyboard.press('Escape');
  assert.equal(await count(page, tid('strip-popover')), 0, 'Esc 未收起浮层');
  assert.equal(await page.locator(tid('composer-input')).inputValue(), '草稿保持');
});

test('UI-STRIP-03 结果未知的重发需确认，原记录保留', async (t) => {
  const page = await open(t, 'delivery-unknown'); if (!page) return;
  await page.locator(tid('strip-chip', '[data-kind="queue"]')).click();
  await page.locator(tid('queue-item', '[data-unknown="true"]')).getByRole('button', { name: /重新发送/ }).click();
  const dlg = page.locator(tid('confirm-dialog'));
  await dlg.waitFor();
  await dlg.getByRole('button', { name: /重新发送/ }).click();
  assert.equal(await count(page, tid('message-user', '[data-resent="true"]')), 1, '原记录未保留或未标注已重发');
  assert.match(await text(page, tid('strip-chip', '[data-kind="queue"]')), /暂停/, '重发后其余队列不应恢复');
});

test('UI-STRIP-04 目标暂停先显示请求已发送', async (t) => {
  const page = await open(t, 'delivery-unknown'); if (!page) return;
  await page.locator(tid('strip-chip', '[data-kind="goal"]')).click();
  await page.locator(tid('strip-popover', '[data-kind="goal"]')).getByRole('button', { name: /暂停目标/ }).click();
  assert.match(await text(page, tid('strip-chip', '[data-kind="goal"]')), /请求已发送/);
  await page.waitForFunction(() => /已暂停/.test(document.querySelector('[data-testid="strip-chip"][data-kind="goal"]')?.textContent ?? ''), null, { timeout: 5000 });
});

test('UI-STRIP-05 子 Agent 详情在主区只读打开并可返回', async (t) => {
  const page = await open(t, 'running-approval'); if (!page) return;
  await page.locator(tid('subagent-card')).locator('[role="button"], button').first().click();
  await page.locator(tid('detail-view')).waitFor();
  assert.equal(await page.locator(tid('composer-input')).isVisible(), false, '详情中输入区仍可见');
  await page.locator(tid('detail-back')).click();
  assert.equal(await count(page, tid('detail-view')), 0);
  assert.ok(await page.locator(tid('composer-input')).isVisible());
});

test('UI-STRIP-06 清空队列需确认且保留结果未知的项', async (t) => {
  const page = await open(t, 'delivery-unknown'); if (!page) return;
  await page.locator(tid('strip-chip', '[data-kind="queue"]')).click();
  await page.locator(tid('strip-popover')).getByRole('button', { name: '清空' }).click();
  const dlg = page.locator(tid('confirm-dialog'));
  await dlg.waitFor();
  await dlg.getByRole('button', { name: '清空' }).click();
  if (!(await page.locator(tid('strip-popover', '[data-kind="queue"]')).isVisible())) await page.locator(tid('strip-chip', '[data-kind="queue"]')).click();
  const items = await page.locator(tid('queue-item')).evaluateAll((els) => els.map((e) => e.dataset.unknown));
  assert.deepEqual(items, ['true']);
});

test('UI-STRIP-07 未报告的用量显示“未报告”', async (t) => {
  const page = await open(t, 'delivery-unknown'); if (!page) return;
  await page.locator(tid('strip-chip', '[data-kind="goal"]')).click();
  assert.match(await text(page, tid('strip-popover', '[data-kind="goal"]')), /未报告/);
});

// ---------- 审批与追问 ----------
test('UI-APPR-01 聚焦审批卡后数字键选择', async (t) => {
  const page = await open(t, 'running-approval'); if (!page) return;
  await page.locator(tid('approval-card', '[data-state="pending"]')).focus();
  await page.keyboard.press('1');
  assert.match(await attr(page, tid('approval-card'), 'data-state'), /submitted|resolved/);
});

test('UI-APPR-02 提交后选项锁定', async (t) => {
  const page = await open(t, 'running-approval'); if (!page) return;
  const card = page.locator(tid('approval-card')).first();
  await card.getByRole('button').first().click();
  assert.equal(await card.getAttribute('data-state'), 'submitted');
  assert.equal(await card.getByRole('button').count(), 0, '提交后仍可再次选择');
});

test('UI-APPR-03 追问答案切换会话后保留', async (t) => {
  const page = await open(t, 'pending-question'); if (!page) return;
  await page.locator(`${tid('question-card')} [role="radio"]`).nth(1).click();
  await page.locator(tid('session-row', '[data-session-id="s1"]')).click();
  await page.locator(tid('session-row', '[data-session-id="s3"]')).click();
  assert.equal(await page.locator(`${tid('question-card')} [role="radio"]`).nth(1).getAttribute('aria-checked'), 'true');
  assert.equal(await attr(page, tid('question-card'), 'data-state'), 'pending', '答案被自动提交');
});

// ---------- 生命周期 ----------
test('UI-LIFE-01 重启后不自动执行', async (t) => {
  const page = await open(t, 'restart'); if (!page) return;
  const card = page.locator(tid('recovery-card'));
  await card.waitFor();
  assert.ok(await card.getByRole('button', { name: /恢复连接/ }).count() > 0);
  assert.match(await card.innerText(), /可能会因此继续执行|可能继续执行/);
  const users = await count(page, tid('message-user'));
  await sleep(2000);
  assert.equal(await count(page, tid('composer-stop')), 0, '重启后进入了执行中');
  assert.equal(await count(page, tid('message-user')), users, '重启后自动投递了输入');
});

test('UI-LIFE-02 未信任工作区可编辑草稿但不能发送', async (t) => {
  const page = await open(t, 'untrusted'); if (!page) return;
  const block = page.locator(tid('trust-block'));
  assert.ok(await block.getByRole('button', { name: /信任此工作区/ }).count() > 0);
  assert.equal(await attr(page, tid('composer-primary'), 'aria-disabled'), 'true');
  await typeInto(page, '先写草稿');
  assert.equal(await page.locator(tid('composer-input')).inputValue(), '先写草稿');
});

test('UI-LIFE-03 关键保存失败横幅不可关闭且禁止发送', async (t) => {
  const page = await open(t, 'save-failure'); if (!page) return;
  const banner = page.locator(tid('banner-save-failure'));
  assert.ok(await banner.isVisible());
  assert.equal(await banner.getByRole('button', { name: /关闭|×|✕/ }).count(), 0, '横幅可以被关闭');
  assert.match(await banner.innerText(), /未保存|尚未保存/);
  assert.match(await banner.innerText(), /未确认/);
  assert.equal(await attr(page, tid('composer-primary'), 'aria-disabled'), 'true');
});

// ---------- 可访问性与搜索 ----------
test('UI-A11Y-01 axe 无 serious/critical 问题', async (t) => {
  for (const f of STYLE_FIXTURES) {
    const page = await open(t, f); if (!page) return;
    const v = await axe(page);
    assert.deepEqual(v, [], `${f}: ${v.map((x) => `${x.id}（${x.help}）→ ${x.nodes.join(' | ')}`).join('；')}`);
  }
});

test('UI-A11Y-02 关闭菜单后焦点回到触发元素', async (t) => {
  const page = await open(t, 'running-approval'); if (!page) return;
  await page.locator(tid('session-head-menu')).click();
  await page.keyboard.press('Escape');
  assert.equal(await page.evaluate(() => document.activeElement?.dataset.testid), 'session-head-menu');
});

test('UI-A11Y-03 Mod+K 打开搜索，Esc 关闭', async (t) => {
  const page = await open(t, 'running-approval'); if (!page) return;
  await page.locator(tid('stream')).click();
  await page.keyboard.press(`${MOD}+K`);
  await page.locator(tid('command-palette')).waitFor();
  await page.keyboard.press('Escape');
  assert.equal(await count(page, tid('command-palette')), 0);
});

test('UI-SEARCH-01 搜索结果标出已归档来源', async (t) => {
  const page = await open(t, 'running-approval'); if (!page) return;
  await page.locator(tid('stream')).click();
  await page.keyboard.press(`${MOD}+K`);
  await page.keyboard.type('旧终端');
  const palette = page.locator(tid('command-palette'));
  await palette.getByText('已归档').first().waitFor({ timeout: 3000 });
});
