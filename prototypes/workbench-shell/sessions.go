package main

// 一次性实验：侧栏会话条目。Agent 图标、悬停时的常用操作（置顶、归档）、右键菜单（全部操作）、
// 就地改名、归档与“已归档”列表。规则来自界面方案 §4.2。
//
// Agent 图标是示意的抽象图形；正式实现由 Agent 包提供图标，不使用任何品牌标志。

import (
	"fmt"
	"time"

	"github.com/egoist/mygo/plugins/glass"
	"github.com/egoist/mygo/ui"
)

var (
	icPin     = svg(`<path d="M9.5 2.5 13.5 6.5 11 7.5 8.5 10l.5 3-1 1-3-3-3 3M5 11 3 9l1-1 3 .5L9.5 6l1-2.5"/>`)
	icArchive = svg(`<rect x="2" y="3" width="12" height="3" rx="1"/><path d="M3 6v6.5a1 1 0 0 0 1 1h8a1 1 0 0 0 1-1V6M6.5 9h3"/>`)
	icUnarch  = svg(`<rect x="2" y="3" width="12" height="3" rx="1"/><path d="M3 6v6.5a1 1 0 0 0 1 1h8a1 1 0 0 0 1-1V6M8 12V8.5M6.5 10 8 8.5 9.5 10"/>`)
	icRename  = svg(`<path d="M10.5 2.5l3 3-8 8H2.5v-3z"/>`)
	icCopy    = svg(`<rect x="5" y="5" width="8.5" height="8.5" rx="1.5"/><path d="M3 10.5V3.5A1 1 0 0 1 4 2.5h6.5"/>`)
	icRef     = svg(`<path d="M6 2.5 4.5 13.5M11.5 2.5 10 13.5M2.5 6h11M2 10h11"/>`)
	icOpen    = svg(`<path d="M6.5 3.5h-3v9h9v-3M9 2.5h4.5V7M13.5 2.5 7.5 8.5"/>`)
	// 示意的 Agent 图标（抽象图形，不是品牌标志）
	logoA = svg(`<path d="M8 3.5a4.5 4.5 0 1 0 4.2 6" stroke-width="1.8"/>`)
	logoB = svg(`<path d="M4.5 5.5 7 8l-2.5 2.5M8.5 11h3" stroke-width="1.8"/>`)
)

// agentLogo：16 DIP 圆角方块，作为图像角色并以 Agent 名称为无障碍名称。
func (app *App) agentLogo(c *ui.Context, p Palette, agent string, size float32) ui.Element {
	bg, icon := p.Attention, logoA
	if agent == "Codex" {
		bg, icon = p.Text, logoB
	}
	return ui.Box(c).Size(size, size).Radius(size * 0.28).Center().Background(bg).TextColor(p.Bg).Role(ui.RoleImage).Label(agent).Children(func() {
		ui.Icon(c, icon).Size(size*0.75, size*0.75).AccessHidden(true)
	})
}

func (app *App) sessionRunning(i int) bool { return i == app.sh.cur && app.running }

func (app *App) sessionRow(c *ui.Context, p Palette, i int) {
	s := app.sh.sessions[i]
	cat := app.sh.sessions[i].cat
	selected := i == app.sh.cur
	box := ui.Row(c.Key("srow:"+s.name)).AlignItems(ui.Center).Gap(2).Height(SizeControl+4).Padding(0, 4, 0, 0)
	showActions := box.Hovered() || box.FocusWithin() || s.menuOpen
	// 外观随组件方案（A：选中为玻璃胶囊）
	switch {
	case selected && app.sh.glass && app.variant() == 0:
		box.Radius(RadiusPill).Material(glass.Glass{})
	case selected:
		box.Radius(RadiusCard).Background(p.Select)
	case box.Hovered():
		box.Radius(map[bool]float32{true: RadiusPill, false: RadiusCard}[app.sh.glass && app.variant() == 0]).Background(p.Hover)
	}
	if s.archived {
		box.Opacity(1) // 归档会话仍可阅读，文字弱化见下
	}
	box.Children(func() {
		if s.renaming {
			app.renameInput(c, p, i)
			return
		}
		main := ui.ButtonBase(c.Key("open")).Label(s.name).Grow(1).Shrink(1).Height(SizeControl+4).Padding(0, 6, 0, 10).Gap(Space2).AlignItems(ui.Center).Justify(ui.Start)
		if s.pinned {
			main.Description("已置顶")
		}
		main.Children(func() {
			app.agentLogo(c, p, cat.Agent, 16)
			fg := p.Text
			if s.archived {
				fg = p.Muted
			}
			ui.Text(c, s.name).MaxLines(1).Shrink(1).Grow(1).TextColor(fg).FontWeight(map[bool]int{true: 600, false: 400}[selected])
			if s.pinned && !showActions {
				ui.Icon(c, icPin).Size(12, 12).TextColor(p.Muted).AccessHidden(true)
			}
		})
		if main.Clicked() {
			app.selectSession(i)
		}
		if main.Focused() && main.Shortcut(0, ui.KeyF2) {
			app.startRename(i)
		}
		if showActions {
			if !s.archived {
				pin := app.rowAction(c, p, map[bool]string{true: "取消置顶", false: "置顶"}[s.pinned], icPin, s.pinned)
				if pin.Clicked() {
					s.pinned = !s.pinned
				}
				if app.rowAction(c, p, "归档", icArchive, false).Clicked() {
					app.requestArchive(i)
				}
			} else if app.rowAction(c, p, "取消归档", icUnarch, false).Clicked() {
				s.archived = false
			}
		}
	})
	app.sessionMenu(c, p, box, i)
}

// rowAction：会话行内的小图标按钮（悬停或键盘焦点在行内时出现）。
func (app *App) rowAction(c *ui.Context, p Palette, label string, icon *ui.SVG, on bool) ui.Element {
	b := ui.ButtonBase(c.Key("act:"+label)).Label(label).Tooltip(label).Size(24, 24).Center().Radius(RadiusPill).TextColor(p.Muted)
	if on {
		b.TextColor(p.Accent)
	}
	if b.Hovered() || b.Focused() {
		b.Background(p.Hover).TextColor(p.Text)
	}
	b.Children(func() { ui.Icon(c, icon).Size(14, 14) })
	return b
}

// sessionMenu：右键（或菜单键、Shift+F10）显示的全部操作。
func (app *App) sessionMenu(c *ui.Context, p Palette, target ui.Element, i int) {
	s := app.sh.sessions[i]
	m := ui.ContextMenuBase(c, target, &s.menuOpen)
	items := []gitem{
		{label: "打开", icon: icOpen, off: i == app.sh.cur, fn: func() { app.selectSession(i) }},
	}
	if !s.archived {
		items = append(items,
			gitem{label: map[bool]string{true: "取消置顶", false: "置顶"}[s.pinned], icon: icPin, fn: func() { s.pinned = !s.pinned }},
			gitem{label: "改名", icon: icRename, hint: "F2", fn: func() { app.startRename(i) }},
			gitem{label: "在输入框中引用", icon: icRef, off: i == app.sh.cur || app.curArchived(), fn: func() {
				app.chips = append(app.chips, Chip{Kind: "ref", Label: s.name})
			}},
		)
	}
	items = append(items,
		gitem{label: "复制会话信息", icon: icCopy, fn: func() {
			c.WriteClipboard(fmt.Sprintf("%s\nAgent：%s\n工作区：%s\n模型：%s", s.name, s.cat.Agent, app.sh.workspaces[s.ws].path, app.sessionModelLabel(i)))
			app.toast("已复制会话信息", nil)
		}},
		gitem{sep: true},
	)
	if s.archived {
		items = append(items, gitem{label: "取消归档", icon: icUnarch, fn: func() { s.archived = false }})
	} else {
		items = append(items, gitem{label: "归档", icon: icArchive, fn: func() { app.requestArchive(i) }})
	}
	m.Popup(func(panel ui.Element) {
		app.glassPanel(panel)
		app.glassItems(c, p, m, items)
	})
}

func (app *App) sessionModelLabel(i int) string {
	if i == app.sh.cur {
		return app.pickerLabel()
	}
	s := app.sh.sessions[i]
	return s.model + " · " + s.effort
}

// ---- 改名 ----

func (app *App) startRename(i int) {
	s := app.sh.sessions[i]
	s.renaming, s.renameText = true, s.name
	app.sh.renameFocus = true
}

func (app *App) renameInput(c *ui.Context, p Palette, i int) {
	s := app.sh.sessions[i]
	in := ui.TextInputBase(c.Key("rename"), &s.renameText).Label("会话名称").Grow(1).Height(SizeControl).Padding(0, 10).Radius(RadiusCard).
		Background(p.Raised).Border(1, p.Accent)
	if app.sh.renameFocus {
		in.Focus()
		app.sh.renameFocus = false
	}
	switch {
	case in.Submitted():
		if t := s.renameText; t != "" {
			s.name = t
		}
		s.renaming = false
	case in.Shortcut(0, ui.KeyEscape):
		s.renaming = false
	}
}

// ---- 归档 ----

// requestArchive：没有活动时直接归档并提供 5 秒撤销；有活动时先确认“停止后归档”（界面方案 §4.2）。
func (app *App) requestArchive(i int) {
	if app.sessionRunning(i) {
		app.sh.archiveAsk = i
		app.sh.archiveAskOpen = true
		return
	}
	app.archive(i)
}

func (app *App) archive(i int) {
	s := app.sh.sessions[i]
	s.archived, s.pinned = true, false
	app.toast("已归档「"+s.name+"」", func() { s.archived = false })
}

func (app *App) curArchived() bool { return app.sh.sessions[app.sh.cur].archived }

// archiveDialog：有活动时的确认框。
func (app *App) archiveDialog(c *ui.Context) {
	if !app.sh.archiveAskOpen {
		return
	}
	i := app.sh.archiveAsk
	s := app.sh.sessions[i]
	switch ui.AlertDialog(c, &app.sh.archiveAskOpen, "归档「"+s.name+"」前需要停止活动",
		"这个会话有正在进行的回合。停止得到确认后才会归档；未发送的队列会保留并标记为暂停。", "取消", "停止后归档") {
	case 1:
		app.stop()
		app.archive(i)
	}
}

// ---- 撤销提示 ----

type toastState struct {
	text  string
	undo  func()
	until time.Time
	seq   int
}

func (app *App) toast(text string, undo func()) {
	app.sh.toast.seq++
	seq := app.sh.toast.seq
	app.sh.toast = toastState{text: text, undo: undo, until: time.Now().Add(5 * time.Second), seq: seq}
	if app.win != nil {
		go func() {
			time.Sleep(5 * time.Second)
			app.update(func() {
				if app.sh.toast.seq == seq {
					app.sh.toast = toastState{seq: seq}
				}
			})
		}()
	}
}

// toastView：侧栏底部的玻璃提示，带“撤销”。
func (app *App) toastView(c *ui.Context, p Palette) {
	t := app.sh.toast
	if t.text == "" {
		return
	}
	ui.Row(c).Margin(0, Space2, Space2, Space2).Padding(6, 6, 6, 12).Gap(Space2).AlignItems(ui.Center).Radius(RadiusPill).
		Material(glass.Glass{}).Shadow(0, 6, 20, 0, popoverShadow).Role(ui.RoleStatus).Label(t.text).Children(func() {
		ui.Text(c, t.text).FontSize(FontSizeSmall).MaxLines(1).Shrink(1).Grow(1)
		if t.undo != nil {
			u := ui.ButtonBase(c.Key("undo")).Label("撤销").Padding(3, 10).Radius(RadiusPill).TextColor(p.Accent)
			if u.Hovered() {
				u.Background(p.Hover)
			}
			u.Children(func() { ui.Text(c, "撤销").FontSize(FontSizeSmall).FontWeight(600) })
			if u.Clicked() {
				t.undo()
				app.sh.toast = toastState{seq: app.sh.toast.seq}
			}
		}
	})
}

// archivedBanner：归档会话可以阅读、搜索，不能发送；输入区替换为“取消归档后继续”。
func (app *App) archivedBanner(c *ui.Context, p Palette) {
	ui.Row(c).Gap(Space2).AlignItems(ui.Center).Padding(Space1, Space1).Children(func() {
		ui.Icon(c, icArchive).Size(14, 14).TextColor(p.Muted)
		ui.Text(c, "此会话已归档，可以阅读和搜索，不能发送。").TextColor(p.Muted).Grow(1)
		if app.kbtn(c, p, "取消归档后继续", icUnarch, kPrimary, false).Clicked() {
			app.sh.sessions[app.sh.cur].archived = false
		}
	})
}
