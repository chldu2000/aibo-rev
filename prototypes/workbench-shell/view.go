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

func ms(d time.Duration) float64 { return float64(d.Microseconds()) / 1000 }

func (app *App) streamView(c *ui.Context, p Palette) {
	if first, _ := app.list.Visible(); first < 5 && len(app.older) > 0 {
		app.loadOlder(80)
	}
	ui.Column(c).Grow(1).AlignItems(ui.Stretch).Selectable().Children(func() {
		ui.List(c, &app.list, len(app.items), func(i int) {
			app.guardedItem(c, p, app.items[i])
		}).Grow(1).Padding(app.streamPad()...).ScrollbarInsets(app.scrollInsets()...).Gap(Space4).Justify(ui.End).Label("会话内容")
	})
}

func (app *App) item(c *ui.Context, p Palette, it *Item) {
	ui.Column(c).AlignSelf(ui.Center).Width(SizeStreamMax).MaxWidthPercent(100).Children(func() {
		switch it.Kind {
		case KindUser:
			ui.Row(c).Role(ui.RoleGroup).Label("用户消息").Children(func() {
				ui.Spacer(c)
				ui.Text(c, it.Text).Padding(Space2, Space3).Radius(RadiusBubble).Background(p.Bubble).MaxWidth(560)
			})
		case KindAgent:
			ui.Column(c).AlignItems(ui.Stretch).Role(ui.RoleGroup).Label("Agent 回复").Children(func() {
				if it.Done {
					markdown.View(c, it.Text, &app.md)
				} else {
					markdown.Stream(c, it.Text, &app.md)
				}
			})
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
					if app.btn(c, p, "重新发送（可能重复执行）", kNormal).Clicked() {
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
				if app.btn(c, p, "允许一次", kPrimary).Clicked() {
					app.approve(it, true)
				}
				if app.btn(c, p, "拒绝", kNormal).Clicked() {
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
					Radius(RadiusPopover).Background(p.Raised).Border(1, p.Border).Shadow(0, 8, 28, 0, popoverShadow).Label("队列")
				ui.Text(c, "待发送").FontSize(FontSizeSmall).TextColor(p.Muted)
				for i, q := range app.queue {
					ui.Row(c.Key(i)).Gap(Space2).AlignItems(ui.Center).Children(func() {
						ui.Text(c, q).Grow(1)
						idx := i
						ui.Button(c, "移除").Label("移除：" + q).OnClick(func() {
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
	if app.curArchived() {
		app.archivedBanner(c, p)
		return
	}
	box := ui.Column(c).Gap(Space2)
	if !app.sh.glass {
		box.Padding(Space2)
	}
	app.composerSurface(box, p)
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
					idx := i
					if app.sh.glass {
						app.kattach(c.Key(fmt.Sprint(i, ch.Label)), p, map[string]string{"file": "文件", "ref": "引用", "cmd": "命令", "image": "图片", "drop": "拖入"}[ch.Kind], ch.Label, func() {
							app.chips = append(app.chips[:idx:idx], app.chips[idx+1:]...)
						})
						continue
					}
					ui.Row(c.Key(fmt.Sprint(i, ch.Label))).Gap(Space1).AlignItems(ui.Center).Padding(2, 6).Radius(RadiusControl).Background(p.Surface).Border(1, p.Border).Children(func() {
						ui.Text(c, map[string]string{"file": "文件", "ref": "引用", "cmd": "命令", "image": "图片", "drop": "拖入"}[ch.Kind]+" · "+ch.Label).FontSize(FontSizeSmall)
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
		app.maybePanic("候选弹层")
		panel := ui.Column(c).AttachTo(box, ui.AnchorTopLeft, ui.AnchorBottomLeft).Margin(0, 0, 6, 0).Width(420).Padding(Space1).Gap(2).
			Role(ui.RoleListBox).Label("候选")
		if app.sh.glass {
			app.glassPanel(panel)
		} else {
			panel.Radius(RadiusPopover).Background(p.Raised).Border(1, p.Border).Shadow(0, 8, 28, 0, popoverShadow)
		}
		panel.Children(func() {
			for i, s := range cs {
				row := ui.ButtonBase(c.Key(s)).Justify(ui.Start).Padding(6, Space2).Radius(RadiusControl).Label(s).Role(ui.RoleOption)
				if app.sh.glass {
					row.Height(SizeControl).Padding(0, 10).Radius(10)
					if i == app.pick.hi {
						row.Background(p.Accent).TextColor(p.OnAccent)
					}
				} else if i == app.pick.hi {
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
	if app.sh.glass {
		app.kSettings(c, p)
		return
	}
	ui.Row(c).Gap(Space2).AlignItems(ui.Center).Children(func() {
		cat := app.catalog()
		names := make([]string, len(cat.Models))
		for i, m := range cat.Models {
			names[i] = m.Name
		}
		model := app.model
		ui.Select(c, &model, names).Label("模型")
		if model != app.model {
			app.setModel(model)
		}
		if m := cat.model(app.model); len(m.Efforts) > 0 {
			ui.Select(c, &app.effort, m.Efforts).Label("推理强度")
		} else {
			app.unsupportedSetting(c, p, "推理强度", "当前模型不提供推理强度")
		}
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

// btn：内容层按钮。玻璃外观下按组件方案绘制，标准外观用 MyGO 默认按钮。
func (app *App) btn(c *ui.Context, p Palette, label string, k kind) ui.Element {
	if app.sh.glass {
		return app.cbtn(c, p, label, k)
	}
	if k == kPrimary {
		return ui.PrimaryButton(c, label)
	}
	return ui.Button(c, label)
}
