package main

// 一次性实验：侧栏的工作区分组（界面方案 §4.1、§4.2）。
//
// 分组标题可折叠；未信任时带锁形标记；悬停显示路径；悬停或焦点时出现“新建会话”和“⋯”，右键与“⋯”
// 打开同一个菜单：新建会话、信任设置、在 Finder/资源管理器中显示、移除工作区。底部“添加工作区”“已归档”，
// 已归档进入与普通列表结构相同的归档列表视图。

import (
	"fmt"
	"path/filepath"
	"runtime"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

type workspace struct {
	name      string
	path      string
	trusted   bool
	collapsed bool
	removed   bool
	menuOpen  bool // 右键菜单
	moreOpen  bool // “⋯”菜单
}

var (
	icLock   = svg(`<rect x="3.5" y="7" width="9" height="6.5" rx="1.5"/><path d="M5.5 7V5a2.5 2.5 0 0 1 5 0v2"/>`)
	icPlus   = svg(`<path d="M8 3v10M3 8h10"/>`)
	icFolder = svg(`<path d="M2 4.5a1 1 0 0 1 1-1h3l1.5 1.5H13a1 1 0 0 1 1 1V12a1 1 0 0 1-1 1H3a1 1 0 0 1-1-1z"/>`)
	icTrust  = svg(`<path d="M8 2 13 4v4c0 3-2.2 5-5 6-2.8-1-5-3-5-6V4z"/><path d="M5.8 8.2 7.4 9.8 10.4 6.6"/>`)
	icRemove = svg(`<path d="M3 4.5h10M6.5 4.5V3h3v1.5M4.5 4.5l.7 8.5h5.6l.7-8.5"/>`)
	icBack   = svg(`<path d="M9.5 3.5 5 8l4.5 4.5"/>`)
)

func newWorkspaces() []*workspace {
	root, _ := filepath.Abs("../..") // 原型在 aibo-rev/prototypes/workbench-shell 下运行
	parent := filepath.Dir(root)
	return []*workspace{
		{name: "aibo-rev", path: root, trusted: true},
		{name: "aibo", path: filepath.Join(parent, "aibo"), trusted: true},
		{name: "aibo-plugins", path: filepath.Join(parent, "aibo-plugins")},
	}
}

func fileManager() string {
	if runtime.GOOS == "windows" {
		return "在资源管理器中显示"
	}
	return "在 Finder 中显示"
}

func (app *App) sessionsIn(w int, archived bool) []int {
	var pinned, rest []int
	for i, s := range app.sh.sessions {
		if s.ws != w || s.archived != archived {
			continue
		}
		if s.pinned && !archived {
			pinned = append(pinned, i)
		} else {
			rest = append(rest, i)
		}
	}
	return append(pinned, rest...)
}

func (app *App) archivedCount() int {
	n := 0
	for _, s := range app.sh.sessions {
		if s.archived && !app.sh.workspaces[s.ws].removed {
			n++
		}
	}
	return n
}

// sessionList：工作区分组（或归档列表视图），底部为“添加工作区”“已归档”。
func (app *App) sessionList(c *ui.Context, p Palette) {
	archive := app.sh.archiveView
	if archive {
		b := ui.ButtonBase(c.Key("archive-back")).Label("返回会话列表").Padding(4, Space2).Gap(6).AlignItems(ui.Center).Radius(RadiusControl).TextColor(p.Muted)
		if b.Hovered() {
			b.Background(p.Hover)
		}
		b.Children(func() {
			ui.Icon(c, icBack).Size(13, 13)
			ui.Text(c, "已归档").FontWeight(600).TextColor(p.Text)
		})
		if b.Clicked() {
			app.sh.archiveView = false
		}
	}
	shown := 0
	for w, ws := range app.sh.workspaces {
		if ws.removed {
			continue
		}
		rows := app.sessionsIn(w, archive)
		if archive && len(rows) == 0 {
			continue
		}
		shown++
		app.workspaceHeader(c, p, w)
		if ws.collapsed {
			continue
		}
		for _, i := range rows {
			app.sessionRow(c, p, i)
		}
		if len(rows) == 0 {
			ui.Text(c, "暂无会话").FontSize(FontSizeSmall).TextColor(p.Muted).Padding(2, Space2, Space1, 34)
		}
	}
	if archive && shown == 0 {
		ui.Text(c, "没有已归档的会话").FontSize(FontSizeSmall).TextColor(p.Muted).Padding(Space2)
	}
}

// sidebarFooter：侧栏底部的“添加工作区”“已归档”。
func (app *App) sidebarFooter(c *ui.Context, p Palette) {
	ui.Column(c).Padding(Space1, Space2, Space2, Space2).Gap(2).AlignItems(ui.Stretch).Children(func() {
		foot := func(label string, icon *ui.SVG, trailing string) ui.Element {
			b := ui.ButtonBase(c.Key("foot:"+label)).Label(label).Height(SizeControl).Padding(0, Space2).Gap(Space2).AlignItems(ui.Center).
				Radius(RadiusPill).TextColor(p.Muted)
			if b.Hovered() || b.Focused() {
				b.Background(p.Hover).TextColor(p.Text)
			}
			b.Children(func() {
				ui.Icon(c, icon).Size(14, 14)
				ui.Text(c, label).Grow(1)
				if trailing != "" {
					ui.Text(c, trailing).FontSize(FontSizeSmall)
				}
			})
			return b
		}
		if foot("添加工作区", icPlus, "").Clicked() {
			app.addWorkspace()
		}
		if n := app.archivedCount(); n > 0 || app.sh.archiveView {
			a := foot("已归档", icArchive, fmt.Sprint(n)).Selected(app.sh.archiveView)
			if app.sh.archiveView {
				a.Background(p.Hover).TextColor(p.Text)
			}
			if a.Clicked() {
				app.sh.archiveView = !app.sh.archiveView
			}
		}
	})
}

func (app *App) workspaceHeader(c *ui.Context, p Palette, w int) {
	ws := app.sh.workspaces[w]
	box := ui.Row(c.Key("ws:"+ws.name)).AlignItems(ui.Center).Height(SizeControl).Margin(Space2, 0, 2, 0).Padding(0, 4, 0, 0).Radius(RadiusPill)
	showActions := !app.sh.archiveView && (box.Hovered() || box.FocusWithin() || ws.menuOpen || ws.moreOpen)
	if box.Hovered() {
		box.Background(p.Hover)
	}
	box.Children(func() {
		h := ui.ButtonBase(c.Key("toggle")).Label(ws.name).Expanded(!ws.collapsed).Tooltip(ws.path).Description(ws.path).
			Grow(1).Shrink(1).Height(SizeControl).Padding(0, 6, 0, 8).Gap(6).AlignItems(ui.Center).Justify(ui.Start).TextColor(p.Muted)
		if !ws.trusted {
			h.Description(ws.path + "（未信任）")
		}
		h.Children(func() {
			ui.Icon(c, map[bool]*ui.SVG{true: icRight, false: icChevron}[ws.collapsed]).Size(11, 11)
			ui.Text(c, ws.name).FontSize(FontSizeSmall).FontWeight(600).MaxLines(1).Shrink(1)
			if !ws.trusted {
				ui.Icon(c, icLock).Size(12, 12).AccessHidden(true)
			}
		})
		if h.Clicked() {
			ws.collapsed = !ws.collapsed
		}
		if showActions {
			if app.rowAction(c, p, "在「"+ws.name+"」中新建会话", icPlus, false).Clicked() {
				app.newSession(w)
			}
			app.workspaceMoreMenu(c, p, w)
		}
	})
	if !app.sh.archiveView {
		m := ui.ContextMenuBase(c, box, &ws.menuOpen)
		m.Popup(func(panel ui.Element) {
			app.glassPanel(panel)
			app.glassItems(c, p, m, app.workspaceItems(w))
		})
	}
}

func (app *App) workspaceMoreMenu(c *ui.Context, p Palette, w int) {
	ws := app.sh.workspaces[w]
	m := ui.MenuBase(c, &ws.moreOpen)
	tr := m.Trigger.Label("「"+ws.name+"」工作区操作").Tooltip("工作区操作").Size(24, 24).Center().Radius(RadiusPill).TextColor(p.Muted)
	if tr.Hovered() || m.Open() {
		tr.Background(p.Hover).TextColor(p.Text)
	}
	tr.Children(func() { ui.Icon(c, icDots).Size(14, 14) })
	m.Popup(func(panel ui.Element) {
		panel.Margin(4, 0, 0, 0)
		app.glassPanel(panel)
		app.glassItems(c, p, m, app.workspaceItems(w))
	})
}

func (app *App) workspaceItems(w int) []gitem {
	ws := app.sh.workspaces[w]
	return []gitem{
		{label: "新建会话", icon: icPlus, hint: "⌘N", fn: func() { app.newSession(w) }},
		{label: map[bool]string{true: "撤销信任…", false: "信任此工作区"}[ws.trusted], icon: icTrust, fn: func() { app.toggleTrust(w) }},
		{label: fileManager(), icon: icFolder, fn: func() { mygo.Shell.ShowItemInFolder(ws.path) }},
		{sep: true},
		{label: "移除工作区…", icon: icRemove, danger: true, fn: func() { app.requestRemoveWorkspace(w) }},
	}
}

// newSession：原型直接在该工作区建一个会话并进入就地改名；正式流程是 §4.3 的新建会话弹层（选择 Agent、检查结果）。
func (app *App) newSession(w int) {
	ws := app.sh.workspaces[w]
	if !ws.trusted {
		app.toast("「"+ws.name+"」未信任：先信任工作区才能启动 Agent", nil)
		return
	}
	s := &session{name: "新会话", ws: w, cat: &catalogMatrix, opened: true}
	s.model, s.effort = catalogMatrix.defaults()
	app.sh.sessions = append(app.sh.sessions, s)
	ws.collapsed = false
	app.selectSession(len(app.sh.sessions) - 1)
	app.startRename(len(app.sh.sessions) - 1)
}

// toggleTrust：信任直接生效；撤销信任与停止活动是不同事实（决议 07），原型只演示标记与提示。
func (app *App) toggleTrust(w int) {
	ws := app.sh.workspaces[w]
	ws.trusted = !ws.trusted
	if ws.trusted {
		app.toast("已信任「"+ws.name+"」", nil)
	} else {
		app.toast("已撤销「"+ws.name+"」的信任，队列已暂停", func() { ws.trusted = true })
	}
}

func (app *App) workspaceBusy(w int) bool {
	for i, s := range app.sh.sessions {
		if s.ws == w && app.sessionRunning(i) {
			return true
		}
	}
	return false
}

// requestRemoveWorkspace：规则同归档——有活动时先确认停止；移除后保留业务历史、项目文件和暂停的队列。
func (app *App) requestRemoveWorkspace(w int) {
	app.sh.removeAsk = w
	app.sh.removeAskOpen = true
}

func (app *App) removeWorkspace(w int) {
	ws := app.sh.workspaces[w]
	ws.removed = true
	if app.sh.sessions[app.sh.cur].ws == w {
		for i, s := range app.sh.sessions {
			if !app.sh.workspaces[s.ws].removed && !s.archived {
				app.selectSession(i)
				break
			}
		}
	}
	app.toast("已移除「"+ws.name+"」", func() { ws.removed = false })
}

func (app *App) removeDialog(c *ui.Context) {
	if !app.sh.removeAskOpen {
		return
	}
	w := app.sh.removeAsk
	ws := app.sh.workspaces[w]
	msg := "保留业务历史、项目文件和暂停的队列；可在搜索中按“已移除”来源找到历史。"
	buttons := []string{"取消", "移除"}
	if app.workspaceBusy(w) {
		msg = "这个工作区有正在进行的活动，停止得到确认后才会移除。" + msg
		buttons = []string{"取消", "停止后移除"}
	}
	switch ui.AlertDialog(c, &app.sh.removeAskOpen, "移除工作区「"+ws.name+"」？", msg, buttons...) {
	case 1:
		if app.workspaceBusy(w) {
			app.stop()
		}
		app.removeWorkspace(w)
	}
}

// addWorkspace：真实窗口中打开目录选择器；无窗口测试中加入一个示例目录。新加入的工作区默认未信任（F01.2）。
func (app *App) addWorkspace() {
	add := func(path string) {
		app.sh.workspaces = append(app.sh.workspaces, &workspace{name: filepath.Base(path), path: path})
	}
	if app.win == nil {
		add("/tmp/示例工作区")
		return
	}
	go func() {
		paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{Parent: app.win, Title: "添加工作区", Directory: true, ButtonLabel: "添加"})
		if err != nil || len(paths) == 0 {
			return
		}
		app.update(func() { add(paths[0]) })
	}()
}
