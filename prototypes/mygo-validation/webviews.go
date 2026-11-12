package main

import (
	"context"

	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

// webViewsTest：同一原生窗口中的两个 WebView（工作台 + 面板）是否各用一个进程。
func webViewsTest() map[string]any {
	r := map[string]any{}
	var w *mygo.Window
	var mainView, panelView *mygo.WebView
	mygo.RunOnMain(func() {
		w = mygo.NewWindow(mygo.WindowOptions{Title: "原生窗口 + 两个 WebView", Width: 900, Height: 500,
			Content: ui.View(func(c *ui.Context) {
				ui.Row(c).Fill().Gap(4).Children(func() {
					if mainView != nil {
						ui.WebView(c, mainView).Grow(2)
					}
					if panelView != nil {
						ui.WebView(c, panelView).Grow(1)
					}
				})
			})})
		mainView, _ = w.NewWebView(mygo.WebViewOptions{URL: "/lab/"})
		panelView, _ = w.NewWebView(mygo.WebViewOptions{URL: "aibopkg://pkg-a/panel.html"})
		w.Invalidate()
	})
	time.Sleep(3 * time.Second)
	if mainView == nil || panelView == nil {
		r["error"] = "web view not created"
		return r
	}
	ev := func(p *mygo.Page, code string) string {
		ctx, cancel := context.WithTimeout(context.Background(), 2*time.Second)
		defer cancel()
		_, err := p.EvalContext(ctx, code)
		return errString(err)
	}
	r["mainPid"] = mainView.Page().ProcessID()
	r["panelPid"] = panelView.Page().ProcessID()
	r["panelHasMygo"] = ev(panelView.Page(), "if (!window.mygo) throw new Error('no mygo'); return 1")
	_ = ev(panelView.Page(), "setTimeout(() => { for (;;) {} }, 0); return 1")
	time.Sleep(500 * time.Millisecond)
	r["mainEvalWhilePanelLoops"] = ev(mainView.Page(), "return 1")
	if pid := panelView.Page().ProcessID(); pid > 0 && pid != mainView.Page().ProcessID() {
		_ = killProcess(pid)
		time.Sleep(800 * time.Millisecond)
		r["mainEvalAfterPanelKilled"] = ev(mainView.Page(), "return 1")
		panelView.Page().Reload()
		time.Sleep(1500 * time.Millisecond)
		r["panelEvalAfterReload"] = ev(panelView.Page(), "return 1")
	}
	mygo.RunOnMain(func() { w.Close() })
	return r
}
