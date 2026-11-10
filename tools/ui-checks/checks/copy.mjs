import { config, read, walk, Report } from '../lib/common.mjs';

export function forbiddenCopy() {
  const r = new Report('forbidden-copy');
  const files = walk(config.copyRoots, ['.svelte', '.html', '.js', '.ts', '.json']);
  for (const file of files) {
    read(file).split('\n').forEach((line, i) => {
      for (const p of config.forbiddenCopy) if (line.includes(p)) r.error(`出现禁用文案“${p}”（UI-GEN-03 / UI-COMP-08）`, `${file}:${i + 1}`);
    });
  }
  r.note(`扫描 ${files.length} 个文件`);
  return r;
}
