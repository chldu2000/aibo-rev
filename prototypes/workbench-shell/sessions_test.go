package main

import (
	"image/png"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

// 侧栏顺序中各会话名首次出现的位置。
func order(tt *ui.Tester, names ...string) []int {
	texts := tt.Texts()
	var r []int
	for _, n := range names {
		r = append(r, slices.Index(texts, n))
	}
	return r
}

func hover(tt *ui.Tester, label string) {
	if b, ok := tt.Find(label); ok {
		tt.Move(b.X+b.W/2, b.Y+b.H/2)
		tt.Frame()
	}
}

// 悬停时出现“置顶”“归档”；置顶后排到最前；标题栏不再显示“Agent · 工作区”，只有 Agent 图标。
func TestSessionHoverActionsAndPin(t *testing.T) {
	app, tt := glassTester(t)
	if tt.HasText("置顶") {
		t.Fatal("actions shown without hover")
	}
	hover(tt, "工具面板装载")
	if !tt.HasText("置顶") || !tt.HasText("归档") {
		t.Fatalf("hover actions missing: %q", tt.Texts())
	}
	if err := tt.Click("置顶"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	if !app.sh.sessions[3].pinned {
		t.Fatal("not pinned")
	}
	o := order(tt, "工具面板装载", "ACP 连接层") // 同在 aibo 工作区
	if o[0] < 0 || o[0] > o[1] {
		t.Fatalf("pinned session not first: %v", o)
	}
	if tt.HasText("Claude Code · aibo-rev") || tt.HasText("Codex · aibo-rev") {
		t.Fatal("header still shows agent · workspace")
	}
	if _, ok := tt.Find("Claude Code"); !ok {
		t.Fatal("agent logo has no accessible name")
	}
}

// 右键菜单：全部操作；“在输入框中引用”加入引用标签；“复制会话信息”写入剪贴板。
func TestSessionContextMenu(t *testing.T) {
	app, tt := glassTester(t)
	if err := tt.RightClick("设计令牌生成"); err != nil {
		t.Fatal(err)
	}
	tt.Frame()
	for _, s := range []string{"打开", "置顶", "改名", "在输入框中引用", "复制会话信息", "归档"} {
		if !tt.HasText(s) {
			t.Fatalf("menu lacks %q: %q", s, tt.Texts())
		}
	}
	_ = tt.Click("在输入框中引用")
	tt.Frame()
	if len(app.chips) != 1 || app.chips[0].Kind != "ref" {
		t.Fatalf("chips = %+v", app.chips)
	}
	_ = tt.RightClick("设计令牌生成")
	tt.Frame()
	_ = tt.Click("复制会话信息")
	tt.Frame()
	if c := tt.Clipboard(); c == "" || !slices.Contains([]bool{true}, len(c) > 0) {
		t.Fatalf("clipboard = %q", c)
	}
}

// F2 就地改名：Enter 确认，Esc 取消。
func TestSessionRename(t *testing.T) {
	app, tt := glassTester(t)
	_ = tt.RightClick("ACP 连接层")
	tt.Frame()
	_ = tt.Click("改名")
	tt.Frame()
	tt.Frame()
	if !tt.Focused("会话名称") {
		t.Fatal("rename input not focused")
	}
	tt.Key(ui.Cmd, ui.KeyA)
	tt.Type("ACP 接入层")
	tt.Key(0, ui.KeyEnter)
	tt.Frame()
	if app.sh.sessions[2].name != "ACP 接入层" {
		t.Fatalf("name = %q", app.sh.sessions[2].name)
	}
	_ = tt.Click("ACP 接入层")
	tt.Key(0, ui.KeyF2)
	tt.Frame()
	tt.Frame()
	tt.Type("不要")
	tt.Key(0, ui.KeyEscape)
	tt.Frame()
	if app.sh.sessions[2].name != "ACP 接入层" || app.sh.sessions[2].renaming {
		t.Fatalf("escape did not cancel: %q", app.sh.sessions[2].name)
	}
}

// 归档：无活动时直接归档并可撤销；有活动时先确认“停止后归档”；归档会话可阅读，输入区换成“取消归档后继续”。
func TestSessionArchive(t *testing.T) {
	app, tt := glassTester(t)
	hover(tt, "设计令牌生成")
	_ = tt.Click("归档")
	tt.Frame()
	if !app.sh.sessions[1].archived || !tt.HasText("撤销") {
		t.Fatalf("archive/undo: %v %q", app.sh.sessions[1].archived, tt.Texts())
	}
	_ = tt.Click("撤销")
	tt.Frame()
	if app.sh.sessions[1].archived {
		t.Fatal("undo failed")
	}
	// 运行中的当前会话
	app.running = true
	tt.Frame()
	hover(tt, "修复队列暂停逻辑")
	_ = tt.Click("归档")
	tt.Frame()
	if app.sh.sessions[0].archived || !tt.HasText("停止后归档") {
		t.Fatalf("running session archived without confirm: %q", tt.Texts())
	}
	_ = tt.Click("停止后归档")
	tt.Frame()
	if !app.sh.sessions[0].archived || app.running {
		t.Fatal("stop-then-archive failed")
	}
	if _, hasInput := tt.Find("输入消息，@ 引用文件，# 引用会话，/ 命令"); !tt.HasText("此会话已归档，可以阅读和搜索，不能发送。") || hasInput {
		t.Fatalf("archived composer: %q", tt.Texts())
	}
	if !tt.HasText("已归档") || !tt.HasText("1") {
		t.Fatal("archived entry missing")
	}
	_ = tt.Click("取消归档后继续")
	tt.Frame()
	if app.sh.sessions[0].archived {
		t.Fatal("unarchive failed")
	}
}

func TestSessionShots(t *testing.T) {
	if os.Getenv("SHOTS") == "" {
		t.Skip("SHOTS=1 时运行")
	}
	for _, dark := range []bool{false, true} {
		app := newApp(400)
		app.dark = dark
		app.sh.glass = true
		app.sh.sessions[3].pinned = true
		app.sh.sessions[1].archived = true
		tt := ui.NewTesterWith(app.view, 1280, 820, ui.TesterOptions{Dark: dark})
		tt.SetScale(2)
		tt.Frame()
		_ = tt.RightClick("ACP 连接层")
		tt.Frame()
		hover(tt, "ACP 连接层")
		tt.Advance(time.Second)
		tt.Frame()
		f, _ := os.Create("reports/sessions" + map[bool]string{false: "-light", true: "-dark"}[dark] + ".png")
		_ = png.Encode(f, tt.Image())
		f.Close()
	}
}
