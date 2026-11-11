package main

// 一次性实验：Liquid Glass 组件的三套方案，在玻璃外观下用底部切换条（或 ←/→）切换。
//
//	A 玻璃胶囊：所有外框控件都是可按压的玻璃胶囊，相邻按钮合成一组（macOS 26 工具栏）。
//	B 安静无框：控件平时无底色，只在悬停时出现浅底；主操作是实心圆形按钮；状态是圆点加文字。
//	C 浮动工具岛：标题行操作集中在一块玻璃工具岛里，状态以实心徽记贴在输入区上缘，模型与推理强度合成一个分段控件。
//
// 内容层（审批卡、结果未知卡）里的按钮在三套方案中都不透明（UI-VIS-09），只改形状与层级。

import (
	"fmt"

	"github.com/egoist/mygo/plugins/glass"
	"github.com/egoist/mygo/ui"
)

var variantNames = []string{"A 玻璃胶囊", "B 安静无框", "C 浮动工具岛"}

func svg(shapes string) *ui.SVG {
	return ui.MustParseSVG([]byte(`<svg xmlns="http://www.w3.org/2000/svg" viewBox="0 0 16 16" fill="none" stroke="currentColor" stroke-width="1.4" stroke-linecap="round" stroke-linejoin="round">` + shapes + `</svg>`))
}

var (
	icPlay    = svg(`<path d="M5 3.5v9l7-4.5z"/>`)
	icPanel   = svg(`<rect x="1.5" y="2.5" width="13" height="11" rx="2"/><path d="M10 2.5v11"/>`)
	icDots    = svg(`<circle cx="3.5" cy="8" r=".8"/><circle cx="8" cy="8" r=".8"/><circle cx="12.5" cy="8" r=".8"/>`)
	icUp      = svg(`<path d="M8 13V3M4 7l4-4 4 4"/>`)
	icStop    = svg(`<rect x="4" y="4" width="8" height="8" rx="1.5"/>`)
	icChevron = svg(`<path d="M4.5 6.5 8 10l3.5-3.5"/>`)
	icDown    = svg(`<path d="M8 3v10M4 9l4 4 4-4"/>`)
	icQueue   = svg(`<path d="M3 4.5h10M3 8h10M3 11.5h6"/>`)
	icReload  = svg(`<path d="M13 8a5 5 0 1 1-1.5-3.5M13 2.5v3h-3"/>`)
	icClose   = svg(`<path d="M4.5 4.5l7 7M11.5 4.5l-7 7"/>`)
)

type kind int

const (
	kNormal kind = iota
	kPrimary
)

type tone int

const (
	tNeutral tone = iota
	tRunning
	tAttention
	tUnknown
)

func (app *App) variant() int { return app.sh.variant }

func (app *App) toneColor(p Palette, t tone) ui.Color {
	switch t {
	case tRunning:
		return p.Accent
	case tAttention:
		return p.Attention
	case tUnknown:
		return p.Unknown
	}
	return p.Muted
}

// ---- 外框层（玻璃之上）的按钮 ----

// kbtn：外框层按钮。icon 可为空；iconOnly 时只显示图标，文字作为无障碍名称与提示。
func (app *App) kbtn(c *ui.Context, p Palette, label string, icon *ui.SVG, k kind, iconOnly bool) ui.Element {
	b := ui.ButtonBase(c.Key("kbtn:" + label)).Label(label).Gap(6).AlignItems(ui.Center)
	if iconOnly {
		b.Tooltip(label)
	}
	fg := p.Text
	switch app.variant() {
	case 0: // 玻璃胶囊
		if iconOnly {
			b.Size(SizeControl+4, SizeControl+4).Center().Radius(RadiusPill)
		} else {
			b.Height(SizeControl+4).Padding(0, 14).Radius(RadiusPill)
		}
		g := glass.Glass{Interactive: true}
		if k == kPrimary {
			g.Tint = p.Accent
			fg = p.OnAccent
		}
		b.Material(g)
	case 1: // 安静无框
		if k == kPrimary {
			b.Size(SizeControl+4, SizeControl+4).Center().Radius(RadiusPill).Background(p.Accent)
			fg = p.OnAccent
			iconOnly = icon != nil
		} else {
			if iconOnly {
				b.Size(SizeControl, SizeControl).Center().Radius(RadiusCard)
			} else {
				b.Height(SizeControl).Padding(0, Space2).Radius(RadiusCard)
			}
			fg = p.Muted
			if b.Hovered() || b.Pressed() {
				b.Background(p.Hover)
				fg = p.Text
			}
		}
	case 2: // 浮动工具岛：岛内按钮无底色，悬停时浅底；主操作为着色玻璃圆
		if k == kPrimary {
			b.Size(SizeControl+6, SizeControl+6).Center().Radius(RadiusPill).Material(glass.Glass{Tint: p.Accent, Interactive: true})
			fg = p.OnAccent
			iconOnly = icon != nil
		} else {
			b.Height(SizeControl).Padding(0, Space2+2).Radius(RadiusPill)
			if iconOnly {
				b.Width(SizeControl).Padding(0).Center()
			}
			if b.Hovered() || b.Pressed() {
				b.Background(p.Hover)
			}
		}
	}
	b.TextColor(fg).Children(func() {
		if icon != nil {
			ui.Icon(c, icon).Size(14, 14)
		}
		if !iconOnly || icon == nil {
			ui.Text(c, label).FontWeight(map[bool]int{true: 600, false: 500}[k == kPrimary])
		}
	})
	return b
}

// kchip：状态标签/徽记。可点击时返回按钮。
func (app *App) kchip(c *ui.Context, p Palette, label string, t tone, clickable bool) ui.Element {
	col := app.toneColor(p, t)
	var e ui.Element
	if clickable {
		e = ui.ButtonBase(c.Key("chip:" + label)).Label(label)
	} else {
		e = ui.Row(c.Key("chip:" + label)).Label(label).Role(ui.RoleStatus)
	}
	e.Gap(5).AlignItems(ui.Center)
	switch app.variant() {
	case 0: // 着色玻璃小胶囊
		e.Padding(3, 10).Radius(RadiusPill).Material(glass.Glass{Tint: col.Alpha(0.16), Interactive: clickable})
		e.Children(func() { ui.Text(c, label).FontSize(FontSizeSmall).FontWeight(500).TextColor(col) })
	case 1: // 圆点 + 文字，无底色
		e.Padding(2, 4).Radius(RadiusControl)
		if clickable && e.Hovered() {
			e.Background(p.Hover)
		}
		e.Children(func() {
			ui.Box(c).Size(7, 7).Radius(4).Background(col).AccessHidden(true)
			ui.Text(c, label).FontSize(FontSizeSmall).TextColor(p.Muted)
		})
	case 2: // 实心徽记
		bg := col
		fg := p.OnAccent
		if t == tNeutral {
			bg, fg = p.Text, p.Bg
		}
		e.Padding(2, 9).Radius(RadiusPill).Background(bg).Shadow(0, 2, 8, 0, popoverShadow)
		e.Children(func() { ui.Text(c, label).FontSize(FontSizeSmall).FontWeight(600).TextColor(fg) })
	}
	return e
}

// kselect：模型、推理强度等选择器。
func (app *App) kselect(c *ui.Context, p Palette, value *string, opts []string, label string) bool {
	prev := *value
	sel := ui.SelectBase(c, value)
	tr := sel.Trigger.Label(label).Gap(4).AlignItems(ui.Center)
	switch app.variant() {
	case 0:
		tr.Height(SizeControl).Padding(0, 10, 0, 12).Radius(RadiusPill).Material(glass.Glass{Interactive: true})
	case 1:
		tr.Height(SizeControl).Padding(0, 6).Radius(RadiusCard).TextColor(p.Muted)
		if tr.Hovered() || sel.Open() {
			tr.Background(p.Hover).TextColor(p.Text)
		}
	case 2:
		tr.Height(SizeControl).Padding(0, 10).Radius(RadiusPill)
		if tr.Hovered() || sel.Open() {
			tr.Background(p.Hover)
		}
	}
	tr.Children(func() {
		ui.Text(c, *value).FontSize(FontSizeSmall + 1)
		ui.Icon(c, icChevron).Size(12, 12)
	})
	sel.Popup(func(panel ui.Element) {
		panel.Margin(6, 0, 0, 0)
		app.glassPanel(panel)
		for _, o := range opts {
			it := sel.Item(o).Height(SizeControl).Padding(0, 10).Radius(10).AlignItems(ui.Center)
			if it.Highlighted() {
				it.Background(p.Accent).TextColor(p.OnAccent)
			}
			it.Children(func() { ui.Text(c, o) })
		}
	})
	return *value != prev
}

// kattach：输入框上方的附件/引用标签。
func (app *App) kattach(c *ui.Context, p Palette, kind, label string, remove func()) {
	row := ui.Row(c).Gap(Space1).AlignItems(ui.Center).Label(kind + "：" + label)
	switch app.variant() {
	case 0: // 玻璃小胶囊，类别用弱化文字
		row.Height(26).Padding(0, 4, 0, 10).Radius(RadiusPill).Material(glass.Glass{Interactive: true})
	case 1: // 浅底小圆角
		row.Height(24).Padding(0, 2, 0, 8).Radius(RadiusControl).Background(p.Hover)
	case 2: // 描边胶囊
		row.Height(26).Padding(0, 4, 0, 10).Radius(RadiusPill).Border(1, p.BorderStrong)
	}
	row.Children(func() {
		ui.Text(c, kind).FontSize(FontSizeSmall).TextColor(p.Muted)
		ui.Text(c, label).FontSize(FontSizeSmall).FontWeight(500)
		x := ui.ButtonBase(c).Label("移除 "+label).Size(18, 18).Center().Radius(9).TextColor(p.Muted)
		if x.Hovered() {
			x.Background(p.Hover).TextColor(p.Text)
		}
		x.Children(func() { ui.Icon(c, icClose).Size(10, 10) })
		if x.Clicked() {
			remove()
		}
	})
}

// ---- 内容层（不透明卡片里）的按钮 ----

func (app *App) cbtn(c *ui.Context, p Palette, label string, k kind) ui.Element {
	b := ui.ButtonBase(c.Key("cbtn:" + label)).Label(label).Height(SizeControl).AlignItems(ui.Center)
	fg := p.Text
	switch app.variant() {
	case 0: // 胶囊：主操作实心，次操作描边
		b.Padding(0, 14).Radius(RadiusPill)
		if k == kPrimary {
			b.Background(p.Accent)
			fg = p.OnAccent
		} else {
			b.Background(p.Raised).Border(1, p.BorderStrong)
		}
	case 1: // 安静：主操作实心小圆角，次操作文字按钮
		b.Padding(0, Space3).Radius(RadiusCard)
		if k == kPrimary {
			b.Background(p.Accent)
			fg = p.OnAccent
		} else {
			fg = p.Accent
			if b.Hovered() {
				b.Background(p.Hover)
			}
		}
	case 2: // 工具岛：主操作实心，次操作描边，均为 8px 圆角
		b.Padding(0, Space3).Radius(RadiusCard)
		if k == kPrimary {
			b.Background(p.Accent)
			fg = p.OnAccent
		} else {
			b.Border(1, p.Accent)
			fg = p.Accent
		}
	}
	if b.Hovered() && k == kPrimary {
		b.Background(p.Accent.Mix(p.Text, 0.12))
	}
	b.TextColor(fg).Children(func() { ui.Text(c, label).FontWeight(map[bool]int{true: 600, false: 500}[k == kPrimary]) })
	return b
}

// ---- 各区域在三套方案下的布局 ----

// kHeader：会话区标题行。
func (app *App) kHeader(c *ui.Context, p Palette, hdr ui.Element) {
	hdr.Children(func() {
		ui.Column(c).Gap(0).Shrink(1).Children(func() {
			ui.Text(c, app.sh.sessions[app.sh.cur].name).FontSize(FontSizeTitle).FontWeight(600).MaxLines(1)
			ui.Text(c, "Claude Code · aibo-rev").FontSize(FontSizeSmall).TextColor(p.Muted).MaxLines(1)
		})
		ui.Spacer(c)
		switch app.variant() {
		case 0: // 合成一组的玻璃胶囊 + 单独的“更多”
			g := glass.NewGroup(c)
			g.Radius = RadiusPill
			g.Gap(2).Children(func() {
				for _, a := range []struct {
					label string
					icon  *ui.SVG
					fn    func()
					off   bool
				}{
					{"开始长回复", icPlay, app.startLong, app.running},
					{map[bool]string{true: "收起工具面板", false: "展开工具面板"}[app.sh.panelOpen], icPanel, func() { app.sh.panelOpen = !app.sh.panelOpen }, false},
				} {
					b := ui.ButtonBase(c.Key(a.label)).Label(a.label).Tooltip(a.label).Size(SizeControl+8, SizeControl+4).Center().Radius(RadiusPill).TextColor(p.Text).AccessDisabled(a.off)
					if b.Hovered() && !a.off {
						b.Background(p.Hover)
					}
					b.Children(func() { ui.Icon(c, a.icon).Size(15, 15) })
					g.Pane(b)
					if b.Clicked() && !a.off {
						a.fn()
					}
				}
			})
			app.glassMenu(c, p, "更多", &app.sh.menuMore, true, app.moreItems())
		case 1: // 只有图标，平时无底色
			if app.kbtn(c, p, "开始长回复", icPlay, kNormal, true).AccessDisabled(app.running).Clicked() && !app.running {
				app.startLong()
			}
			if app.kbtn(c, p, "切换工具面板", icPanel, kNormal, true).Clicked() {
				app.sh.panelOpen = !app.sh.panelOpen
			}
			app.glassMenu(c, p, "更多", &app.sh.menuMore, true, app.moreItems())
		case 2: // 一块玻璃工具岛，内含带文字的按钮与分隔线
			ui.Row(c).Height(SizeControl+8).Padding(0, 4).Gap(2).AlignItems(ui.Center).Radius(RadiusPill).Material(glass.Glass{}).Children(func() {
				if app.kbtn(c, p, "开始长回复", icPlay, kNormal, false).AccessDisabled(app.running).Clicked() && !app.running {
					app.startLong()
				}
				ui.Box(c).Size(1, 16).Background(p.Border).AccessHidden(true)
				if app.kbtn(c, p, "工具面板", icPanel, kNormal, false).Clicked() {
					app.sh.panelOpen = !app.sh.panelOpen
				}
				ui.Box(c).Size(1, 16).Background(p.Border).AccessHidden(true)
				app.glassMenu(c, p, "更多", &app.sh.menuMore, true, app.moreItems())
			})
		}
	})
}

func (app *App) moreMenu(c *ui.Context, b ui.Element) {
	b.PopoverMenuButton(func(m *ui.Menu) {
		if m.Item("标准外观").Chosen() {
			app.setGlass(false)
		}
		m.Separator()
		for _, name := range []string{"侧栏", "会话内容", "工具卡", "输入区", "候选弹层", "工具面板"} {
			if m.Item("故障注入：" + name).Chosen() {
				app.sh.inject[name] = true
			}
		}
	})
}

// kStatus：状态条。C 方案的徽记贴在输入区玻璃的上缘，由 kComposerPanel 放置。
func (app *App) kStatus(c *ui.Context, p Palette) {
	ui.Row(c).Grow(1).Gap(Space2).AlignItems(ui.Center).Children(func() {
		if app.running {
			app.kchip(c, p, "运行中", tRunning, false)
		}
		if n := app.pendingApprovals(); n > 0 {
			app.kchip(c, p, fmt.Sprintf("%d 个审批待处理", n), tAttention, false)
		}
		if len(app.queue) > 0 {
			q := app.kchip(c, p, fmt.Sprintf("队列 %d", len(app.queue)), tNeutral, true)
			if q.Clicked() {
				app.queueOpen = !app.queueOpen
			}
			app.queuePopover(c, p, q)
		}
		if app.running && app.variant() != 2 {
			ui.Spacer(c)
			if app.kbtn(c, p, "停止", icStop, kNormal, false).Clicked() {
				app.stop()
			}
		}
	})
}

func (app *App) queuePopover(c *ui.Context, p Palette, anchor ui.Element) {
	ui.PopoverBase(c, anchor, &app.queueOpen, func(panel ui.Element) {
		panel.AttachTo(anchor, ui.AnchorTopLeft, ui.AnchorBottomLeft).Margin(0, 0, 8, 0).Padding(Space2).Gap(Space1).Width(360).
			Radius(18).Material(glass.Glass{}).Shadow(0, 10, 30, 0, popoverShadow).Label("队列")
		ui.Text(c, "待发送").FontSize(FontSizeSmall).TextColor(p.Muted).Padding(0, Space1)
		for i, q := range app.queue {
			ui.Row(c.Key(i)).Gap(Space2).AlignItems(ui.Center).Padding(0, Space1).Children(func() {
				ui.Text(c, q).Grow(1)
				idx := i
				if app.kbtn(c, p, "移除："+q, icClose, kNormal, true).Clicked() {
					app.queue = append(app.queue[:idx:idx], app.queue[idx+1:]...)
				}
			})
		}
	})
}

// kSettings：输入框下方的设置与发送。
func (app *App) kSettings(c *ui.Context, p Palette) {
	ui.Row(c).Gap(Space1).AlignItems(ui.Center).Children(func() {
		if app.variant() == 2 {
			// 模型与推理强度合成一个分段控件
			ui.Row(c).Gap(0).AlignItems(ui.Center).Radius(RadiusPill).Border(1, p.Border).Children(func() {
				if app.kselect(c, p, &app.model, models, "模型") {
					app.changeModel()
				}
				ui.Box(c).Size(1, 14).Background(p.Border).AccessHidden(true)
				app.kselect(c, p, &app.effort, efforts, "推理强度")
			})
		} else {
			if app.kselect(c, p, &app.model, models, "模型") {
				app.changeModel()
			}
			app.kselect(c, p, &app.effort, efforts, "推理强度")
		}
		if app.applying {
			ui.Text(c, "正在应用设置…").FontSize(FontSizeSmall).TextColor(p.Muted)
		}
		ui.Spacer(c)
		if app.variant() == 2 && app.running {
			if app.kbtn(c, p, "停止", icStop, kNormal, false).Clicked() {
				app.stop()
			}
		}
		empty := app.draft == "" && len(app.chips) == 0
		reason := ""
		switch {
		case app.applying:
			reason = "正在应用模型设置，完成后才能发送"
		case empty:
			reason = "输入内容后才能发送"
		}
		label := "发送"
		icon := icUp
		if app.running {
			label, icon = "加入队列", icQueue
		}
		send := app.kbtn(c, p, label, icon, kPrimary, false).AccessDisabled(reason != "")
		if reason != "" {
			send.Tooltip(reason).Description(reason)
			// 禁用：去掉强调色，保持形状（UI-VIS：不用半透明表达禁用）
			switch app.variant() {
			case 0, 2:
				send.Material(glass.Glass{}).TextColor(p.Muted)
			case 1:
				send.Background(p.Hover).TextColor(p.Muted)
			}
		}
		if send.Clicked() && reason == "" {
			app.send(app.draft)
			app.draft = ""
		}
	})
}

// kComposerPanel：浮动输入区的玻璃面板。
func (app *App) kComposerPanel(c *ui.Context, p Palette) {
	if !app.list.AtEnd() {
		nb := app.kbtn(c, p, "有新内容", icDown, kNormal, app.variant() == 1).AlignSelf(ui.Center)
		if nb.Clicked() {
			app.list.ScrollToEnd()
		}
	}
	if app.variant() == 2 {
		// 徽记压在玻璃面板上缘
		ui.Box(c).AlignSelf(ui.Stretch).Children(func() {
			ui.Column(c).AlignItems(ui.Stretch).Radius(RadiusGlass).Padding(Space3+2, Space3, Space2, Space3).Gap(Space1).Material(glass.Glass{}).Children(func() {
				app.composer(c, p)
			})
			ui.Row(c).Absolute().Top(-10).Left(Space3).Children(func() { app.kStatus(c, p) })
		})
		return
	}
	if app.variant() == 0 {
		// 状态标签作为独立的玻璃小胶囊，悬浮在输入面板之外
		ui.Row(c).AlignSelf(ui.Stretch).Padding(0, Space1).Children(func() { app.kStatus(c, p) })
		ui.Column(c).AlignItems(ui.Stretch).Radius(RadiusGlass).Padding(Space2, Space3).Gap(Space1).Material(glass.Glass{}).Children(func() {
			app.composer(c, p)
		})
		return
	}
	ui.Column(c).AlignItems(ui.Stretch).Radius(RadiusGlass).Padding(Space2, Space3).Gap(Space1).Material(glass.Glass{}).Children(func() {
		app.kStatus(c, p)
		app.composer(c, p)
	})
}

// kSessionRow：侧栏会话行（侧栏为窗口材质）。
func (app *App) kSessionRow(c *ui.Context, p Palette, row ui.Element, selected bool) {
	switch app.variant() {
	case 0:
		row.Radius(RadiusPill).Padding(7, 12)
		if selected {
			row.Material(glass.Glass{Interactive: true})
		} else if row.Hovered() {
			row.Background(p.Hover)
		}
	case 1:
		row.Radius(RadiusCard)
		if selected {
			row.Background(p.Hover)
		} else if row.Hovered() {
			row.Background(p.Hover.Alpha(0.5))
		}
	case 2:
		row.Radius(10).Padding(7, 10)
		if selected {
			row.Material(glass.Glass{Tint: p.Accent.Alpha(0.22), Interactive: true})
		} else if row.Hovered() {
			row.Background(p.Hover)
		}
	}
}

// variantSwitcher：原型用的方案切换条，固定在窗口底部左侧（与设计无关）。
func (app *App) variantSwitcher(c *ui.Context) {
	if !app.sh.glass {
		return
	}
	if c.Shortcut(0, ui.KeyRight) {
		app.sh.variant = (app.sh.variant + 1) % len(variantNames)
	}
	if c.Shortcut(0, ui.KeyLeft) {
		app.sh.variant = (app.sh.variant + len(variantNames) - 1) % len(variantNames)
	}
	dark := ui.Hex("#111114") // ui-check-allow: color 原型切换条，不属于设计
	ui.Row(c).Absolute().Left(Space3).Bottom(Space3).Height(36).Padding(0, 4).Gap(4).AlignItems(ui.Center).Radius(18).
		Background(dark).Shadow(0, 6, 20, 0, popoverShadow).Label("原型方案切换").Children(func() {
		arrow := func(s string, d int) {
			b := ui.ButtonBase(c.Key("sw"+s)).Label(map[int]string{-1: "上一个方案", 1: "下一个方案"}[d]).Size(28, 28).Center().Radius(14).TextColor(ui.Hex("#FFFFFF")) // ui-check-allow: color 原型切换条
			b.Children(func() { ui.Text(c, s) })
			if b.Clicked() {
				app.sh.variant = (app.sh.variant + len(variantNames) + d) % len(variantNames)
			}
		}
		arrow("‹", -1)
		ui.Text(c, variantNames[app.sh.variant]).FontSize(FontSizeSmall).FontWeight(600).TextColor(ui.Hex("#FFFFFF")).Width(96).TextAlign(ui.Center) // ui-check-allow: color 原型切换条
		arrow("›", 1)
	})
}

// ---- 玻璃菜单（与模型/推理强度下拉一致的外观） ----

var (
	icAppearance = svg(`<circle cx="8" cy="8" r="5.5"/><path d="M8 2.5v11" />`)
	icBug        = svg(`<rect x="5" y="5" width="6" height="8" rx="3"/><path d="M8 5V3M5 8H2.5M13.5 8H11M5 11l-2 1.5M11 11l2 1.5M5.5 5.5 4 4M10.5 5.5 12 4"/>`)
	icFreeze     = svg(`<path d="M8 2v12M3 5l10 6M13 5 3 11"/>`)
	icKill       = svg(`<circle cx="8" cy="8" r="5.5"/><path d="M5.5 5.5l5 5"/>`)
	icCollapse   = svg(`<rect x="1.5" y="2.5" width="13" height="11" rx="2"/><path d="M10 2.5v11M7 6.5 5.5 8 7 9.5"/>`)
	icRight      = svg(`<path d="M6.5 4.5 10 8l-3.5 3.5"/>`)
)

type gitem struct {
	label   string
	icon    *ui.SVG
	hint    string // 右侧的弱化说明（快捷键等）
	danger  bool
	off     bool
	fn      func()
	sub     []gitem
	heading string // 非空时是小标题，不可选
	sep     bool   // 分隔线
}

func (app *App) glassPanel(panel ui.Element) {
	panel.Padding(Space1).Gap(2).MinWidth(200).Radius(16).Material(glass.Glass{}).Shadow(0, 10, 30, 0, popoverShadow)
}

// glassMenu：图标按钮触发的玻璃菜单。end 为 true 时菜单与按钮右缘对齐（标题行最右侧的按钮）。
func (app *App) glassMenu(c *ui.Context, p Palette, label string, open *bool, end bool, items []gitem) {
	m := ui.MenuBase(c, open)
	tr := m.Trigger.Label(label).Tooltip(label).Size(SizeControl+4, SizeControl+4).Center().Radius(RadiusPill).TextColor(p.Text)
	switch app.variant() {
	case 0:
		tr.Material(glass.Glass{Interactive: true})
	default:
		if tr.Hovered() || m.Open() {
			tr.Background(p.Hover)
		}
	}
	tr.Children(func() { ui.Icon(c, icDots).Size(14, 14) })
	m.Popup(func(panel ui.Element) {
		if end {
			panel.AttachTo(m.Trigger, ui.AnchorBottomEnd, ui.AnchorTopEnd)
		}
		panel.Margin(6, 0, 0, 0)
		app.glassPanel(panel)
		app.glassItems(c, p, m, items)
	})
}

func (app *App) glassItems(c *ui.Context, p Palette, m ui.MenuParts, items []gitem) {
	for _, it := range items {
		switch {
		case it.sep:
			ui.Box(c).Height(1).Margin(Space1, Space2).Background(p.Border).AccessHidden(true)
			continue
		case it.heading != "":
			ui.Text(c, it.heading).FontSize(FontSizeSmall).TextColor(p.Muted).Padding(Space1, 10, 2, 10)
			continue
		}
		var row ui.Element
		if it.sub != nil {
			row = m.SubmenuItem()
		} else {
			row = m.Item()
		}
		row.Label(it.label).Height(SizeControl).Padding(0, 10).Gap(Space2).Radius(10).AlignItems(ui.Center).Disabled(it.off)
		fg, iconFg := p.Text, p.Muted
		if it.danger {
			fg, iconFg = p.Danger, p.Danger
		}
		if it.off {
			fg, iconFg = p.Muted, p.Muted
		} else if row.Highlighted() {
			bg := p.Accent
			if it.danger {
				bg = p.Danger
			}
			row.Background(bg)
			fg, iconFg = p.OnAccent, p.OnAccent
		}
		row.TextColor(fg).Children(func() {
			ui.Box(c).Size(16, 16).Center().TextColor(iconFg).Children(func() {
				if it.icon != nil {
					ui.Icon(c, it.icon).Size(15, 15)
				}
			})
			ui.Text(c, it.label).Grow(1)
			if it.hint != "" {
				ui.Text(c, it.hint).FontSize(FontSizeSmall).TextColor(map[bool]ui.Color{true: fg, false: p.Muted}[row.Highlighted()])
			}
			if it.sub != nil {
				ui.Icon(c, icRight).Size(12, 12)
			}
		})
		if it.sub != nil {
			sub := it.sub
			m.Submenu(row, func(panel ui.Element) {
				app.glassPanel(panel)
				app.glassItems(c, p, m, sub)
			})
		} else if row.Clicked() && it.fn != nil {
			it.fn()
		}
	}
}

func (app *App) moreItems() []gitem {
	var faults []gitem
	for _, n := range []string{"侧栏", "会话内容", "工具卡", "输入区", "候选弹层", "工具面板"} {
		name := n
		faults = append(faults, gitem{label: "让「" + name + "」构建时 panic", danger: true, fn: func() { app.sh.inject[name] = true }})
	}
	return []gitem{
		{heading: "外观"},
		{label: "标准外观", icon: icAppearance, fn: func() { app.setGlass(false) }},
		{label: map[bool]string{true: "收起工具面板", false: "展开工具面板"}[app.sh.panelOpen], icon: icPanel, hint: "⌘\\", fn: func() { app.sh.panelOpen = !app.sh.panelOpen }},
		{sep: true},
		{heading: "原型"},
		{label: "故障注入", icon: icBug, sub: faults},
	}
}

func (app *App) panelItems() []gitem {
	return []gitem{
		{label: "重新载入", icon: icReload, fn: func() { go app.reloadPanel() }},
		{label: "收起", icon: icCollapse, hint: "⌘\\", fn: func() { app.sh.panelOpen = false }},
		{sep: true},
		{heading: "故障演练"},
		{label: "让面板页面卡死", icon: icFreeze, danger: true, fn: func() { go app.evalPanel("setTimeout(() => { for (;;) {} }, 0); return 1") }},
		{label: "结束面板渲染进程", icon: icKill, danger: true, fn: func() { app.killPanel() }},
	}
}
