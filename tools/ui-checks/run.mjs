#!/usr/bin/env node
/**
 * 界面硬性要求的本地验证入口（docs/design/ui-requirements.md）。
 *   node tools/ui-checks/run.mjs static         静态检查（令牌、对比度、样式规则、禁用文案）
 *   node tools/ui-checks/run.mjs requirements   要求文档自身的完整性与测试映射
 *   node tools/ui-checks/run.mjs e2e [--target 名称] [--browser webkit|chromium] [--grep 编号]
 *   node tools/ui-checks/run.mjs selftest       用 selftest/ 中的样例确认各静态规则能命中违规
 *   node tools/ui-checks/run.mjs all            依次运行 selftest、static、requirements、e2e
 * 任一错误时以非零状态退出。
 */
import { spawnSync } from 'node:child_process';
import { abs, config } from './lib/common.mjs';
import { tokensComplete, contrast, tokenSync } from './checks/tokens.mjs';
import { noLeftAccent, noHardcodedColor, minFontSize, focusOutline, reducedMotion } from './checks/styles.mjs';
import { forbiddenCopy } from './checks/copy.mjs';
import { requirements } from './checks/requirements.mjs';

export const STATIC = {
  'tokens-complete': tokensComplete,
  contrast,
  'token-sync': tokenSync,
  'no-left-accent': noLeftAccent,
  'no-hardcoded-color': noHardcodedColor,
  'min-font-size': minFontSize,
  'focus-outline': focusOutline,
  'reduced-motion': reducedMotion,
  'forbidden-copy': forbiddenCopy,
};

function print(r) {
  const mark = r.ok ? (r.warnings.length ? '△' : '✓') : '✗';
  console.log(`${mark} ${r.name}${r.notes.length ? `  — ${r.notes.join('；')}` : ''}`);
  for (const e of r.errors) console.log(`    ✗ ${e}`);
  for (const w of r.warnings) console.log(`    △ ${w}`);
  return r.ok;
}

function runStatic() {
  console.log('\n静态检查');
  return Object.values(STATIC).map((fn) => print(fn())).every(Boolean);
}
function runRequirements() {
  console.log('\n要求文档');
  return print(requirements(STATIC));
}
function runE2e(args) {
  console.log('\n验收测试（e2e）');
  const gi = args.indexOf('--grep');
  const pattern = gi >= 0 ? [`--test-name-pattern=${args[gi + 1]}`] : [];
  const res = spawnSync(process.execPath, ['--test', '--test-concurrency=1', '--test-reporter=spec', ...pattern, abs('tools/ui-checks/e2e/acceptance.test.mjs')], {
    stdio: 'inherit',
    env: { ...process.env, UI_CHECKS_ARGS: JSON.stringify(args) },
  });
  return res.status === 0;
}

/** 规则自检：bad/ 中每条规则都必须报错，good/ 中不得报错（contrast/tokens-complete 只看 tokens.json，不参与） */
function runSelftest() {
  console.log('\n规则自检');
  const saved = { style: config.styleRoots, copy: config.copyRoots };
  const fileRules = ['token-sync', 'no-left-accent', 'no-hardcoded-color', 'min-font-size', 'focus-outline', 'reduced-motion', 'forbidden-copy'];
  let ok = true;
  const run = (dir) => { config.styleRoots = [dir]; config.copyRoots = [dir]; return Object.fromEntries(fileRules.map((k) => [k, STATIC[k]()])); };
  const bad = run('tools/ui-checks/selftest/bad');
  const good = run('tools/ui-checks/selftest/good');
  for (const k of fileRules) {
    // token-sync 同时核对生成文件，自检只看样式中的令牌声明错误
    const badErrors = bad[k].errors.filter((e) => k !== 'token-sync' || e.includes('selftest'));
    const goodErrors = good[k].errors.filter((e) => k !== 'token-sync' || e.includes('selftest'));
    const pass = badErrors.length > 0 && goodErrors.length === 0;
    ok &&= pass;
    console.log(`${pass ? '✓' : '✗'} ${k}  — 违规样例命中 ${badErrors.length} 处，合规样例误报 ${goodErrors.length} 处`);
    for (const e of goodErrors) console.log(`    ✗ 误报：${e}`);
  }
  // left-accent 三种形态都要命中
  const la = bad['no-left-accent'].errors.join('\n');
  for (const [label, re] of [['border-left', /border-left/], ['inset 阴影', /inset/], ['伪元素竖条', /伪元素/]]) {
    const hit = re.test(la); ok &&= hit;
    console.log(`  ${hit ? '✓' : '✗'} no-left-accent 识别 ${label}`);
  }
  config.styleRoots = saved.style; config.copyRoots = saved.copy;
  return ok;
}

const [cmd = 'all', ...rest] = process.argv.slice(2);
const results = [];
if (cmd === 'selftest' || cmd === 'all') results.push(runSelftest());
if (cmd === 'static' || cmd === 'all') results.push(runStatic());
if (cmd === 'requirements' || cmd === 'all') results.push(runRequirements());
if (cmd === 'e2e' || cmd === 'all') results.push(runE2e(rest));
if (!results.length) { console.error(`未知命令 ${cmd}`); process.exit(2); }
const ok = results.every(Boolean);
console.log(ok ? '\n全部通过' : '\n存在未通过的检查');
process.exit(ok ? 0 : 1);
