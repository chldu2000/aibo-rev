/**
 * 代码高亮的主线程入口：请求转发给 Worker，结果按 (lang, code) 缓存。
 * 缓存命中时同步返回，虚拟列表重新挂载代码块时不会闪烁。
 */
const cache = new Map<string, string>();
const waiting = new Map<number, { key: string; resolve: (html: string) => void; reject: (e: Error) => void }>();
let worker: Worker | null = null;
let seq = 0;

function getWorker(): Worker {
  if (worker) return worker;
  worker = new Worker(new URL('./highlight.worker.ts', import.meta.url), { type: 'module' });
  worker.onmessage = (e: MessageEvent<{ id: number; html?: string; error?: string }>) => {
    const w = waiting.get(e.data.id);
    if (!w) return;
    waiting.delete(e.data.id);
    if (e.data.html !== undefined) { cache.set(w.key, e.data.html); w.resolve(e.data.html); }
    else w.reject(new Error(e.data.error));
  };
  return worker;
}

/** 启动后空闲时预热：在 Worker 中加载 Shiki，并用真实片段编译各语言的常用规则 */
export function preloadHighlighter() {
  const samples: Record<string, string> = {
    typescript: 'export interface A { b: string }\nexport function f(x: A): Promise<A> { return Promise.resolve(x) }',
    go: 'func (c *C) F(ctx context.Context) error {\n\tif err := c.g(); err != nil { return fmt.Errorf("x: %w", err) }\n\treturn nil\n}',
    bash: 'pnpm test --filter @aibo/core -- draft',
    json: '{ "a": [1, true, null] }',
  };
  for (const [lang, code] of Object.entries(samples)) void highlight(code, lang);
}

export function highlightCached(code: string, lang: string): string | undefined {
  return cache.get(lang + '\u0000' + code);
}

export function highlight(code: string, lang: string): Promise<string> {
  const key = lang + '\u0000' + code;
  const hit = cache.get(key);
  if (hit) return Promise.resolve(hit);
  return new Promise((resolve, reject) => {
    const id = ++seq;
    waiting.set(id, { key, resolve, reject });
    getWorker().postMessage({ id, code, lang });
  });
}
