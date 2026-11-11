package main

import (
	"fmt"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/transfer"
	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/markdown"
)

// palette 返回当前外观的令牌。
func (app *App) palette() Palette {
	if app.dark {
		return Dark
	}
	return Light
}

// theme 把设计令牌映射到 MyGO 的主题；主题没有的令牌直接从 Palette 取。
func (app *App) theme() *ui.Theme {
	p := app.palette()
	var t *ui.Theme
	if app.dark {
		t = ui.DarkTheme()
	} else {
		t = ui.LightTheme()
	}
	t.Background, t.Surface, t.SurfaceHover = p.Bg, p.Surface, p.Hover
	t.Border, t.Text, t.TextMuted = p.Border, p.Text, p.Muted
	t.Accent, t.AccentText = p.Accent, p.OnAccent
	t.Danger, t.Warning, t.Success = p.Danger, p.Attention, p.Success
	t.Selection, t.Focus = p.Select, p.Accent
	t.Inverse, t.InverseText = p.TooltipBg, p.TooltipText
	t.Radius, t.FontSize, t.Spacing = RadiusControl, FontSizeUi, Space1
	t.FocusRingWidth, t.FocusRingOffset = 2, 1
	return t
}

func (app *App) mdTheme(t *ui.Theme) *markdown.Theme {
	p := app.palette()
	th := markdown.DefaultTheme(t)
	th.FontSize = FontSizeBody
	th.LineHeight = FontLineBody
	th.CodeFont = "SF Mono, Cascadia Code, Menlo, Consolas, monospace"
	th.CodeBlockBackground = p.CodeBg
	th.CodeBlockBorder = p.Border
	th.CodeBlockRadius = RadiusCard
	// 设计规范 §2.4：不用左侧色条。引用块改为整块浅底色。
	th.QuoteBarWidth = 0
	return th
}

func (app *App) view(c *ui.Context) {
	start := time.Now()
	defer func() { app.perf.record(time.Since(start)) }()
	t := app.theme()
	c.SetTheme(t)
	p := app.palette()
	if app.md.Theme == nil || app.md.Theme.CodeBlockBackground != p.CodeBg {
		app.md.Theme = app.mdTheme(t)
	}
	app.md.ShareSelection = true
	app.list.Key = func(i int) any { return app.items[i].ID }
	app.list.FollowEnd = true
	app.list.FollowAnimation = &ui.Smooth

	ui.Row(c).Fill().AlignItems(ui.Stretch).Children(func() {
		app.sidebar(c, p)
		ui.Column(c).Grow(1).AlignItems(ui.Stretch).Children(func() {
			app.header(c, p)
			app.streamView(c, p)
			ui.Column(c).AlignSelf(ui.Center).Width(SizeStreamMax).MaxWidth(10000).Padding(0, Space4, Space4, Space4).Gap(Space2).Children(func() {
				if !app.list.AtEnd() {
					if ui.Button(c, "有新内容 ↓").AlignSelf(ui.Center).Clicked() {
						app.list.ScrollToEnd()
					}
				}
				app.statusStrip(c, p)
				app.composer(c, p)
			})
		})
	})
}

func (pf *perf) record(d time.Duration) {
	pf.builds++
	pf.buildTotal += d
	pf.buildMax = max(pf.buildMax, d)
	if !pf.publishedAt.IsZero() {
		pf.latencies = append(pf.latencies, time.Since(pf.publishedAt))
		pf.publishedAt = time.Time{}
		pf.streamFrames++
	}
}

func (pf *perf) summary() (avg, mx time.Duration) {
	for _, l := range pf.latencies {
		avg += l
		mx = max(mx, l)
	}
	if n := len(pf.latencies); n > 0 {
		avg /= time.Duration(n)
	}
	return
}

func (app *App) sidebar(c *ui.Context, p Palette) {
	ui.Column(c).Width(SizeSidebar).Background(p.Surface).BorderWidth(0, 1, 0, 0).BorderColor(p.Border).Padding(Space3, Space2).Gap(Space1).Children(func() {
		ui.Text(c, "会话").FontSize(FontSizeSmall).TextColor(p.Muted).Padding(Space1, Space2)
		for i, name := range []string{"修复队列暂停逻辑", "设计令牌生成", "ACP 连接层", "工具面板装载"} {
			row := ui.ButtonBase(c.Key(name)).Justify(ui.Start).Padding(6, Space2).Radius(RadiusControl).Label(name)
			if i == 0 {
				row.Background(p.Select)
			} else if row.Hovered() {
				row.Background(p.Hover)
			}
			row.Children(func() {
				ui.Text(c, name).FontWeight(map[bool]int{true: 600, false: 400}[i == 0])
			})
		}
	})
}

func (app *App) header(c *ui.Context, p Palette) {
	ui.Row(c).Height(SizeHeader).AlignItems(ui.Center).Gap(Space2).Padding(0, Space4).BorderWidth(0, 0, 1, 0).BorderColor(p.Border).Children(func() {
		ui.Text(c, "修复队列暂停逻辑").FontSize(FontSizeTitle).FontWeight(600)
		ui.Text(c, "Claude Code · aibo-rev").TextColor(p.Muted)
		ui.Spacer(c)
		avg, mx := app.perf.summary()
		ui.Text(c, fmt.Sprintf("条目 %d · 构建平均 %.2fms 最大 %.1fms · 发布→构建 平均 %.1fms 最大 %.1fms",
			len(app.items)+len(app.older), ms(app.perf.buildTotal)/float64(max(app.perf.builds, 1)), ms(app.perf.buildMax), ms(avg), ms(mx))).
			FontSize(FontSizeSmall).TextColor(p.Muted)
		if ui.Button(c, "开始长回复").Disabled(app.running).Clicked() {
			app.startLong()
		}
		if ui.Button(c, "载入全部历史").Disabled(len(app.older) == 0).Clicked() {
			app.loadAll()
		}
	})
}

func ms(d time.Duration) float64 { return float64(d.Microseconds()) / 1000 }

func (app *App) streamView(c *ui.Context, p Palette) {
	if first, _ := app.list.Visible(); first < 5 && len(app.older) > 0 {
		app.loadOlder(80)
	}
	ui.Column(c).Grow(1).AlignItems(ui.Stretch).Selectable().Children(func() {
		ui.List(c, &app.list, len(app.items), func(i int) {
			app.item(c, p, app.items[i])
		}).Grow(1).Padding(Space4, Space4).Gap(Space4).Justify(ui.End).Label("会话内容")
	})
}

func (app *App) item(c *ui.Context, p Palette, it *Item) {
	ui.Column(c).AlignSelf(ui.Center).Width(SizeStreamMax).MaxWidth(10000).Children(func() {
		switch it.Kind {
		case KindUser:
			ui.Row(c).Children(func() {
				ui.Spacer(c)
				ui.Text(c, it.Text).Padding(Space2, Space3).Radius(RadiusBubble).Background(p.Bubble).MaxWidth(560)
			})
		case KindAgent:
			if it.Done {
				markdown.View(c, it.Text, &app.md)
			} else {
				markdown.Stream(c, it.Text, &app.md)
			}
		case KindTool:
			ui.Column(c).Border(1, p.Border).Radius(RadiusCard).Padding(Space2, Space3).Children(func() {
				ui.Collapsible(c, "工具 · "+it.Tool, &it.ToolOpen, func() {
					ui.Text(c, it.Text).Font("SF Mono, Menlo, monospace").FontSize(FontSizeSmall).Padding(Space2, 0, 0, 0)
				})
			})
		case KindApproval:
			app.approvalCard(c, p, it)
		case KindUnknown:
			ui.Column(c).Border(1, p.Unknown).BorderStyle(ui.BorderDashed).Background(p.UnknownBg).Radius(RadiusCard).Padding(Space3).Gap(Space2).Children(func() {
				ui.Text(c, "结果未知").FontWeight(600).TextColor(p.Unknown)
				ui.Text(c, it.Text)
				ui.Row(c).Children(func() {
					if ui.Button(c, "重新发送（可能重复执行）").Clicked() {
						app.deliver("重新发送上一条消息")
					}
				})
			})
		}
	})
}

func (app *App) approvalCard(c *ui.Context, p Palette, it *Item) {
	border, bg := p.Attention, p.AttentionBg
	if it.Approval != "pending" {
		border, bg = p.Border, ui.Transparent
	}
	ui.Column(c).Border(1, border).Background(bg).Radius(RadiusCard).Padding(Space3).Gap(Space2).Role(ui.RoleGroup).Label("审批请求").Children(func() {
		ui.Text(c, "需要你的批准").FontWeight(600).TextColor(map[bool]ui.Color{true: p.Attention, false: p.Muted}[it.Approval == "pending"])
		ui.Text(c, it.Text)
		ui.Row(c).Gap(Space2).Children(func() {
			switch it.Approval {
			case "pending":
				if ui.PrimaryButton(c, "允许一次").Clicked() {
					app.approve(it, true)
				}
				if ui.Button(c, "拒绝").Clicked() {
					app.approve(it, false)
				}
			case "allowed":
				ui.Text(c, "已允许一次").TextColor(p.Success)
			case "denied":
				ui.Text(c, "已拒绝").TextColor(p.Danger)
			}
		})
	})
}

func (app *App) statusStrip(c *ui.Context, p Palette) {
	ui.Row(c).Gap(Space2).AlignItems(ui.Center).Children(func() {
		chip := func(label string, fg, bg ui.Color) {
			ui.Text(c, label).FontSize(FontSizeSmall).TextColor(fg).Padding(2, Space2).Radius(RadiusPill).Background(bg)
		}
		if app.running {
			chip("运行中", p.Accent, p.Select)
		}
		if n := app.pendingApprovals(); n > 0 {
			chip(fmt.Sprintf("%d 个审批待处理", n), p.Attention, p.AttentionBg)
		}
		if len(app.queue) > 0 {
			b := ui.ButtonBase(c).Padding(2, Space2).Radius(RadiusPill).Background(p.Surface).Border(1, p.Border).Label(fmt.Sprintf("队列 %d", len(app.queue)))
			b.Children(func() { ui.Text(c, fmt.Sprintf("队列 %d", len(app.queue))).FontSize(FontSizeSmall) })
			if b.Clicked() {
				app.queueOpen = !app.queueOpen
			}
			ui.PopoverBase(c, b, &app.queueOpen, func(panel ui.Element) {
				panel.AttachTo(b, ui.AnchorTopLeft, ui.AnchorBottomLeft).Margin(0, 0, 6, 0).Padding(Space2).Gap(Space1).Width(360).
					Radius(RadiusPopover).Background(p.Raised).Border(1, p.Border).Shadow(0, 8, 28, 0, ui.RGBA(0, 0, 0, 0.12)).Label("队列")
				ui.Text(c, "待发送").FontSize(FontSizeSmall).TextColor(p.Muted)
				for i, q := range app.queue {
					ui.Row(c.Key(i)).Gap(Space2).AlignItems(ui.Center).Children(func() {
						ui.Text(c, q).Grow(1)
						idx := i
						ui.Button(c, "移除").Label("移除："+q).OnClick(func() {
							app.queue = append(app.queue[:idx:idx], app.queue[idx+1:]...)
						})
					})
				}
			})
		}
		if app.running {
			ui.Spacer(c)
			if ui.Button(c, "停止").Clicked() {
				app.stop()
			}
		}
	})
}

func (app *App) composer(c *ui.Context, p Palette) {
	box := ui.Column(c).Border(1, p.BorderStrong).Radius(RadiusCard).Background(p.Raised).Padding(Space2).Gap(Space2)
	if files := box.DroppedFiles(); files != nil {
		for _, f := range files {
			app.chips = append(app.chips, Chip{Kind: "drop", Label: f})
		}
	}
	if box.FileDragOver() {
		box.Border(2, p.Accent)
	}
	box.Children(func() {
		if len(app.chips) > 0 {
			ui.Row(c).Gap(Space1).Wrap().Children(func() {
				for i, ch := range app.chips {
					ui.Row(c.Key(fmt.Sprint(i, ch.Label))).Gap(Space1).AlignItems(ui.Center).Padding(2, 6).Radius(RadiusControl).Background(p.Surface).Border(1, p.Border).Children(func() {
						ui.Text(c, map[string]string{"file": "文件", "ref": "引用", "cmd": "命令", "image": "图片", "drop": "拖入"}[ch.Kind]+" · "+ch.Label).FontSize(FontSizeSmall)
						idx := i
						ui.ButtonBase(c).Label("移除 "+ch.Label).Padding(0, 4).Children(func() { ui.Text(c, "×").TextColor(p.Muted) }).OnClick(func() {
							app.chips = append(app.chips[:idx:idx], app.chips[idx+1:]...)
						})
					})
				}
			})
		}
		in := ui.TextAreaBase(c.Key("draft"), &app.draft).Lines(1, 8).Placeholder("输入消息，@ 引用文件，# 引用会话，/ 命令").Label("消息").FontSize(FontSizeBody)
		in.OnTextKey(func(k ui.TextKey) bool { return app.onKey(k) })
		in.OnPaste(func(text string) bool {
			if text != "" {
				return false
			}
			if png, err := mygo.Clipboard.ReadFormat(transfer.PNG); err == nil && len(png) > 0 {
				app.chips = append(app.chips, Chip{Kind: "image", Label: fmt.Sprintf("粘贴的图片 %d KB", len(png)/1024)})
				return true
			}
			return false
		})
		_, caret := in.TextSelection()
		app.detectTrigger(app.draft, caret)
		if app.pick.open {
			app.pickerPanel(c, p, box, in)
		}
		app.settingsBar(c, p)
	})
}

func (app *App) onKey(k ui.TextKey) bool {
	switch {
	case app.pick.open && k.Mods == 0 && (k.Key == ui.KeyDown || k.Key == ui.KeyUp):
		n := len(app.candidates())
		if k.Key == ui.KeyDown {
			app.pick.hi = (app.pick.hi + 1) % n
		} else {
			app.pick.hi = (app.pick.hi + n - 1) % n
		}
		return true
	case app.pick.open && k.Mods == 0 && (k.Key == ui.KeyEnter || k.Key == ui.KeyTab):
		text := k.Text()
		k.SetText(app.choose(text, len([]rune(text)), app.pick.hi))
		return true
	case app.pick.open && k.Key == ui.KeyEscape:
		app.pick.closed = k.Text()
		app.pick.open = false
		return true
	case k.Key == ui.KeyEnter && k.Mods == 0:
		if app.applying {
			return true // 设置应用中：不发送，草稿保留（UI-COMP）
		}
		app.send(k.Text())
		k.SetText("")
		return true
	}
	return false // Shift+Enter 换行
}

func (app *App) pickerPanel(c *ui.Context, p Palette, box, in ui.Element) {
	cs := app.candidates()
	ui.Overlay(c, func() {
		panel := ui.Column(c).AttachTo(box, ui.AnchorTopLeft, ui.AnchorBottomLeft).Margin(0, 0, 6, 0).Width(420).Padding(Space1).Gap(2).
			Radius(RadiusPopover).Background(p.Raised).Border(1, p.Border).Shadow(0, 8, 28, 0, ui.RGBA(0, 0, 0, 0.12)).
			Role(ui.RoleListBox).Label("候选")
		panel.Children(func() {
			for i, s := range cs {
				row := ui.ButtonBase(c.Key(s)).Justify(ui.Start).Padding(6, Space2).Radius(RadiusControl).Label(s).Role(ui.RoleOption)
				if i == app.pick.hi {
					row.Background(p.Select)
				}
				row.Children(func() { ui.Text(c, s) })
				idx := i
				row.OnClick(func() {
					_, caret := in.TextSelection()
					text := app.choose(app.draft, caret, idx)
					in.ReplaceText(0, 1<<30, text)
				})
			}
		})
		if panel.PressedOutside() {
			app.pick.closed = app.draft
			app.pick.open = false
		}
	})
}

func (app *App) settingsBar(c *ui.Context, p Palette) {
	ui.Row(c).Gap(Space2).AlignItems(ui.Center).Children(func() {
		prev := app.model
		ui.Select(c, &app.model, models).Label("模型")
		if app.model != prev {
			app.changeModel()
		}
		ui.Select(c, &app.effort, efforts).Label("推理强度")
		if app.applying {
			ui.Text(c, "正在应用设置…").FontSize(FontSizeSmall).TextColor(p.Muted)
		}
		ui.Spacer(c)
		empty := app.draft == "" && len(app.chips) == 0
		reason := ""
		switch {
		case app.applying:
			reason = "正在应用模型设置，完成后才能发送"
		case empty:
			reason = "输入内容后才能发送"
		}
		label := "发送"
		if app.running {
			label = "加入队列"
		}
		send := ui.PrimaryButton(c, label).AccessDisabled(reason != "")
		if reason != "" {
			// 禁用用形态和文字表达，不用半透明（UI-VIS）
			send.Background(p.Surface).TextColor(p.Muted).Border(1, p.Border).Tooltip(reason).Description(reason)
		}
		if send.Clicked() && reason == "" {
			app.send(app.draft)
			app.draft = ""
		}
	})
}
