package main

import (
	"encoding/json"
	"os"
	"syscall"
	"time"

	"github.com/egoist/mygo"
)

// onMain 在界面线程执行 fn 并等待其结束。
func (app *App) onMain(fn func()) {
	done := make(chan struct{})
	app.update(func() { fn(); close(done) })
	select {
	case <-done:
	case <-time.After(5 * time.Second):
	}
}

// mainLatency 测量界面线程处理一次投递的延迟。
func (app *App) mainLatency(n int) (avgMs, maxMs float64) {
	var total, mx time.Duration
	for i := 0; i < n; i++ {
		t := time.Now()
		app.onMain(func() {})
		d := time.Since(t)
		total += d
		mx = max(mx, d)
		time.Sleep(20 * time.Millisecond)
	}
	return ms(total) / float64(n), ms(mx)
}

func (app *App) panelState() map[string]any {
	v, err := app.evalPanel("return { ticks: window.__panel.ticks, loadedAt: window.__panel.loadedAt, w: innerWidth, h: innerHeight, focus: document.hasFocus(), active: document.activeElement && (document.activeElement.id || document.activeElement.tagName) }")
	if err != nil {
		return map[string]any{"error": shortErr(err)}
	}
	m, _ := v.(map[string]any)
	return m
}

func runAuto(app *App, out string) {
	rep := map[string]any{}
	save := func() {
		b, _ := json.MarshalIndent(rep, "", "  ")
		_ = os.WriteFile(out, b, 0o644)
	}
	wait := time.Sleep
	wait(2500 * time.Millisecond)

	// 1. 进程归属
	var panelPid int
	app.onMain(func() { panelPid = app.sh.panel.Page().ProcessID() })
	rep["1_processes"] = map[string]any{"app": syscall.Getpid(), "panel": panelPid}
	rep["1_panelInitial"] = app.panelState()

	// 2. 分隔条：面板宽度 360 → 480；窗口变宽后面板宽度保持
	app.onMain(func() { app.sh.panelW = 480; app.win.Invalidate() })
	wait(600 * time.Millisecond)
	s480 := app.panelState()
	app.onMain(func() { app.win.SetSize(1500, 820) })
	wait(800 * time.Millisecond)
	sWide := app.panelState()
	var mainW float32
	app.onMain(func() { mainW = float32(1500) - SizeSidebar - app.sh.panelW })
	app.onMain(func() { app.win.SetSize(1280, 820) })
	wait(600 * time.Millisecond)
	rep["2_split"] = map[string]any{"panelW480": s480, "afterWindow1500": sWide, "mainColumnW": mainW}

	// 3. 焦点：给面板键盘焦点，再回到原生输入框
	app.onMain(func() { app.sh.panel.Focus() })
	wait(400 * time.Millisecond)
	f1 := app.panelState()
	_, _ = app.evalPanel("document.getElementById('q').focus(); return 1")
	wait(200 * time.Millisecond)
	f2 := app.panelState()
	app.onMain(func() { app.win.Focus() })
	rep["3_focus"] = map[string]any{"afterWebViewFocus": f1, "afterInputFocus": f2}

	// 4. 收起 3 秒再展开：页面是否保留、计时器是否继续
	before := app.panelState()
	app.onMain(func() { app.sh.panelOpen = false; app.win.Invalidate() })
	wait(3 * time.Second)
	hidden := app.panelState()
	app.onMain(func() { app.sh.panelOpen = true; app.win.Invalidate() })
	wait(600 * time.Millisecond)
	after := app.panelState()
	rep["4_collapse"] = map[string]any{"before": before, "whileHidden": hidden, "afterExpand": after}
	save()

	// 5. 面板死循环：原生界面仍流式更新，界面线程延迟；随后恢复
	_, _ = app.evalPanel("setTimeout(() => { for (;;) {} }, 0); return 1")
	wait(500 * time.Millisecond)
	app.onMain(func() { app.perf = perf{} })
	app.onMain(func() { app.startLong() })
	wait(300 * time.Millisecond)
	app.onMain(func() {
		for _, it := range app.items {
			if it.Kind == KindApproval && it.Approval == "pending" {
				app.approve(it, true)
			}
		}
	})
	la, lm := app.mainLatency(30)
	for i := 0; i < 150; i++ {
		wait(100 * time.Millisecond)
		running := true
		app.onMain(func() { running = app.running })
		if !running {
			break
		}
	}
	var loop map[string]any
	app.onMain(func() {
		avg, mx := app.perf.summary()
		loop = map[string]any{"streamFrames": app.perf.streamFrames, "publishToBuildAvgMs": ms(avg), "publishToBuildMaxMs": ms(mx),
			"buildMaxMs": ms(app.perf.buildMax), "mainThreadAvgMs": la, "mainThreadMaxMs": lm}
	})
	loop["panelEvalWhileLooping"] = app.panelState()
	t := time.Now()
	app.reloadPanel()
	loop["recoverMs"] = ms(time.Since(t))
	app.onMain(func() { loop["status"] = app.sh.panelStatus; loop["newPanelPid"] = app.sh.panel.Page().ProcessID() })
	loop["panelAfterRecover"] = app.panelState()
	rep["5_panelLoop"] = loop
	save()

	// 6. 结束面板渲染进程：工作台照常，重新载入恢复
	app.onMain(func() { app.killPanel() })
	wait(800 * time.Millisecond)
	kill := map[string]any{}
	app.onMain(func() { kill["status"] = app.sh.panelStatus })
	kill["mainThreadAvgMs"], kill["mainThreadMaxMs"] = app.mainLatency(10)
	t = time.Now()
	app.reloadPanel()
	kill["recoverMs"] = ms(time.Since(t))
	kill["panelAfterRecover"] = app.panelState()
	rep["6_panelKilled"] = kill
	save()

	// 7. 区域级 panic：注入工具卡与候选弹层故障，界面继续工作
	app.onMain(func() { app.sh.inject["工具卡"] = true; app.sh.inject["侧栏"] = true; app.win.Invalidate() })
	wait(600 * time.Millisecond)
	pan := map[string]any{}
	app.onMain(func() {
		faults := map[string]string{}
		for k, v := range app.sh.faults {
			faults[k] = v
		}
		pan["faults"] = faults
		app.perf = perf{}
		app.draft = "panic 之后仍可编辑"
		app.win.Invalidate()
	})
	wait(400 * time.Millisecond)
	app.onMain(func() { pan["buildsAfterPanic"] = app.perf.builds; pan["draft"] = app.draft })
	app.onMain(func() {
		delete(app.sh.inject, "工具卡")
		delete(app.sh.inject, "侧栏")
		for k := range app.sh.faults {
			delete(app.sh.faults, k)
		}
		app.draft = ""
		app.win.Invalidate()
	})
	rep["7_panic"] = pan
	save()

	// 8. 关闭窗口后重开：宿主持有的界面状态是否恢复；面板页面是否保留
	app.onMain(func() {
		app.selectSession(2)
		app.draft = "会话 3 的草稿"
		app.sh.panelW = 420
		app.list.ScrollTo(5, 0)
	})
	wait(500 * time.Millisecond)
	beforeClose := app.panelState()
	var first int
	app.onMain(func() { first, _ = app.list.Visible() })
	app.onMain(func() { app.win.Close() })
	wait(1500 * time.Millisecond)
	var alive bool
	app.onMain(func() { alive = app.win == nil || !app.win.IsVisible() })
	app.onMain(func() { app.openWindow() })
	wait(2000 * time.Millisecond)
	reopen := map[string]any{"closeMode": *closeMode, "backgroundAfterClose": alive, "panelBeforeClose": beforeClose, "panelAfterReopen": app.panelState()}
	app.onMain(func() {
		f2, _ := app.list.Visible()
		reopen["session"] = app.sh.sessions[app.sh.cur].name
		reopen["draft"] = app.draft
		reopen["panelW"] = app.sh.panelW
		reopen["firstVisibleBefore"] = first
		reopen["firstVisibleAfter"] = f2
	})
	rep["8_reopen"] = reopen
	app.onMain(func() { rep["events"] = append([]string(nil), app.sh.events...) })
	save()
	mygo.App.Exit(0)
}
