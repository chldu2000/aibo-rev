// 原生 UI 原型（一次性实验，不是产品实现）：会话流与输入区。
//
//	go generate && CGO_ENABLED=0 go build -o bin/native . && ./bin/native
//	MYGO_FRAME_STATS=8 ./bin/native -auto reports/auto.json      自动跑一次流式长回复并记录帧耗时
//	go test ./...                                                  无窗口的交互检查
package main

//go:generate sh -c "cd ../../tools/ui-checks && go run . tokens"

import (
	"encoding/json"
	"flag"
	"log"
	"os"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

func main() {
	auto := flag.String("auto", "", "自动测量并写出报告")
	flag.Parse()
	app := newApp(10000)
	mygo.App.SetName("Aibo Native UI Prototype")
	mygo.App.WhenReady(func() {
		app.dark = mygo.Theme.IsDark()
		mygo.Theme.OnUpdated(func() {
			app.dark = mygo.Theme.IsDark()
			app.md.Theme = nil
			if app.win != nil {
				app.win.Invalidate()
			}
		})
		app.win = mygo.NewWindow(mygo.WindowOptions{
			Title: "Aibo 原生 UI 原型", Width: 1280, Height: 820, MinWidth: SizeWindowMinWidth, MinHeight: SizeWindowMinHeight,
			Content: ui.View(app.view),
		})
		if *auto != "" {
			go runAuto(app, *auto)
		}
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}

func runAuto(app *App, out string) {
	rep := map[string]any{}
	wait := func(d time.Duration) { time.Sleep(d) }
	wait(1500 * time.Millisecond)
	snap := func() map[string]any {
		var r map[string]any
		done := make(chan struct{})
		app.win.Update(func() {
			avg, mx := app.perf.summary()
			r = map[string]any{"items": len(app.items) + len(app.older), "loaded": len(app.items), "builds": app.perf.builds,
				"buildAvgMs": ms(app.perf.buildTotal) / float64(max(app.perf.builds, 1)), "buildMaxMs": ms(app.perf.buildMax),
				"publishToBuildAvgMs": ms(avg), "publishToBuildMaxMs": ms(mx), "streamFrames": app.perf.streamFrames, "atEnd": app.list.AtEnd()}
			close(done)
		})
		<-done
		return r
	}
	reset := func() {
		app.win.Update(func() { app.perf = perf{} })
	}
	reset()
	wait(500 * time.Millisecond)
	rep["idle"] = snap()
	// 1. 流式长回复（需审批）
	reset()
	app.win.Update(func() { app.startLong() })
	wait(500 * time.Millisecond)
	app.win.Update(func() {
		for _, it := range app.items {
			if it.Kind == KindApproval && it.Approval == "pending" {
				app.approve(it, true)
			}
		}
	})
	for i := 0; i < 200; i++ {
		wait(100 * time.Millisecond)
		running := true
		done := make(chan struct{})
		app.win.Update(func() { running = app.running; close(done) })
		<-done
		if !running {
			break
		}
	}
	wait(800 * time.Millisecond)
	rep["streaming"] = snap()
	// 2. 载入全部 10,000 条，滚到中部
	reset()
	app.win.Update(func() { app.loadAll() })
	wait(500 * time.Millisecond)
	app.win.Update(func() { app.list.ScrollTo(len(app.items)/2, ui.Start) })
	wait(800 * time.Millisecond)
	rep["allHistory"] = snap()
	app.win.Update(func() { app.list.ScrollToEnd() })
	wait(800 * time.Millisecond)
	b, _ := json.MarshalIndent(rep, "", "  ")
	_ = os.WriteFile(out, b, 0o644)
	mygo.App.Exit(0)
}
