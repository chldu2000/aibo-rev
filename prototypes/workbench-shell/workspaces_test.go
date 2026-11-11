package main

import (
	"image/png"
	"os"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

// 分组：标题可折叠；未信任带锁并在描述中说明；悬停出现新建与“⋯”；右键菜单项齐全。
func TestWorkspaceGroups(t *testing.T) {
	app, tt := glassTester(t)
	for _, n := range []string{"aibo-rev", "aibo", "aibo-plugins", "暂无会话"} {
		if !tt.HasText(n) {
			t.Fatalf("missing %q: %q", n, tt.Texts())
		}
	}
	_ = tt.Click("aibo")
	tt.Frame()
	if !app.sh.workspaces[1].collapsed || tt.HasText("ACP 连接层") {
		t.Fatal("collapse failed")
	}
	_ = tt.Click("aibo")
	tt.Frame()
	hover(tt, "aibo-rev")
	if _, ok := tt.Find("在「aibo-rev」中新建会话"); !ok {
		t.Fatal("hover new-session button missing")
	}
	_ = tt.RightClick("aibo-rev")
	tt.Frame()
	for _, s := range []string{"新建会话", "撤销信任…", fileManager(), "移除工作区…"} {
		if !tt.HasText(s) {
			t.Fatalf("menu lacks %q", s)
		}
	}
	tt.Key(0, ui.KeyEscape)
	_ = tt.RightClick("aibo-plugins")
	tt.Frame()
	if !tt.HasText("信任此工作区") {
		t.Fatal("untrusted workspace should offer trust")
	}
}

// 新建会话：在该工作区出现并进入改名；未信任工作区不能新建。
func TestWorkspaceNewSession(t *testing.T) {
	app, tt := glassTester(t)
	n := len(app.sh.sessions)
	app.newSession(2)
	if len(app.sh.sessions) != n {
		t.Fatal("created session in untrusted workspace")
	}
	app.newSession(1)
	tt.Frame()
	tt.Frame()
	if len(app.sh.sessions) != n+1 || app.sh.sessions[app.sh.cur].ws != 1 || !tt.Focused("会话名称") {
		t.Fatalf("new session: n=%d cur ws=%d", len(app.sh.sessions), app.sh.sessions[app.sh.cur].ws)
	}
	if app.catalog().Agent != "Claude Code" {
		t.Fatalf("new session agent = %q", app.catalog().Agent)
	}
}

// 移除工作区：先确认；有活动时为“停止后移除”；移除后分组消失，可撤销。
func TestWorkspaceRemove(t *testing.T) {
	app, tt := glassTester(t)
	app.running = true // 当前会话在 aibo-rev 中运行
	tt.Frame()
	app.requestRemoveWorkspace(0)
	tt.Frame()
	if !tt.HasText("停止后移除") {
		t.Fatalf("no stop-then-remove: %q", tt.Texts())
	}
	_ = tt.Click("停止后移除")
	tt.Frame()
	if !app.sh.workspaces[0].removed || app.running || tt.HasText("修复队列暂停逻辑") && app.sh.cur == 0 {
		t.Fatal("remove failed")
	}
	if app.sh.sessions[app.sh.cur].ws == 0 {
		t.Fatal("current session still in removed workspace")
	}
	_ = tt.Click("撤销")
	tt.Frame()
	if app.sh.workspaces[0].removed {
		t.Fatal("undo failed")
	}
}

// 已归档：底部入口进入归档列表视图，按工作区分组，可返回。
func TestArchiveView(t *testing.T) {
	app, tt := glassTester(t)
	app.archive(1)
	app.archive(2)
	tt.Frame()
	if err := tt.Click("已归档"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !app.sh.archiveView || !tt.HasText("设计令牌生成") || !tt.HasText("ACP 连接层") || tt.HasText("工具面板装载") {
		t.Fatalf("archive view: %q", tt.Texts())
	}
	if tt.HasText("aibo-plugins") {
		t.Fatal("empty group shown in archive view")
	}
	_ = tt.Click("返回会话列表")
	tt.Frame()
	if app.sh.archiveView {
		t.Fatal("back failed")
	}
}

func TestWorkspaceShots(t *testing.T) {
	if os.Getenv("SHOTS") == "" {
		t.Skip("SHOTS=1 时运行")
	}
	for _, dark := range []bool{false, true} {
		app := newApp(400)
		app.dark = dark
		app.sh.glass = true
		app.sh.sessions[3].pinned = true
		app.archive(1)
		app.sh.toast = toastState{}
		tt := ui.NewTesterWith(app.view, 1280, 820, ui.TesterOptions{Dark: dark})
		tt.SetScale(2)
		tt.Frame()
		hover(tt, "aibo")
		_ = tt.Click("「aibo」工作区操作")
		tt.Advance(time.Second)
		tt.Frame()
		f, _ := os.Create("reports/workspaces" + map[bool]string{false: "-light", true: "-dark"}[dark] + ".png")
		_ = png.Encode(f, tt.Image())
		f.Close()
	}
}
