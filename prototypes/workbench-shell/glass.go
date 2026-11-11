package main

// 一次性实验：Liquid Glass 外观。玻璃只用在外框层（侧栏材质、顶部滚动边缘、浮动输入区），
// 内容层（消息、工具卡、审批卡）保持不透明。

import (
	"github.com/egoist/mygo"
	"github.com/egoist/mygo/plugins/glass"
	"github.com/egoist/mygo/ui"
)

func (app *App) setGlass(on bool) {
	app.sh.glass = on
	if app.win != nil {
		if on {
			app.win.SetVibrancy(mygo.VibrancySidebar)
		} else {
			app.win.SetVibrancy(mygo.VibrancyNone)
		}
	}
	app.event("外观：glass=%v", on)
}

func (app *App) headerH(c *ui.Context) float32 { return max(c.TitleBar().Height, SizeHeader) }

// streamPad：玻璃外观下会话流延伸到标题栏与输入区下方，用内边距留出首尾。
func (app *App) streamPad() []float32 {
	if !app.sh.glass {
		return []float32{Space4, Space4}
	}
	return []float32{SizeHeader + Space4, Space4, max(app.sh.composerH, 120) + Space4, Space4}
}

func (app *App) scrollInsets() []float32 {
	if !app.sh.glass {
		return []float32{0}
	}
	return []float32{SizeHeader, 0, max(app.sh.composerH, 120), 0}
}

func (app *App) composerSurface(box ui.Element, p Palette) {
	if app.sh.glass {
		return // 玻璃面板在外层（glassMain），输入框本身透明
	}
	box.Border(1, p.BorderStrong).Radius(RadiusCard).Background(p.Raised)
}

// glassMain：会话流铺满，标题栏与输入区浮在其上。
func (app *App) glassMain(c *ui.Context, p Palette, box ui.Element) {
	hh := app.headerH(c)
	box.Background(p.Bg).Children(func() {
		ui.Box(c).Grow(1).AlignSelf(ui.Stretch).Children(func() {
			app.guard(c, "会话内容", ui.Column(c).Fill().AlignItems(ui.Stretch), func() { app.streamView(c, p) })
			// 顶部：内容滚入标题栏下方时渐隐（macOS 26 的 soft scroll edge）
			ui.Box(c).Absolute().Top(0).Left(0).Right(0).Height(hh + 10).PassThrough().Material(glass.ScrollEdge{Hard: true, Background: p.Bg})
			ui.Column(c).Absolute().Top(0).Left(0).Right(0).AlignItems(ui.Stretch).Children(func() { app.header(c, p) })
			// 底部：浮动的玻璃输入区，下方内容渐隐
			ui.Box(c).Absolute().Bottom(0).Left(0).Right(0).Height(max(app.sh.composerH, 120) * 0.6).PassThrough().Material(glass.ScrollEdge{Bottom: true, Background: p.Bg})
			ui.Column(c).Absolute().Bottom(0).Left(0).Right(0).AlignItems(ui.Center).PassThrough().Children(func() {
				area := app.guard(c, "输入区", ui.Column(c).Width(SizeStreamMax).MaxWidthPercent(100).Padding(0, Space4, Space4, Space4).Gap(Space2), func() {
					if !app.list.AtEnd() {
						ui.Row(c).AlignSelf(ui.Center).Padding(4, 12).Radius(16).Material(glass.Glass{Interactive: true}).Label("有新内容").Children(func() {
							ui.Text(c, "有新内容 ↓").FontSize(FontSizeSmall)
						}).OnClick(func() { app.list.ScrollToEnd() })
					}
					// 状态条与输入框在同一块玻璃上
					ui.Column(c).AlignItems(ui.Stretch).Radius(RadiusGlass).Padding(Space2, Space3).Gap(Space1).Material(glass.Glass{}).Children(func() {
						app.statusStrip(c, p)
						app.composer(c, p)
					})
				})
				if b := area.Bounds(); b.H > 0 {
					app.sh.composerH = b.H
				}
			})
		})
	})
}
