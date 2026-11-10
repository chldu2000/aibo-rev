import { readFileSync, readdirSync, statSync, existsSync } from 'node:fs';
import { join, relative, resolve, dirname, extname } from 'node:path';
import { fileURLToPath } from 'node:url';

export const ROOT = resolve(dirname(fileURLToPath(import.meta.url)), '../../..');
export const rel = (p) => relative(ROOT, p);
export const abs = (p) => resolve(ROOT, p);
export const read = (p) => readFileSync(abs(p), 'utf8');
export const config = JSON.parse(read('tools/ui-checks/ui-checks.config.json'));

const SKIP_DIRS = new Set(['node_modules', 'dist', '.git', '.cache', '.runtime']);

/** 递归列出 roots 下扩展名匹配的文件（相对仓库根） */
export function walk(roots, exts) {
  const out = [];
  const visit = (p) => {
    if (!existsSync(p)) return;
    const st = statSync(p);
    if (st.isDirectory()) {
      for (const name of readdirSync(p)) if (!SKIP_DIRS.has(name)) visit(join(p, name));
    } else if (exts.includes(extname(p))) out.push(rel(p));
  };
  for (const r of roots) visit(abs(r));
  return out.sort();
}

/** 偏移 → 行号 */
export function lineIndex(text) {
  const starts = [0];
  for (let i = 0; i < text.length; i++) if (text[i] === '\n') starts.push(i + 1);
  return (offset) => {
    let lo = 0, hi = starts.length - 1;
    while (lo < hi) { const mid = (lo + hi + 1) >> 1; if (starts[mid] <= offset) lo = mid; else hi = mid - 1; }
    return lo + 1;
  };
}

/**
 * 已说明的例外：
 * - 同一行或上一行有 `ui-check-allow: <rule> <理由>`；
 * - 或处在 `ui-check-allow-begin: <rule> <理由>` 与 `ui-check-allow-end` 之间（用于非产品界面的整块样式）。
 */
export function allowed(lines, line, rule) {
  const re = new RegExp(`ui-check-allow:\\s*${rule}\\b`);
  if (re.test(lines[line - 1] ?? '') || re.test(lines[line - 2] ?? '')) return true;
  const begin = new RegExp(`ui-check-allow-begin:\\s*(${rule}|all)\\b`);
  for (let i = line - 1; i >= 0; i--) {
    if (/ui-check-allow-end/.test(lines[i] ?? '')) return false;
    if (begin.test(lines[i] ?? '')) return true;
  }
  return false;
}

export class Report {
  constructor(name) { this.name = name; this.errors = []; this.warnings = []; this.notes = []; }
  error(msg, where) { this.errors.push(where ? `${where}  ${msg}` : msg); }
  warn(msg, where) { this.warnings.push(where ? `${where}  ${msg}` : msg); }
  note(msg) { this.notes.push(msg); }
  get ok() { return this.errors.length === 0; }
}
