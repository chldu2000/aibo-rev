package main

import (
	"image/png"
	"os"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

// SHOTS=1：三套组件方案的深浅色截图。
func TestKitShots(t *testing.T) {
	if os.Getenv("SHOTS") == "" {
		t.Skip("SHOTS=1 时运行")
	}
	for v := range variantNames {
		for _, dark := range []bool{false, true} {
			app := newApp(400)
			app.dark = dark
			app.sh.glass = true
			app.sh.variant = v
			app.startLong()
			app.queue = []string{"补充测试用例"}
			app.running = true
			app.chips = []Chip{{Kind: "file", Label: "src/core/session.ts"}}
			tt := ui.NewTesterWith(app.view, 1280, 820, ui.TesterOptions{Dark: dark})
			tt.SetScale(2)
			tt.Frame()
			app.list.ScrollToEnd()
			tt.Advance(time.Second)
			tt.Frame()
			f, _ := os.Create("reports/kit-" + string(rune('a'+v)) + map[bool]string{false: "-light", true: "-dark"}[dark] + ".png")
			_ = png.Encode(f, tt.Image())
			f.Close()
		}
	}
}
