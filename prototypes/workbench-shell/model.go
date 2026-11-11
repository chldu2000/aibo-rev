package main

import (
	"fmt"
	"math/rand/v2"
	"strings"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/markdown"
)

// 一次性实验：会话模型与模拟数据。不是产品实现。

type Kind int

const (
	KindUser Kind = iota
	KindAgent
	KindTool
	KindApproval
	KindUnknown
)

type Item struct {
	ID       int
	Kind     Kind
	Text     string
	Done     bool
	Tool     string
	ToolOpen bool
	Approval string // pending / allowed / denied
}

type Chip struct {
	Kind  string // file / ref / cmd / image / drop
	Label string
}

type picker struct {
	open    bool
	trigger rune
	query   string
	start   int // 触发符在草稿中的位置（rune）
	hi      int
	closed  string // 用户按 Esc 关闭时的草稿，草稿变化前不再自动弹出
}

type perf struct {
	builds       int
	buildTotal   time.Duration
	buildMax     time.Duration
	publishedAt  time.Time
	latencies    []time.Duration
	streamStart  time.Time
	streamFrames int
}

type App struct {
	win   *mygo.Window
	items []*Item
	older []*Item // 尚未载入的更早历史
	next  int

	list ui.ListState
	md   markdown.Options
	dark bool

	draft  string
	chips  []Chip
	pick   picker
	model  string
	effort string

	applying  bool
	running   bool
	queue     []string
	queueOpen bool
	sent      []string // 测试用：已投递的消息

	perf perf
	log  []string

	sh shell // 外壳：多会话、工具面板、故障边界（shell.go）
}

var (
	models  = []string{"claude-opus-5-5", "claude-sonnet-5-5", "claude-haiku-5-5"}
	efforts = []string{"低", "中", "高"}
	files   = []string{"src/core/session.ts", "src/core/markdown.ts", "src/lib/Composer.svelte", "src/lib/Stream.svelte", "docs/design/ui-requirements.md", "docs/adr/0005-runtime-storage-and-packages.md"}
	refs    = []string{"修复队列暂停逻辑", "设计令牌生成", "ACP 连接层", "工具面板装载"}
	cmds    = []string{"/compact", "/review", "/clear", "/model"}
)

func newApp(history int) *App {
	app := &App{model: models[0], effort: efforts[1]}
	app.sh = newShell()
	r := rand.New(rand.NewPCG(1, 2))
	all := make([]*Item, 0, history)
	for i := 0; i < history; i++ {
		app.next++
		if i%2 == 0 {
			all = append(all, &Item{ID: app.next, Kind: KindUser, Text: fmt.Sprintf("第 %d 条：看一下 %s 的暂停逻辑", i, files[r.IntN(len(files))]), Done: true})
		} else {
			all = append(all, &Item{ID: app.next, Kind: KindAgent, Text: histReply(r, i), Done: true})
		}
	}
	// 先载入尾部 200 条，其余向上滚动时分段载入
	cut := max(0, len(all)-200)
	app.older, app.items = all[:cut], all[cut:]
	app.append(&Item{Kind: KindUnknown, Text: "上一条消息的投递结果未知：连接在发送后断开。"})
	return app
}

func histReply(r *rand.Rand, i int) string {
	s := []string{
		"队列在取消后应**保持暂停**，直到用户明确恢复。",
		"已核对 `session.ts`：`pause()` 只改变队列状态，不会取消当前回合。",
		"建议把投递尝试先写入业务库，再调用 `session/prompt`。",
		"这条历史用于测量长会话的滚动与挂载。",
	}
	out := s[r.IntN(len(s))]
	if i%7 == 1 {
		out += "\n\n- 第一点：保留未发送队列\n- 第二点：结果未知时不自动重发"
	}
	if i%23 == 1 {
		out += "\n\n```go\nfunc (q *Queue) Pause() { q.paused = true }\n```"
	}
	return out
}

func (app *App) append(it *Item) *Item {
	app.next++
	it.ID = app.next
	app.items = append(app.items, it)
	return it
}

// loadOlder 在顶部插入一段更早的历史。
func (app *App) loadOlder(n int) {
	if len(app.older) == 0 {
		return
	}
	cut := max(0, len(app.older)-n)
	chunk := app.older[cut:]
	app.older = app.older[:cut]
	app.items = append(append([]*Item{}, chunk...), app.items...)
}

func (app *App) loadAll() {
	app.items = append(app.older, app.items...)
	app.older = nil
}

func (app *App) pendingApprovals() int {
	n := 0
	for _, it := range app.items {
		if it.Kind == KindApproval && it.Approval == "pending" {
			n++
		}
	}
	return n
}

// ---- 模拟 Agent ----

const longReply = "我会先读取会话控制模块，再修改暂停逻辑。\n\n" +
	"## 修改方案\n\n取消当前执行时，**保留并暂停**未发送队列；恢复只能由用户明确触发。\n\n" +
	"```ts\nexport function cancel(session: Session) {\n  session.queue.pause('cancelled');\n  session.requestStop();\n  // 停止请求不代表已经停止\n  return session.awaitStopEvidence();\n}\n```\n\n" +
	"对应的 Go 宿主侧：\n\n" +
	"```go\nfunc (s *Session) Cancel(ctx context.Context) error {\n\ts.queue.Pause(PauseCancelled)\n\treturn s.conn.Cancel(ctx, s.id)\n}\n```\n\n" +
	"| 状态 | 队列 | 说明 |\n|---|---|---|\n| 运行中 | 推进 | 正常投递 |\n| 已取消 | 暂停 | 等待用户恢复 |\n| 结果未知 | 暂停 | 只允许手动重发 |\n\n" +
	"运行测试：\n\n```bash\ngo test ./internal/session -run TestCancelKeepsQueuePaused\n```\n\n" +
	"测试通过后，我会更新文档里的状态说明，并在状态条中显示“队列已暂停”。这一段文字稍长一些，用来观察流式输出时逐词显示、换行和代码块高亮的表现，以及跟随底部是否稳定。"

// startLong 模拟一次需要审批的长回复。
func (app *App) startLong() {
	app.running = true
	app.append(&Item{Kind: KindTool, Tool: "read_file  src/core/session.ts", Text: "export class Session {\n  queue = new Queue();\n  ...\n}", Done: true})
	app.append(&Item{Kind: KindApproval, Approval: "pending", Text: "Agent 请求运行命令：go test ./internal/session"})
}

func (app *App) approve(it *Item, ok bool) {
	if ok {
		it.Approval = "allowed"
		app.stream(longReply, 6)
	} else {
		it.Approval = "denied"
		app.running = false
	}
}

// stream 以每 16ms 若干字符的速度推送一条回复。
func (app *App) stream(src string, perTick int) {
	reply := app.append(&Item{Kind: KindAgent})
	app.running = true
	app.perf.streamStart = time.Now()
	app.perf.latencies = app.perf.latencies[:0]
	app.perf.streamFrames = 0
	if app.win == nil { // 无窗口测试：一次完成
		reply.Text, reply.Done, app.running = src, true, false
		return
	}
	runes := []rune(src)
	go func() {
		for i := 0; i < len(runes); {
			i = min(len(runes), i+perTick)
			text := string(runes[:i])
			time.Sleep(16 * time.Millisecond)
			stop := false
			app.update(func() {
				if reply.Done {
					stop = true
					return
				}
				reply.Text = text
				app.perf.publishedAt = time.Now()
			})
			if stop {
				return
			}
		}
		app.update(func() {
			reply.Done = true
			app.running = false
			app.drainQueue()
		})
	}()
}

func (app *App) stop() {
	for _, it := range app.items {
		if it.Kind == KindAgent && !it.Done {
			it.Done = true
		}
	}
	app.running = false
}

// send 投递或入队（执行中 Enter 入队，UI-COMP）。
func (app *App) send(text string) {
	text = strings.TrimSpace(text)
	if text == "" && len(app.chips) == 0 {
		return
	}
	for _, ch := range app.chips {
		text += " [" + ch.Label + "]"
	}
	app.chips = nil
	if app.running {
		app.queue = append(app.queue, text)
		return
	}
	app.deliver(text)
}

func (app *App) deliver(text string) {
	app.sent = append(app.sent, text)
	app.append(&Item{Kind: KindUser, Text: text, Done: true})
	app.stream("收到。我会先确认 **"+firstRunes(text, 12)+"** 涉及的文件，再给出修改建议。", 3)
}

func (app *App) drainQueue() {
	if len(app.queue) == 0 {
		return
	}
	next := app.queue[0]
	app.queue = app.queue[1:]
	app.deliver(next)
}

func firstRunes(s string, n int) string {
	r := []rune(s)
	if len(r) > n {
		return string(r[:n]) + "…"
	}
	return s
}

func (app *App) candidates() []string {
	var src []string
	switch app.pick.trigger {
	case '@':
		src = files
	case '#':
		src = refs
	case '/':
		src = cmds
	}
	var out []string
	for _, s := range src {
		if strings.Contains(strings.ToLower(s), strings.ToLower(app.pick.query)) {
			out = append(out, s)
		}
	}
	return out
}

// detectTrigger 根据光标前的文字判断是否显示候选弹层。
func (app *App) detectTrigger(text string, caret int) {
	r := []rune(text)
	caret = min(caret, len(r))
	i := caret
	for i > 0 && r[i-1] != ' ' && r[i-1] != '\n' {
		i--
	}
	tok := r[i:caret]
	if text == app.pick.closed {
		app.pick.open = false
		return
	}
	app.pick.closed = ""
	if len(tok) > 0 && (tok[0] == '@' || tok[0] == '#' || (tok[0] == '/' && i == 0)) {
		q := string(tok[1:])
		if !app.pick.open || app.pick.trigger != tok[0] || app.pick.query != q {
			app.pick.hi = 0
		}
		app.pick.open, app.pick.trigger, app.pick.query, app.pick.start = true, tok[0], q, i
		if len(app.candidates()) == 0 {
			app.pick.open = false
		}
		return
	}
	app.pick.open = false
}

// choose 把候选变成输入框上方的标签，并从草稿中去掉触发文字；返回新的草稿。
func (app *App) choose(text string, caret int, idx int) string {
	cs := app.candidates()
	if idx < 0 || idx >= len(cs) {
		return text
	}
	kind := map[rune]string{'@': "file", '#': "ref", '/': "cmd"}[app.pick.trigger]
	app.chips = append(app.chips, Chip{Kind: kind, Label: cs[idx]})
	r := []rune(text)
	caret = min(caret, len(r))
	out := string(r[:app.pick.start]) + string(r[caret:])
	app.pick = picker{}
	return out
}

func (app *App) changeModel() {
	app.applying = true
	if app.win == nil {
		return
	}
	go func() {
		time.Sleep(900 * time.Millisecond)
		app.update(func() { app.applying = false })
	}()
}
