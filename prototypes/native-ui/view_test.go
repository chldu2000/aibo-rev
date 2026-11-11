package main

import (
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

func setup(t *testing.T, history int) (*App, *ui.Tester) {
	t.Helper()
	app := newApp(history)
	tt := ui.NewTester(app.view, 1280, 820)
	if err := tt.Click("消息"); err != nil {
		t.Fatalf("聚焦输入框: %v", err)
	}
	return app, tt
}

func lastSent(app *App) string {
	if len(app.sent) == 0 {
		return ""
	}
	return app.sent[len(app.sent)-1]
}

// UI-COMP：Enter 发送，Shift+Enter 换行。
func TestEnterSendsShiftEnterBreaks(t *testing.T) {
	app, tt := setup(t, 20)
	tt.Type("你好")
	tt.Key(ui.Shift, ui.KeyEnter)
	tt.Type("第二行")
	if app.draft != "你好\n第二行" {
		t.Fatalf("草稿 %q", app.draft)
	}
	tt.Key(0, ui.KeyEnter)
	if lastSent(app) != "你好\n第二行" || app.draft != "" {
		t.Fatalf("发送 %q，草稿 %q", lastSent(app), app.draft)
	}
}

// 输入法组字时的 Enter 只确认候选，不发送。
func TestIMEEnterDoesNotSend(t *testing.T) {
	app, tt := setup(t, 20)
	tt.Compose("ce shi", 6)
	tt.Key(0, ui.KeyEnter)
	if len(app.sent) != 0 {
		t.Fatalf("组字时 Enter 发送了 %q", app.sent)
	}
	tt.Type("测试") // 提交组字结果
	if app.draft != "测试" {
		t.Fatalf("草稿 %q", app.draft)
	}
	tt.Key(0, ui.KeyEnter)
	if lastSent(app) != "测试" {
		t.Fatalf("发送 %q", lastSent(app))
	}
}

// @ 候选：焦点留在输入框，方向键选择，Enter 变成标签并去掉触发文字。
func TestTriggerPicker(t *testing.T) {
	app, tt := setup(t, 20)
	tt.Type("看 @src/co")
	if !tt.HasText("src/core/session.ts") || !tt.HasText("src/core/markdown.ts") {
		t.Fatalf("候选未显示: %q", tt.Texts())
	}
	if !tt.Focused("消息") {
		t.Fatal("焦点离开了输入框")
	}
	tt.Key(0, ui.KeyDown)
	tt.Key(0, ui.KeyEnter)
	if len(app.chips) != 1 || app.chips[0].Label != "src/core/markdown.ts" {
		t.Fatalf("标签 %+v", app.chips)
	}
	if app.draft != "看 " || len(app.sent) != 0 {
		t.Fatalf("草稿 %q，已发送 %q", app.draft, app.sent)
	}
	if !tt.HasText("文件 · src/core/markdown.ts") {
		t.Fatal("标签未显示")
	}
}

func TestPickerEscape(t *testing.T) {
	app, tt := setup(t, 20)
	option := func() bool { _, ok := tt.AccessByRole(ui.RoleOption, "ACP 连接层"); return ok }
	tt.Type("#")
	if !option() {
		t.Fatal("# 候选未显示")
	}
	tt.Key(0, ui.KeyEscape)
	if option() || app.draft != "#" {
		t.Fatalf("Esc 后候选仍显示或草稿改变: %q", app.draft)
	}
	tt.Type("A")
	if !option() {
		t.Fatal("继续输入后候选应重新出现")
	}
}

// 执行中 Enter 入队；队列浮层 Esc 关闭，焦点回到触发器。
func TestRunningEnterQueuesAndPopover(t *testing.T) {
	app, tt := setup(t, 20)
	app.running = true
	tt.Frame()
	tt.Type("排队消息")
	tt.Key(0, ui.KeyEnter)
	if len(app.queue) != 1 || len(app.sent) != 0 {
		t.Fatalf("队列 %q 已发送 %q", app.queue, app.sent)
	}
	if err := tt.Click("队列 1"); err != nil {
		t.Fatal(err)
	}
	if !tt.HasText("待发送") {
		t.Fatal("队列浮层未打开")
	}
	tt.Key(0, ui.KeyEscape)
	if tt.HasText("待发送") {
		t.Fatal("Esc 未关闭浮层")
	}
	if !tt.Focused("队列 1") {
		t.Error("关闭后焦点未回到触发器")
	}
}

func TestApprovalInline(t *testing.T) {
	app, tt := setup(t, 20)
	app.startLong()
	tt.Frame()
	tt.Advance(time.Second) // 跟随底部的滚动动画
	if !tt.HasText("1 个审批待处理") {
		t.Fatal("状态条没有审批提示")
	}
	if err := tt.Click("允许一次"); err != nil {
		t.Fatal(err)
	}
	if !tt.HasText("已允许一次") || tt.HasText("1 个审批待处理") {
		t.Fatalf("审批状态未更新: %q", tt.Texts())
	}
}

// 设置应用中：Enter 不发送、草稿保留；发送按钮对读屏为禁用并给出原因。
func TestApplyingBlocksSend(t *testing.T) {
	app, tt := setup(t, 20)
	app.applying = true
	tt.Type("等设置生效")
	tt.Key(0, ui.KeyEnter)
	if len(app.sent) != 0 || app.draft != "等设置生效" {
		t.Fatalf("已发送 %q 草稿 %q", app.sent, app.draft)
	}
	a, ok := tt.AccessByRole(ui.RoleButton, "发送")
	if !ok || !a.Disabled || !strings.Contains(a.Description, "正在应用") {
		t.Fatalf("发送按钮: %+v", a)
	}
}

func TestEveryButtonHasName(t *testing.T) {
	app, tt := setup(t, 20)
	app.startLong()
	app.chips = append(app.chips, Chip{Kind: "file", Label: "a.go"})
	tt.Frame()
	for _, a := range tt.AccessAll() {
		if a.Role == ui.RoleButton && strings.TrimSpace(a.Name) == "" {
			t.Errorf("无名按钮 %v", a.Bounds)
		}
	}
}

// 一万条历史只构建可见行。
func TestVirtualization(t *testing.T) {
	app, tt := setup(t, 10000)
	app.loadAll()
	tt.Frame()
	n := 0
	for _, s := range tt.Texts() {
		if strings.HasPrefix(s, "第 ") {
			n++
		}
	}
	if n == 0 || n > 60 {
		t.Fatalf("构建了 %d 条用户消息", n)
	}
	t.Logf("10,000 条历史中构建的用户消息：%d", n)
}

// 流式输出时每帧的构建 + 布局 + 绘制耗时（CPU 渲染，偏保守）。
func TestStreamingFrameCost(t *testing.T) {
	app, tt := setup(t, 10000)
	app.loadAll()
	reply := app.append(&Item{Kind: KindAgent})
	runes := []rune(longReply)
	var total, worst time.Duration
	frames := 0
	for i := 0; i < len(runes); i += 6 {
		reply.Text = string(runes[:min(len(runes), i+6)])
		s := time.Now()
		tt.Frame()
		d := time.Since(s)
		total += d
		worst = max(worst, d)
		frames++
	}
	t.Logf("流式 %d 帧：平均 %.2f ms，最长 %.2f ms（含 CPU 绘制）", frames, ms(total)/float64(frames), ms(worst))
	// 打字：流式进行中向输入框输入一个字
	s := time.Now()
	tt.Type("字")
	t.Logf("流式期间输入一个字：%.2f ms", ms(time.Since(s)))
}
