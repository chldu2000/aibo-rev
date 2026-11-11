package main

import (
	"image/png"
	"os"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func TestGlassMenu(t *testing.T) {
	app := newApp(50)
	app.sh.glass = true
	tt := ui.NewTester(app.view, 1280, 820)
	tt.Frame()
	if err := tt.Click("更多"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !tt.HasText("标准外观") || !tt.HasText("故障注入") {
		t.Fatalf("menu not open: %q", tt.Texts())
	}
	// 键盘：第一个可选项是“标准外观”
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if app.sh.glass {
		t.Fatal("keyboard choice did not apply")
	}
	if tt.HasText("标准外观") {
		t.Fatal("menu still open")
	}
}

func TestMenuShots(t *testing.T) {
	if os.Getenv("SHOTS") == "" {
		t.Skip("SHOTS=1 时运行")
	}
	for _, dark := range []bool{false, true} {
		for _, which := range []string{"更多", "面板操作"} {
			app := newApp(400)
			app.dark = dark
			app.sh.glass = true
			tt := ui.NewTesterWith(app.view, 1280, 820, ui.TesterOptions{Dark: dark})
			tt.SetScale(2)
			tt.Frame()
			_ = tt.Click(which)
			tt.Frame()
			if which == "更多" {
				if b, ok := tt.Find("故障注入"); ok {
					tt.Move(b.X+b.W/2, b.Y+b.H/2)
				}
			} else if b, ok := tt.Find("让面板页面卡死"); ok {
				tt.Move(b.X+b.W/2, b.Y+b.H/2)
			}
			tt.Advance(time.Second)
			tt.Frame()
			name := map[string]string{"更多": "more", "面板操作": "panel"}[which]
			f, _ := os.Create("reports/menu-" + name + map[bool]string{false: "-light", true: "-dark"}[dark] + ".png")
			_ = png.Encode(f, tt.Image())
			f.Close()
		}
	}
}
