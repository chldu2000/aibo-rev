package main

import (
	"strings"
	"testing"

	"github.com/egoist/mygo/ui"
)

func newTester(t *testing.T, app *App) *ui.Tester {
	t.Helper()
	tt := ui.NewTester(app.view, 1280, 820)
	tt.Frame()
	return tt
}

func hasPrefix(texts []string, p string) bool {
	for _, s := range texts {
		if strings.HasPrefix(s, p) {
			return true
		}
	}
	return false
}

// 工具卡（List 行 → Column → Collapsible 内）构建时 panic：只有该条显示故障，其余照常，输入不受影响。
func TestPanicInListRowIsContained(t *testing.T) {
	app := newApp(400)
	app.startLong() // 追加一个工具卡与审批卡
	tt := newTester(t, app)
	app.sh.inject["工具卡"] = true
	tt.Frame()
	tt.Frame()
	if len(app.sh.faults) != 1 {
		t.Fatalf("faults = %v", app.sh.faults)
	}
	if !hasPrefix(tt.Texts(), "条目 ") || !tt.HasText("需要你的批准") || !tt.HasText("修复队列暂停逻辑") {
		t.Fatalf("other regions missing: %q", tt.Texts())
	}
	if err := tt.Click("消息"); err != nil {
		t.Fatal(err)
	}
	tt.Type("panic 之后")
	if app.draft != "panic 之后" {
		t.Fatalf("draft = %q", app.draft)
	}
	// 审批按钮仍可用
	if err := tt.Click("允许一次"); err != nil {
		t.Fatal(err)
	}
	// 修复后重试
	delete(app.sh.inject, "工具卡")
	for k := range app.sh.faults {
		if err := tt.Click("重试：" + k); err != nil {
			t.Fatal(err)
		}
	}
	tt.Frame()
	if len(app.sh.faults) != 0 || !hasPrefix(tt.Texts(), "工具 · read_file") {
		t.Fatalf("after retry faults=%v", app.sh.faults)
	}
}

// 候选弹层（Overlay 内）panic：之后输入区、队列浮层等其它弹层仍正常。
func TestPanicInOverlayIsContained(t *testing.T) {
	app := newApp(50)
	tt := newTester(t, app)
	app.sh.inject["候选弹层"] = true
	if err := tt.Click("消息"); err != nil {
		t.Fatal(err)
	}
	tt.Type("@src")
	tt.Frame()
	tt.Frame()
	if _, ok := app.sh.faults["输入区"]; !ok {
		t.Fatalf("faults = %v", app.sh.faults)
	}
	if !tt.HasText("输入区无法显示") || !tt.HasText("会话") {
		t.Fatalf("texts %q", tt.Texts())
	}
	delete(app.sh.inject, "候选弹层")
	if err := tt.Click("重试：输入区"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	// 输入区恢复，草稿仍在（状态在 App 上）
	if app.draft != "@src" || !tt.Focused("消息") && tt.Click("消息") != nil {
		t.Fatalf("draft=%q", app.draft)
	}
	app.running = true
	tt.Key(0, ui.KeyEnter) // 弹层可能重新出现；先关闭
	tt.Key(0, ui.KeyEscape)
	tt.Type(" 继续")
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if len(app.queue) == 0 {
		t.Fatalf("queue empty, draft=%q", app.draft)
	}
	if err := tt.Click("队列 1"); err != nil {
		t.Fatal(err)
	}
	if !tt.HasText("待发送") {
		t.Fatalf("queue popover not open: %q", tt.Texts())
	}
}

// 侧栏与工具面板整区 panic。
func TestPanicInSidebarAndPanel(t *testing.T) {
	app := newApp(50)
	tt := newTester(t, app)
	app.sh.inject["侧栏"] = true
	app.sh.inject["工具面板"] = true
	tt.Frame()
	tt.Frame()
	if !tt.HasText("侧栏无法显示") || !tt.HasText("工具面板无法显示") || !tt.HasText("开始长回复") {
		t.Fatalf("texts %q", tt.Texts())
	}
	if err := tt.Click("消息"); err != nil {
		t.Fatal(err)
	}
	tt.Type("ok")
	if app.draft != "ok" {
		t.Fatal(app.draft)
	}
}

// 会话切换保留各自草稿与滚动位置。
func TestSessionSwitchKeepsDrafts(t *testing.T) {
	app := newApp(400)
	tt := newTester(t, app)
	_ = tt.Click("消息")
	tt.Type("会话一草稿")
	if err := tt.Click("ACP 连接层"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if app.draft != "" || !tt.HasText("ACP 连接层 · 第 29 条") {
		t.Fatalf("draft=%q texts=%q", app.draft, tt.Texts())
	}
	_ = tt.Click("消息")
	tt.Type("会话三草稿")
	_ = tt.Click("修复队列暂停逻辑")
	tt.Frame()
	if app.draft != "会话一草稿" {
		t.Fatalf("draft=%q", app.draft)
	}
	_ = tt.Click("ACP 连接层")
	if app.draft != "会话三草稿" {
		t.Fatalf("draft=%q", app.draft)
	}
}

// 新窗口（新的 Tester 代表重建的视图）从宿主状态恢复：会话、草稿、面板宽度、滚动位置。
// 焦点与选区不在宿主状态中，记录实际结果。
func TestRebuildRestoresHostState(t *testing.T) {
	app := newApp(2000)
	tt := newTester(t, app)
	_ = tt.Click("设计令牌生成")
	_ = tt.Click("修复队列暂停逻辑")
	_ = tt.Click("消息")
	tt.Type("关窗前的草稿")
	app.sh.panelW = 420
	app.list.ScrollTo(20, ui.Start)
	tt.Frame()
	tt.Frame()
	first, _ := app.list.Visible()

	tt2 := newTester(t, app)
	tt2.Frame()
	first2, _ := app.list.Visible()
	t.Logf("first visible before=%d after=%d; focused(消息) after rebuild=%v", first, first2, tt2.Focused("消息"))
	if app.draft != "关窗前的草稿" || app.sh.panelW != 420 || app.sh.cur != 0 {
		t.Fatalf("draft=%q panelW=%v cur=%d", app.draft, app.sh.panelW, app.sh.cur)
	}
	// 输入框绑定 app.draft：在新视图中追加输入，确认编辑的是恢复后的草稿
	_ = tt2.Click("消息")
	tt2.Key(ui.Cmd, ui.KeyDown)
	tt2.Type("。")
	if app.draft != "关窗前的草稿。" {
		t.Fatalf("draft after typing in rebuilt view = %q", app.draft)
	}
	if first2 != first {
		t.Errorf("scroll not restored: %d → %d", first, first2)
	}
}

// 分隔条：拖动面板宽度后，窗口变宽时面板宽度不变，会话区变宽。
func TestSplitKeepsPanelWidth(t *testing.T) {
	app := newApp(50)
	tt := newTester(t, app)
	app.sh.panelW = 480
	tt.Frame()
	tt.SetSize(1600, 820)
	tt.Frame()
	tt.Frame()
	if app.sh.panelW != 480 {
		t.Fatalf("panelW = %v", app.sh.panelW)
	}
	tt.Key(ui.Cmd, ui.KeyBackslash)
	tt.Frame()
	if app.sh.panelOpen {
		t.Fatal("⌘\\ did not collapse")
	}
}
