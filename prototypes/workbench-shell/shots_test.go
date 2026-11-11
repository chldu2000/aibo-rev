package main

import (
	"image/png"
	"os"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

// SHOTS=1：标准与玻璃外观、深浅色、减少透明度的截图（测试器 CPU 绘制，没有窗口材质）。
func TestShots(t *testing.T) {
	if os.Getenv("SHOTS") == "" {
		t.Skip("SHOTS=1 时运行")
	}
	_ = os.MkdirAll("reports", 0o755)
	type variant struct {
		name         string
		glass, dark  bool
		reduce, high bool
	}
	for _, v := range []variant{
		{"standard-light", false, false, false, false},
		{"glass-light", true, false, false, false},
		{"glass-dark", true, true, false, false},
		{"glass-reduce-transparency", true, false, true, false},
		{"glass-increase-contrast", true, false, false, true},
	} {
		app := newApp(400)
		app.dark = v.dark
		app.sh.glass = v.glass
		app.startLong()
		app.queue = []string{"补充测试用例"}
		app.running = true
		app.chips = []Chip{{Kind: "file", Label: "src/core/session.ts"}}
		tt := ui.NewTesterWith(app.view, 1280, 820, ui.TesterOptions{Dark: v.dark, Preferences: ui.Preferences{ReduceTransparency: v.reduce, HighContrast: v.high}})
		tt.SetScale(2)
		tt.Frame()
		app.list.ScrollTo(len(app.items)-14, ui.Start) // 让内容滚到标题栏与输入区下方
		tt.Advance(time.Second)
		tt.Frame()
		f, err := os.Create("reports/shot-" + v.name + ".png")
		if err != nil {
			t.Fatal(err)
		}
		_ = png.Encode(f, tt.Image())
		f.Close()
	}
}
