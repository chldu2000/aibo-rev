package main

// 一次性实验：工作台外壳。多会话切换、同窗口工具面板 WebView、区域级 panic 边界、隐藏标题栏。
// 不是产品实现。

import (
	"context"
	"fmt"
	"os"
	"strings"
	"syscall"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

// session 保存非当前会话的内容；切换时与 App 上的当前会话字段交换。
type session struct {
	name  string
	items []*Item
	older []*Item
	list  ui.ListState
	draft string
	chips []Chip
}

type shell struct {
	cur      int
	sessions []*session

	glass     bool    // Liquid Glass 外观（glass.go）
	composerH float32 // 上一帧浮动输入区的高度

	panelOpen   bool
	panelW      float32
	panel       *mygo.WebView // 工具面板；属于窗口，窗口关闭时随之销毁
	panelStatus string

	// 故障边界：区域名 → panic 信息
	faults map[string]string
	// 故障注入（菜单/按钮触发）
	inject map[string]bool
	events []string
}

func newShell() shell {
	names := []string{"修复队列暂停逻辑", "设计令牌生成", "ACP 连接层", "工具面板装载"}
	sh := shell{panelOpen: true, panelW: 360, faults: map[string]string{}, inject: map[string]bool{}}
	for i, n := range names {
		s := &session{name: n}
		if i > 0 {
			for j := 0; j < 30; j++ {
				s.items = append(s.items, &Item{ID: 1_000_000*i + j, Kind: KindUser, Text: fmt.Sprintf("%s · 第 %d 条", n, j), Done: true})
			}
		}
		sh.sessions = append(sh.sessions, s)
	}
	return sh
}

var eventLog *os.File

func (app *App) update(fn func()) {
	if app.win != nil {
		app.win.Update(fn)
	} else {
		mygo.RunOnMain(fn)
	}
}

func (app *App) event(format string, a ...any) {
	e := time.Now().Format("15:04:05.000 ") + fmt.Sprintf(format, a...)
	app.sh.events = append(app.sh.events, e)
	if eventLog != nil {
		fmt.Fprintln(eventLog, e)
	}
	if len(app.sh.events) > 200 {
		app.sh.events = app.sh.events[1:]
	}
}

// selectSession 把当前会话的内容存回 sessions[cur]，再载入 sessions[i]。
func (app *App) selectSession(i int) {
	sh := &app.sh
	if i == sh.cur || i < 0 || i >= len(sh.sessions) {
		return
	}
	old := sh.sessions[sh.cur]
	old.items, old.older, old.list, old.draft, old.chips = app.items, app.older, app.list, app.draft, app.chips
	s := sh.sessions[i]
	app.items, app.older, app.list, app.draft, app.chips = s.items, s.older, s.list, s.draft, s.chips
	app.pick = picker{}
	sh.cur = i
}

// guard 是区域级故障边界：fn 构建时 panic，只让该区域显示故障说明，其余界面照常。
// 依据：MyGO 的 Element.Children 在 fn 正常返回后恢复父节点；在 box 的 Children 内 recover，
// 外层便能把内部嵌套元素留下的父节点指针恢复回来。
func (app *App) guard(c *ui.Context, name string, box ui.Element, fn func()) ui.Element {
	p := app.palette()
	if msg, bad := app.sh.faults[name]; bad {
		return box.Children(func() {
			ui.Column(c).Border(1, p.Danger).Radius(RadiusCard).Padding(Space3).Gap(Space2).Margin(Space2).Role(ui.RoleGroup).Label(name + "故障").Children(func() {
				ui.Text(c, name+"无法显示").FontWeight(600).TextColor(p.Danger)
				ui.Text(c, msg).FontSize(FontSizeSmall).TextColor(p.Muted)
				if ui.Button(c, "重试："+name).Clicked() {
					delete(app.sh.faults, name)
					delete(app.sh.inject, name)
				}
			})
		})
	}
	return box.Children(func() {
		defer func() {
			if r := recover(); r != nil {
				app.sh.faults[name] = fmt.Sprint(r)
				app.event("panic 已隔离：%s：%v", name, r)
				c.Invalidate()
			}
		}()
		fn()
	})
}

func (app *App) maybePanic(name string) {
	if app.sh.inject[name] {
		panic("注入的故障：" + name)
	}
}

// view 是整窗布局：侧栏 | 会话区 | 工具面板（可拖动分隔条）。
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

	if app.sh.glass && c.Vibrancy() {
		c.Root().Background(ui.Transparent) // 侧栏透出窗口材质
	}
	if c.Shortcut(ui.Cmd, ui.KeyBackslash) {
		app.sh.panelOpen = !app.sh.panelOpen
	}

	ui.Row(c).Fill().AlignItems(ui.Stretch).Children(func() {
		app.guard(c, "侧栏", ui.Column(c).Width(SizeSidebar).AlignItems(ui.Stretch), func() { app.sidebar(c, p) })
		if !app.sh.panelOpen {
			app.mainColumn(c, p, ui.Column(c).Grow(1).AlignItems(ui.Stretch))
			return
		}
		w, _ := c.Size()
		room := w - SizeSidebar
		left := room - app.sh.panelW
		ui.Split(c, &left, func() {
			app.mainColumn(c, p, ui.Column(c).Fill().AlignItems(ui.Stretch))
		}, func() {
			app.guard(c, "工具面板", ui.Column(c).Fill().AlignItems(ui.Stretch).Background(p.Bg).BorderWidth(0, 0, 0, 1).BorderColor(p.Border), func() { app.panelView(c, p) })
		}).Grow(1).AlignSelf(ui.Stretch).SplitReset(room - 360)
		app.sh.panelW = max(240, room-left)
	})
}

func (app *App) mainColumn(c *ui.Context, p Palette, box ui.Element) {
	if app.sh.glass {
		app.glassMain(c, p, box)
		return
	}
	box.Children(func() {
		app.header(c, p)
		app.guard(c, "会话内容", ui.Column(c).Grow(1).AlignItems(ui.Stretch), func() { app.streamView(c, p) })
		app.guard(c, "输入区", ui.Column(c).AlignSelf(ui.Center).Width(SizeStreamMax).MaxWidthPercent(100).Padding(0, Space4, Space4, Space4).Gap(Space2), func() {
			if !app.list.AtEnd() {
				if ui.Button(c, "有新内容 ↓").AlignSelf(ui.Center).Clicked() {
					app.list.ScrollToEnd()
				}
			}
			app.statusStrip(c, p)
			app.composer(c, p)
		})
	})
}

func (app *App) sidebar(c *ui.Context, p Palette) {
	bar := c.TitleBar()
	side := ui.Column(c).Grow(1).AlignItems(ui.Stretch)
	if !app.sh.glass || !c.Vibrancy() {
		side.Background(p.Surface).BorderWidth(0, 1, 0, 0).BorderColor(p.Border)
	}
	side.Children(func() {
		// 标题栏区域：交通灯右侧留白，整条可拖动窗口
		ui.Row(c).Height(max(bar.Height, SizeHeader)).Padding(0, Space2, 0, bar.Left+Space2).DragWindow().Label("窗口拖动区（侧栏）")
		ui.Column(c).Padding(Space1, Space2).Gap(Space1).Children(func() {
			app.maybePanic("侧栏")
			ui.Text(c, "会话").FontSize(FontSizeSmall).TextColor(p.Muted).Padding(Space1, Space2)
			for i, s := range app.sh.sessions {
				row := ui.ButtonBase(c.Key(s.name)).Justify(ui.Start).Padding(6, Space2).Radius(RadiusControl).Label(s.name)
				if i == app.sh.cur {
					row.Background(p.Select)
				} else if row.Hovered() {
					row.Background(p.Hover)
				}
				row.Children(func() {
					ui.Text(c, s.name).FontWeight(map[bool]int{true: 600, false: 400}[i == app.sh.cur])
				})
				idx := i
				row.OnClick(func() { app.selectSession(idx) })
			}
		})
		ui.Spacer(c).DragWindow()
	})
}

func (app *App) header(c *ui.Context, p Palette) {
	bar := c.TitleBar()
	hdr := ui.Row(c).Height(max(bar.Height, SizeHeader)).AlignItems(ui.Center).Gap(Space2).Padding(0, Space4).DragWindow()
	if !app.sh.glass {
		hdr.BorderWidth(0, 0, 1, 0).BorderColor(p.Border)
	}
	if hdr.WindowDoubleClicked() {
		app.event("标题栏双击（会话区）")
	}
	hdr.Children(func() {
		ui.Text(c, app.sh.sessions[app.sh.cur].name).FontSize(FontSizeTitle).FontWeight(600)
		ui.Text(c, "Claude Code · aibo-rev").TextColor(p.Muted)
		ui.Spacer(c)
		if ui.Button(c, "开始长回复").Disabled(app.running).Clicked() {
			app.startLong()
		}
		if ui.Button(c, map[bool]string{true: "收起工具面板", false: "展开工具面板"}[app.sh.panelOpen]).Clicked() {
			app.sh.panelOpen = !app.sh.panelOpen
		}
		if ui.Button(c, map[bool]string{true: "标准外观", false: "玻璃外观"}[app.sh.glass]).Clicked() {
			app.setGlass(!app.sh.glass)
		}
		ui.MenuButton(c, "故障注入", func(m *ui.Menu) {
			for _, name := range []string{"侧栏", "会话内容", "工具卡", "输入区", "候选弹层", "工具面板"} {
				if m.Item("让「" + name + "」构建时 panic").Chosen() {
					app.sh.inject[name] = true
				}
			}
		})
	})
}

// item 外包一层故障边界：单条内容出错只影响该条。
func (app *App) guardedItem(c *ui.Context, p Palette, it *Item) {
	app.guard(c, fmt.Sprintf("条目 %d", it.ID), ui.Column(c).AlignItems(ui.Stretch), func() {
		if it.Kind == KindTool {
			app.maybePanic("工具卡")
		}
		app.item(c, p, it)
	})
}

// ---- 工具面板 ----

func (app *App) panelView(c *ui.Context, p Palette) {
	ui.Row(c).Height(SizeHeader).AlignItems(ui.Center).Gap(Space1).Padding(0, Space2).BorderWidth(0, 0, 1, 0).BorderColor(p.Border).DragWindow().Children(func() {
		app.maybePanic("工具面板")
		ui.Text(c, "示例包 · 面板").FontWeight(600)
		ui.Spacer(c)
		if ui.Button(c, "重新载入").Label("重新载入工具面板").Clicked() {
			go app.reloadPanel()
		}
		ui.MenuButton(c, "…", func(m *ui.Menu) {
			if m.Item("让面板页面卡死").Chosen() {
				go app.evalPanel("setTimeout(() => { for (;;) {} }, 0); return 1")
			}
			if m.Item("结束面板渲染进程").Chosen() {
				app.killPanel()
			}
			if m.Item("收起").Chosen() {
				app.sh.panelOpen = false
			}
		}).Label("面板操作")
	})
	if app.sh.panelStatus != "" {
		ui.Text(c, app.sh.panelStatus).FontSize(FontSizeSmall).TextColor(p.Muted).Padding(Space1, Space2)
	}
	if app.sh.panel == nil {
		ui.Text(c, "（无窗口测试：此处为工具面板 WebView）").TextColor(p.Muted).Padding(Space3)
		ui.Spacer(c)
		return
	}
	ui.WebView(c, app.sh.panel).Grow(1).Label("工具面板：示例包")
}

func (app *App) setPanelStatus(s string) {
	app.update(func() { app.sh.panelStatus = s; app.event("面板：%s", s) })
}

func (app *App) evalPanel(code string) (any, error) {
	if app.sh.panel == nil {
		return nil, fmt.Errorf("no panel")
	}
	ctx, cancel := context.WithTimeout(context.Background(), 1500*time.Millisecond)
	defer cancel()
	return app.sh.panel.Page().EvalContext(ctx, code)
}

func (app *App) killPanel() {
	if app.sh.panel == nil {
		return
	}
	if pid := app.sh.panel.Page().ProcessID(); pid > 0 && pid != syscall.Getpid() {
		_ = syscall.Kill(pid, syscall.SIGKILL)
		app.event("已结束面板渲染进程 %d", pid)
	}
}

// reloadPanel：页面有响应时直接 Reload；无响应时先结束渲染进程再 Reload（mygo-validation 发现死循环页面 Reload 无效）。
func (app *App) reloadPanel() {
	if _, err := app.evalPanel("return 1"); err != nil {
		app.setPanelStatus("面板无响应，结束渲染进程后重新载入…")
		app.killPanel()
		time.Sleep(300 * time.Millisecond)
	}
	app.sh.panel.Page().Reload()
	for i := 0; i < 30; i++ {
		time.Sleep(100 * time.Millisecond)
		if _, err := app.evalPanel("return document.readyState"); err == nil {
			app.setPanelStatus(fmt.Sprintf("已重新载入（%d ms）", (i+1)*100))
			return
		}
	}
	app.setPanelStatus("重新载入超时")
}

// makePanel 为当前窗口创建工具面板 WebView。窗口重建时须重新创建：页面状态由包自己持久化。
func (app *App) makePanel() {
	wv, err := app.win.NewWebView(mygo.WebViewOptions{URL: "aibopkg://pkg-demo/panel.html", BackgroundColor: "light-dark(#ffffff, #1e1f22)"})
	if err != nil {
		app.event("创建面板失败：%v", err)
		return
	}
	app.sh.panel = wv
	wv.Page().OnRenderProcessGone(func(reason string) {
		app.setPanelStatus("面板渲染进程已退出（" + reason + "），点“重新载入”恢复")
	})
}

// popoverShadow 是弹层阴影的颜色。tokens.json 中的 shadow 是复合值，生成器尚未输出它。
var popoverShadow = ui.RGBA(0, 0, 0, 0.12)

func shortErr(err error) string {
	if err == nil {
		return "ok"
	}
	s := err.Error()
	if i := strings.Index(s, "\n"); i > 0 {
		s = s[:i]
	}
	return s
}
