import { existsSync } from 'node:fs';
import { abs, config, read, Report } from '../lib/common.mjs';

const E2E_FILE = 'tools/ui-checks/e2e/acceptance.test.mjs';
const LEVELS = ['必须', '不得', '建议'];
const FIELDS = ['条款', '依据', 'Checklist', '验收', '验证'];

export function parseRequirements(text) {
  const reqs = [];
  let cur = null;
  text.split('\n').forEach((line, i) => {
    const h = line.match(/^#### (UI-[A-Z0-9]+-\d{2}) · (\S+) · (.+)$/);
    if (h) { cur = { id: h[1], level: h[2], title: h[3], line: i + 1, fields: {} }; reqs.push(cur); return; }
    if (/^#{1,4} /.test(line)) { cur = null; return; }
    const f = cur && line.match(/^- ([^：]+)：(.*)$/);
    if (f) cur.fields[f[1].trim()] = f[2].trim();
  });
  return reqs;
}

export function e2eTitles() {
  if (!existsSync(abs(E2E_FILE))) return [];
  return [...read(E2E_FILE).matchAll(/test\(\s*['"`](UI-[A-Z0-9]+-\d{2})\b/g)].map((m) => m[1]);
}

export function requirements(staticChecks) {
  const r = new Report('requirements');
  const reqs = parseRequirements(read(config.requirements));
  const inv = read(config.inventory);
  const checked = new Set([...inv.matchAll(/^- \[x\] ((?:F\d{2}\.\d+)|(?:D\d+(?:\.\d+)?))/gm)].map((m) => m[1]));
  const all = new Set([...inv.matchAll(/^- \[[ x]\] ((?:F\d{2}\.\d+)|(?:D\d+(?:\.\d+)?))/gm)].map((m) => m[1]));
  const tests = e2eTitles();
  const seen = new Map();
  const referenced = new Set();
  const counts = { 必须: 0, 不得: 0, 建议: 0, auto: 0, manualOnly: 0 };

  for (const q of reqs) {
    const at = `${config.requirements}:${q.line}`;
    if (seen.has(q.id)) r.error(`编号重复（首次在第 ${seen.get(q.id)} 行）`, `${at} ${q.id}`);
    seen.set(q.id, q.line);
    if (!LEVELS.includes(q.level)) r.error(`等级“${q.level}”不是 ${LEVELS.join('/')}`, `${at} ${q.id}`);
    else counts[q.level]++;
    for (const f of FIELDS) if (!q.fields[f]) r.error(`缺少字段“${f}”`, `${at} ${q.id}`);
    for (const ref of (q.fields.Checklist ?? '').split(/[、,，\s]+/).filter(Boolean)) {
      if (!all.has(ref)) r.error(`Checklist 中没有 ${ref}`, `${at} ${q.id}`);
      else if (!checked.has(ref)) r.error(`${ref} 未勾选，不能作为必需功能的依据`, `${at} ${q.id}`);
      referenced.add(ref);
    }
    const methods = (q.fields['验证'] ?? '').split(/[、,，\s]+/).filter(Boolean);
    let auto = false;
    for (const m of methods) {
      if (m === 'manual') continue;
      if (m === 'e2e' || m === 'perf') {
        auto = true;
        if (!tests.includes(q.id)) r.error(`声明了 ${m}，但 ${E2E_FILE} 中没有以 ${q.id} 开头的测试`, `${at} ${q.id}`);
        continue;
      }
      const s = m.match(/^static:(.+)$/);
      if (s) { auto = true; if (!staticChecks[s[1]]) r.error(`未知的静态检查 ${s[1]}`, `${at} ${q.id}`); continue; }
      r.error(`未知的验证方式“${m}”`, `${at} ${q.id}`);
    }
    if (auto) counts.auto++; else counts.manualOnly++;
  }
  for (const t of new Set(tests)) if (!seen.has(t)) r.error(`测试引用了不存在的要求 ${t}`, E2E_FILE);

  const uncovered = [...checked].filter((c) => !referenced.has(c));
  if (uncovered.length) r.warn(`以下已勾选的 Checklist 项尚无界面要求覆盖（可能不涉及界面，或待补）：${uncovered.join('、')}`);
  r.note(`${reqs.length} 条要求：必须 ${counts['必须']}、不得 ${counts['不得']}、建议 ${counts['建议']}；可自动验证 ${counts.auto}，仅人工 ${counts.manualOnly}；覆盖已勾选 Checklist ${referenced.size}/${checked.size}`);
  return r;
}
