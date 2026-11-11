package main

import (
	"image/png"
	"os"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func glassTester(t *testing.T) (*App, *ui.Tester) {
	app := newApp(50)
	app.sh.glass = true
	tt := ui.NewTester(app.view, 1280, 820)
	tt.Frame()
	return app, tt
}

// 完整映射：矩阵一次选定模型与推理强度。
func TestMatrixClick(t *testing.T) {
	app, tt := glassTester(t)
	if err := tt.Click("Opus 5.5 · 高"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !tt.HasText("由 Claude Code 提供的模型与推理强度") {
		t.Fatalf("matrix not open: %q", tt.Texts())
	}
	if _, ok := tt.Find("Sonnet 5.5 不支持 超高"); !ok {
		t.Fatal("unsupported cell not marked")
	}
	if err := tt.Click("Sonnet 5.5 · 低"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if app.model != "Sonnet 5.5" || app.effort != "低" || app.sh.matrix.open {
		t.Fatalf("model=%q effort=%q open=%v", app.model, app.effort, app.sh.matrix.open)
	}
}

// 键盘：方向键跳过不支持的组合，Enter 选择。
func TestMatrixKeyboard(t *testing.T) {
	app, tt := glassTester(t)
	_ = tt.Click("Opus 5.5 · 高")
	tt.Frame()
	tt.Frame()
	if !tt.Focused("Opus 5.5 · 高") {
		t.Fatal("focus not on current cell")
	}
	tt.Key(0, ui.KeyRight) // 超高
	tt.Key(0, ui.KeyDown)  // Sonnet 没有超高 → 最近的“高”
	tt.Frame()
	if !tt.Focused("Sonnet 5.5 · 高") {
		t.Fatalf("row=%d col=%d", app.sh.matrix.row, app.sh.matrix.col)
	}
	tt.Key(0, ui.KeyDown) // Haiku：整行一个格子
	tt.Frame()
	if !tt.Focused("Haiku 5.5") {
		t.Fatalf("row=%d col=%d", app.sh.matrix.row, app.sh.matrix.col)
	}
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if app.model != "Haiku 5.5" || app.sh.matrix.open {
		t.Fatalf("model=%q open=%v", app.model, app.sh.matrix.open)
	}
	if !tt.HasText("Haiku 5.5") || tt.HasText("Haiku 5.5 · 高") {
		t.Fatalf("trigger label: %q", tt.Texts())
	}
}

// 没有完整映射：两个下拉，推理强度只列当前模型支持的；换模型时不支持的强度退回该模型默认值。
func TestPerModelFallback(t *testing.T) {
	app, tt := glassTester(t)
	_ = tt.Click("ACP 连接层")
	tt.Frame()
	if _, ok := tt.Find("模型"); !ok {
		t.Fatalf("no model select: %q", tt.Texts())
	}
	if _, ok := tt.Find("推理强度"); !ok {
		t.Fatal("no effort select")
	}
	if app.model != "示例模型 Large" || app.effort != "中" {
		t.Fatalf("defaults %q %q", app.model, app.effort)
	}
	app.effort = "最低"
	app.setModel("示例模型 Mini")
	if app.effort != "低" {
		t.Fatalf("unsupported effort kept: %q", app.effort)
	}
	app.effort = "中"
	app.setModel("示例模型 Large")
	if app.effort != "中" {
		t.Fatalf("supported effort changed: %q", app.effort)
	}
	// 回到矩阵会话，各会话的选择互不影响
	_ = tt.Click("修复队列暂停逻辑")
	tt.Frame()
	if app.model != "Opus 5.5" || app.effort != "高" {
		t.Fatalf("session 0 lost its choice: %q %q", app.model, app.effort)
	}
}

func TestPickerShots(t *testing.T) {
	if os.Getenv("SHOTS") == "" {
		t.Skip("SHOTS=1 时运行")
	}
	for _, dark := range []bool{false, true} {
		for _, mode := range []string{"matrix", "fallback"} {
			app := newApp(400)
			app.dark = dark
			app.sh.glass = true
			tt := ui.NewTesterWith(app.view, 1280, 820, ui.TesterOptions{Dark: dark})
			tt.SetScale(2)
			tt.Frame()
			if mode == "matrix" {
				_ = tt.Click("Opus 5.5 · 高")
				tt.Frame()
				if b, ok := tt.Find("Sonnet 5.5 · 中"); ok {
					tt.Move(b.X+b.W/2, b.Y+b.H/2)
				}
			} else {
				_ = tt.Click("ACP 连接层")
				tt.Frame()
				_ = tt.Click("推理强度")
			}
			tt.Advance(time.Second)
			tt.Frame()
			f, _ := os.Create("reports/picker-" + mode + map[bool]string{false: "-light", true: "-dark"}[dark] + ".png")
			_ = png.Encode(f, tt.Image())
			f.Close()
		}
	}
}

// 当前模型不提供推理强度时，推理强度保留为禁用控件并说明原因（UI-COMP-10）。
func TestUnsupportedEffortKept(t *testing.T) {
	app, tt := glassTester(t)
	_ = tt.Click("ACP 连接层")
	tt.Frame()
	catalogPerModel.Models = append(catalogPerModel.Models, modelCap{Name: "示例模型 Tiny"})
	defer func() { catalogPerModel.Models = catalogPerModel.Models[:2] }()
	app.setModel("示例模型 Tiny")
	tt.Frame()
	if _, ok := tt.Find("推理强度"); !ok {
		t.Fatalf("effort control hidden: %q", tt.Texts())
	}
	if !tt.HasText("推理强度 —") {
		t.Fatal("disabled effort control missing")
	}
}
