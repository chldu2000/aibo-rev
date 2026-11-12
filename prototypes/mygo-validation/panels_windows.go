package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"golang.org/x/sys/windows"
)

func probeRenderers(browser int) []int {
	script := fmt.Sprintf("Get-CimInstance Win32_Process -Filter 'ParentProcessId=%d' | Where-Object { $_.Name -eq 'msedgewebview2.exe' -and $_.CommandLine -match '--type=renderer' } | ForEach-Object { $_.ProcessId }", browser)
	cmd := exec.Command("powershell.exe", "-NoProfile", "-NonInteractive", "-Command", script)
	cmd.SysProcAttr = &syscall.SysProcAttr{HideWindow: true}
	b, err := cmd.Output()
	if err != nil {
		panic(err)
	}
	var ids []int
	for _, v := range strings.Fields(string(b)) {
		id, err := strconv.Atoi(v)
		if err != nil {
			panic(err)
		}
		ids = append(ids, id)
	}
	return ids
}

func probeProcess(pid int) map[string]any {
	h, err := windows.OpenProcess(windows.PROCESS_QUERY_LIMITED_INFORMATION, false, uint32(pid))
	if err != nil {
		return map[string]any{"alive": false, "openError": err.Error()}
	}
	defer windows.CloseHandle(h)
	var code uint32
	if err := windows.GetExitCodeProcess(h, &code); err != nil {
		panic(err)
	}
	var created, exited, kernel, user windows.Filetime
	if err := windows.GetProcessTimes(h, &created, &exited, &kernel, &user); err != nil {
		panic(err)
	}
	cpu := (uint64(kernel.HighDateTime)<<32 | uint64(kernel.LowDateTime)) + (uint64(user.HighDateTime)<<32 | uint64(user.LowDateTime))
	return map[string]any{"alive": code == 259, "cpuMs": float64(cpu) / 10000}
}

func runPanelsProbe(out string, recreateLoop bool) {
	rep := map[string]any{}
	var w *mygo.Window
	var a, b *mygo.WebView
	var mu sync.Mutex
	var events []map[string]any
	started := time.Now()
	defer func() {
		mu.Lock()
		rep["events"] = append([]map[string]any(nil), events...)
		mu.Unlock()
		data, err := json.MarshalIndent(rep, "", "  ")
		if err != nil {
			panic(err)
		}
		if err := os.WriteFile(out, data, 0o644); err != nil {
			panic(err)
		}
		mygo.App.Exit(0)
	}()
	onGone := func(name string) func(string) {
		return func(reason string) {
			mu.Lock()
			events = append(events, map[string]any{"panel": name, "reason": reason, "elapsedMs": time.Since(started).Milliseconds()})
			mu.Unlock()
		}
	}
	mygo.RunOnMain(func() {
		w = mygo.NewWindow(mygo.WindowOptions{Title: "17: two package panels", Width: 900, Height: 500, Content: ui.View(func(c *ui.Context) {
			ui.Row(c).Fill().Children(func() {
				if a != nil {
					ui.WebView(c, a).Grow(1)
				}
				if b != nil {
					ui.WebView(c, b).Grow(1)
				}
			})
		})})
		var err error
		a, err = w.NewWebView(mygo.WebViewOptions{URL: "aibopkg://pkg-a/panel.html"})
		if err != nil {
			panic(err)
		}
		a.Page().OnRenderProcessGone(onGone("a"))
		w.Invalidate()
	})
	eval := func(v *mygo.WebView, code string) any {
		ctx, cancel := context.WithTimeout(context.Background(), time.Second)
		defer cancel()
		r, err := v.Page().EvalContext(ctx, code)
		if err != nil {
			return map[string]any{"error": err.Error()}
		}
		return r
	}
	state := "return {origin:location.origin, secureContext:isSecureContext, ready:document.readyState, now:Date.now(), hasMygo:!!window.mygo, bindingResult:document.getElementById('info').textContent}"
	time.Sleep(2 * time.Second)
	browser := a.Page().ProcessID()
	first := probeRenderers(browser)
	rep["browserA"] = browser
	rep["renderersBeforeB"] = first
	rep["aInitial"] = eval(a, state)
	mygo.RunOnMain(func() {
		var err error
		b, err = w.NewWebView(mygo.WebViewOptions{URL: "aibopkg://pkg-b/panel.html"})
		if err != nil {
			panic(err)
		}
		b.Page().OnRenderProcessGone(onGone("b"))
		w.Invalidate()
	})
	time.Sleep(2 * time.Second)
	rep["browserB"] = b.Page().ProcessID()
	rep["renderersAfterB"] = probeRenderers(browser)
	rep["bInitial"] = eval(b, state)
	if len(first) != 1 {
		panic(fmt.Sprintf("cannot uniquely identify A renderer: %v", first))
	}
	rep["aLoopStartMs"] = time.Since(started).Milliseconds()
	rep["startLoop"] = eval(a, "setTimeout(()=>{for(;;){}},0); return 1")
	time.Sleep(12 * time.Second)
	rep["bWhileALoops"] = eval(b, state)
	rep["aWhileLooping"] = eval(a, state)
	if recreateLoop {
		rep["oldRendererBeforeDestroy"] = probeProcess(first[0])
		t0 := time.Now()
		mygo.RunOnMain(func() {
			a.Destroy()
			var err error
			a, err = w.NewWebView(mygo.WebViewOptions{URL: "aibopkg://pkg-a/panel.html"})
			if err != nil {
				panic(err)
			}
			w.Invalidate()
		})
		rep["destroyAndCreateMs"] = float64(time.Since(t0).Microseconds()) / 1000
		time.Sleep(2 * time.Second)
		rep["aAfterLoopRecreate"] = eval(a, state)
		rep["bAfterLoopRecreate"] = eval(b, state)
		rep["oldRendererAfterRecreate"] = probeProcess(first[0])
		time.Sleep(time.Second)
		rep["oldRendererOneSecondLater"] = probeProcess(first[0])
		rep["renderersAfterLoopRecreate"] = probeRenderers(browser)
		// 专用实验进程清理；不触碰其他应用的 WebView2 进程。
		if err := killProcess(browser); err != nil {
			panic(err)
		}
		return
	}
	rep["rendererKilled"] = first[0]
	if err := killProcess(first[0]); err != nil {
		panic(err)
	}
	time.Sleep(500 * time.Millisecond)
	rep["bAfterARendererKilled"] = eval(b, state)
	mygo.RunOnMain(func() { a.Page().Reload() })
	time.Sleep(time.Second)
	rep["aAfterRendererReload"] = eval(a, state)
	rep["browserKilled"] = browser
	if err := killProcess(browser); err != nil {
		panic(err)
	}
	time.Sleep(500 * time.Millisecond)
	rep["aAfterBrowserKilled"] = eval(a, state)
	rep["bAfterBrowserKilled"] = eval(b, state)
	mygo.RunOnMain(func() { a.Page().Reload(); b.Page().Reload() })
	time.Sleep(time.Second)
	rep["aAfterBrowserReload"] = eval(a, state)
	rep["bAfterBrowserReload"] = eval(b, state)
	// 只作恢复实验；是否采用重建 WebView 的 Windows 策略由决议确认。
	mygo.RunOnMain(func() {
		a.Destroy()
		b.Destroy()
		var err error
		a, err = w.NewWebView(mygo.WebViewOptions{URL: "aibopkg://pkg-a/panel.html"})
		if err != nil {
			panic(err)
		}
		b, err = w.NewWebView(mygo.WebViewOptions{URL: "aibopkg://pkg-b/panel.html"})
		if err != nil {
			panic(err)
		}
		w.Invalidate()
	})
	time.Sleep(2 * time.Second)
	rep["aAfterRecreate"] = eval(a, state)
	rep["bAfterRecreate"] = eval(b, state)
}
