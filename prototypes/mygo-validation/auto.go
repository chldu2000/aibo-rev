package main

import (
	"bytes"
	"context"
	"encoding/json"
	"image"
	"image/color"
	"image/png"
	"syscall"
	"time"

	"github.com/egoist/mygo"
)

func trayIcon() []byte {
	img := image.NewNRGBA(image.Rect(0, 0, 32, 32))
	for y := 6; y < 26; y++ {
		for x := 6; x < 26; x++ {
			img.Set(x, y, color.NRGBA{A: 255})
		}
	}
	var b bytes.Buffer
	_ = png.Encode(&b, img)
	return b.Bytes()
}

func waitDOM(d time.Duration) bool {
	select {
	case <-domReady:
		return true
	case <-time.After(d):
		return false
	}
}

func drain() {
	for {
		select {
		case <-domReady:
		default:
			return
		}
	}
}

func eval(code string, d time.Duration) (any, error) {
	p := page()
	if p == nil {
		return nil, context.Canceled
	}
	ctx, cancel := context.WithTimeout(context.Background(), d)
	defer cancel()
	return p.EvalContext(ctx, code)
}

func evalJSON(code string, d time.Duration) any {
	v, err := eval(code, d)
	if err != nil {
		return map[string]any{"error": err.Error()}
	}
	if s, ok := v.(string); ok {
		var out any
		if json.Unmarshal([]byte(s), &out) == nil {
			return out
		}
	}
	return v
}

// mainThreadLatency 在 dur 内每 50ms 调用一次 RunOnMain，返回最大延迟（毫秒）。
func mainThreadLatency(dur time.Duration) float64 {
	max := 0.0
	end := time.Now().Add(dur)
	for time.Now().Before(end) {
		s := time.Now()
		done := make(chan struct{})
		mygo.RunOnMain(func() { close(done) })
		select {
		case <-done:
		case <-time.After(3 * time.Second):
			return 3000
		}
		if ms := float64(time.Since(s).Microseconds()) / 1000; ms > max {
			max = ms
		}
		time.Sleep(50 * time.Millisecond)
	}
	return max
}

func runAuto() map[string]any {
	rep := map[string]any{"mygo": "v0.9.0"}
	waitDOM(15 * time.Second)
	time.Sleep(1500 * time.Millisecond)
	rep["ua"] = evalJSON("return navigator.userAgent", 5*time.Second)
	rep["origin"] = evalJSON("return location.origin", 5*time.Second)
	rep["pid"] = page().ProcessID()

	// 1. 默认界面原型（Svelte + Bits UI + virtua + Shiki Worker）
	rep["defaultUI"] = evalJSON(defaultUIScript, 60*time.Second)

	// 2. 实验页
	drain()
	_ = page().LoadURL("/lab/")
	waitDOM(10 * time.Second)
	time.Sleep(800 * time.Millisecond)
	rep["lab"] = evalJSON("return JSON.stringify(await window.runLab())", 40*time.Second)

	// 3. 页面死循环：原生主线程是否仍可用，Reload / LoadURL 能否恢复，结束渲染进程后能否恢复
	// （实验页最后一步已让 pkg-b 面板死循环，此时页面应仍卡住）
	hang := map[string]any{}
	_, err := eval("return 1", 1500*time.Millisecond)
	hang["evalWhilePanelLooping"] = errString(err)
	hang["mainThreadMaxMs"] = mainThreadLatency(2 * time.Second)
	drain()
	t := time.Now()
	page().Reload()
	hang["reloadRecovered"] = waitDOM(6 * time.Second)
	drain()
	_ = page().LoadURL("/")
	hang["loadURLRecovered"] = waitDOM(6 * time.Second)
	hang["recovery"] = killAndReload()
	_ = t
	rep["hang"] = hang

	// 4. 主页面自身死循环 → 结束渲染进程 → Reload
	time.Sleep(1 * time.Second)
	_, _ = eval("setTimeout(() => { for (;;) {} }, 0); return 1", 2*time.Second)
	time.Sleep(500 * time.Millisecond)
	_, err = eval("return 1", 1500*time.Millisecond)
	rep["mainLoop"] = map[string]any{"evalDuringHang": errString(err), "recovery": killAndReload()}

	// 5. 独立窗口是否使用独立的 WebContent 进程，以及其中死循环是否影响主窗口
	rep["separateWindow"] = separateWindowTest()
	rep["webViews"] = webViewsTest()

	logMu.Lock()
	rep["events"] = append([]logEntry(nil), entries...)
	logMu.Unlock()
	return rep
}

func killAndReload() map[string]any {
	r := map[string]any{}
	pid := page().ProcessID()
	r["pidBefore"] = pid
	if pid <= 0 {
		return r
	}
	for len(gone) > 0 {
		<-gone
	}
	_ = syscall.Kill(pid, syscall.SIGKILL)
	select {
	case g := <-gone:
		r["renderProcessGone"] = g
	case <-time.After(8 * time.Second):
		r["renderProcessGone"] = "timeout"
	}
	drain()
	t := time.Now()
	page().Reload()
	r["recovered"] = waitDOM(15 * time.Second)
	r["ms"] = time.Since(t).Milliseconds()
	r["pidAfter"] = page().ProcessID()
	time.Sleep(300 * time.Millisecond)
	_, err := eval("return 1", 3*time.Second)
	r["evalAfter"] = errString(err)
	return r
}

func separateWindowTest() map[string]any {
	r := map[string]any{}
	ready := make(chan struct{}, 1)
	var w2 *mygo.Window
	mygo.RunOnMain(func() {
		w2 = mygo.NewWindow(mygo.WindowOptions{Title: "面板窗口", Width: 360, Height: 200, URL: "aibopkg://pkg-a/panel.html"})
		w2.Page().OnDOMReady(func() {
			select {
			case ready <- struct{}{}:
			default:
			}
		})
	})
	select {
	case <-ready:
	case <-time.After(10 * time.Second):
		r["error"] = "panel window not ready"
		return r
	}
	time.Sleep(500 * time.Millisecond)
	p2 := w2.Page()
	r["mainPid"] = page().ProcessID()
	r["panelPid"] = p2.ProcessID()
	ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
	_, _ = p2.EvalContext(ctx, "setTimeout(() => { for (;;) {} }, 0); return 1")
	cancel()
	time.Sleep(500 * time.Millisecond)
	_, err := eval("return 1", 2*time.Second)
	r["mainEvalWhilePanelLoops"] = errString(err)
	if pid := p2.ProcessID(); pid > 0 && pid != page().ProcessID() {
		_ = syscall.Kill(pid, syscall.SIGKILL)
		time.Sleep(800 * time.Millisecond)
		_, err = eval("return 1", 2*time.Second)
		r["mainEvalAfterPanelKilled"] = errString(err)
	}
	mygo.RunOnMain(func() { w2.Close() })
	return r
}

func errString(err error) string {
	if err == nil {
		return "ok"
	}
	return err.Error()
}

// defaultUIScript 移植自 prototypes/stream-composer/scripts/webkit-check.mjs 中可用合成事件完成的部分。
const defaultUIScript = `
const wait = (ms) => new Promise((r) => setTimeout(r, ms));
const perf = () => Object.fromEntries([...document.querySelectorAll('.perf dt')].map((d, i) => [d.textContent, document.querySelectorAll('.perf dd')[i].textContent]));
const out = {};
out.workerSupported = typeof Worker === 'function';
document.querySelector('.perf .btn:nth-child(3)').click();
await wait(300);
document.querySelector('[data-approval] .btn.primary')?.click();
await wait(9000);
out.streaming = perf();
out.shikiBlocks = document.querySelectorAll('.codeblock .shiki').length;
out.codeBlocks = document.querySelectorAll('.codeblock').length;
const s = document.querySelector('.stream');
out.followedBottom = s.scrollHeight - s.scrollTop - s.clientHeight < 60;
// @ 候选：用 input 事件模拟输入
const ta = document.querySelector('textarea');
ta.focus();
ta.value = '看 @src/co';
ta.setSelectionRange(ta.value.length, ta.value.length);
ta.dispatchEvent(new InputEvent('input', { bubbles: true, data: 'o', inputType: 'insertText' }));
await wait(300);
out.pickerVisible = !!document.querySelector('.picker');
out.focusStaysInTextarea = document.activeElement?.tagName === 'TEXTAREA';
ta.value = '';
ta.dispatchEvent(new InputEvent('input', { bubbles: true, inputType: 'deleteContentBackward' }));
// 向前加载的滚动锚定
const top = () => s.getBoundingClientRect().top;
const vis = () => { const r = [...s.querySelectorAll('.row')].find((x) => x.getBoundingClientRect().bottom > top() + 1); return r ? { id: r.dataset.id, y: r.getBoundingClientRect().top - top() } : null; };
out.prepend = [];
for (let i = 0; i < 3; i++) { s.scrollTop = 150; await wait(40); const a = vis(); await wait(700); const b = vis(); out.prepend.push(a && b && a.id === b.id ? Math.round(b.y - a.y) : 'changed ' + a?.id + '→' + b?.id); }
// 载入全部历史
document.querySelector('.perf .btn:nth-child(1)').click();
await wait(500);
s.scrollTop = s.scrollHeight / 2;
await wait(800);
out.allHistory = perf();
out.slowPublishes = (window.__slow ?? []).slice(0, 10);
return JSON.stringify(out);
`
