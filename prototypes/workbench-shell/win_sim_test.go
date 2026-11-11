package main

import (
	"image"
	"image/color"
	"image/draw"
	"image/png"
	"os"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

// SHOTS=1：Windows 外观的近似模拟（不是 Windows 实测）。
//   - 标题栏几何按 Windows：窗口按钮在右上角，宽 3×46 DIP，高为 TitleBarHeight（44）。
//   - 窗口材质透明处填近似 Mica 的颜色（浅 #F3F3F3、深 #202020）；真实 Mica 会带壁纸色调。
//   - plugins/glass 的着色器与 macOS 相同（MyGO 文档），所以面板玻璃按同一实现绘制。
//   - 字体仍是 macOS 的字体（本机没有 Segoe UI / 微软雅黑），字形与字宽会不同。
func TestWindowsSim(t *testing.T) {
	if os.Getenv("SHOTS") == "" {
		t.Skip("SHOTS=1 时运行")
	}
	const scale = 1.5 // Windows 常见的 150% 缩放
	for _, dark := range []bool{false, true} {
		for _, panel := range []bool{true, false} {
			app := newApp(400)
			app.dark = dark
			app.sh.glass = true
			app.sh.variant = 0
			app.sh.panelOpen = panel
			app.startLong()
			app.queue = []string{"补充测试用例"}
			app.running = true
			app.chips = []Chip{{Kind: "file", Label: "src/core/session.ts"}}
			tt := ui.NewTesterWith(app.view, 1280, 820, ui.TesterOptions{Dark: dark})
			tt.SetScale(scale)
			tt.SetTitleBar(ui.TitleBar{Height: SizeHeader, Right: 138})
			tt.SetVibrancy(true)
			tt.Frame()
			app.list.ScrollToEnd()
			tt.Advance(time.Second)
			tt.Frame()
			src := tt.Image()
			b := src.Bounds()
			out := image.NewRGBA(b)
			mica := color.RGBA{0xF3, 0xF3, 0xF3, 0xFF}
			fg := color.RGBA{0x1A, 0x1A, 0x1A, 0xFF}
			if dark {
				mica, fg = color.RGBA{0x20, 0x20, 0x20, 0xFF}, color.RGBA{0xFF, 0xFF, 0xFF, 0xFF}
			}
			draw.Draw(out, b, &image.Uniform{mica}, image.Point{}, draw.Src)
			draw.Draw(out, b, src, b.Min, draw.Over)
			captionButtons(out, b.Max.X, int(SizeHeader*scale), int(46*scale), fg, scale)
			name := "reports/win-sim-" + map[bool]string{false: "light", true: "dark"}[dark] + map[bool]string{true: "-panel", false: ""}[panel] + ".png"
			f, _ := os.Create(name)
			_ = png.Encode(f, out)
			f.Close()
		}
	}
}

// captionButtons 在右上角画 Windows 11 的最小化、最大化、关闭按钮（10 DIP 的线条图标）。
func captionButtons(img *image.RGBA, right, h, w int, fg color.RGBA, scale float64) {
	s := int(10 * scale)
	cy := h / 2
	for i := 0; i < 3; i++ {
		cx := right - w*(3-i) + w/2
		switch i {
		case 0: // 最小化
			for x := cx - s/2; x <= cx+s/2; x++ {
				img.Set(x, cy, fg)
			}
		case 1: // 最大化
			for d := -s / 2; d <= s/2; d++ {
				img.Set(cx+d, cy-s/2, fg)
				img.Set(cx+d, cy+s/2, fg)
				img.Set(cx-s/2, cy+d, fg)
				img.Set(cx+s/2, cy+d, fg)
			}
		case 2: // 关闭
			for d := -s / 2; d <= s/2; d++ {
				img.Set(cx+d, cy+d, fg)
				img.Set(cx+d, cy-d, fg)
			}
		}
	}
}
