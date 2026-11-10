#!/usr/bin/env node
/** 从 tokens.json 生成 tokens.css。用法：node tools/ui-checks/build-tokens.mjs */
import { writeFileSync } from 'node:fs';
import { abs, config, read, rel } from './lib/common.mjs';

export function renderTokensCss(tokens) {
  const block = (sel, entries) => `${sel} {\n${entries.map(([k, v]) => `  --${k}: ${v};`).join('\n')}\n}\n`;
  const groups = ['font', 'space', 'radius', 'size', 'z', 'motion', 'focus'];
  const misc = groups.flatMap((g) => Object.entries(tokens[g] ?? {}).filter(([k]) => !k.startsWith('$')).map(([k, v]) => [`${g}-${k}`, v]));
  return [
    '/* 由 tools/ui-checks/build-tokens.mjs 从 tokens.json 生成，请勿手改。 */',
    block(':root, [data-theme="light"]', Object.entries(tokens.color.light)),
    block('[data-theme="dark"]', Object.entries(tokens.color.dark)),
    block(':root', misc),
  ].join('\n');
}

if (import.meta.url === `file://${process.argv[1]}`) {
  const tokens = JSON.parse(read(config.tokens));
  writeFileSync(abs(config.tokensCss), renderTokensCss(tokens));
  console.log(`已生成 ${rel(abs(config.tokensCss))}`);
}
