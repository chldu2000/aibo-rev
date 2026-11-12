// MyGO 实机验证（一次性实验，不是产品实现）。
//
//	go build -o bin/validate . && ./bin/validate            手动验证
//	./bin/validate -auto reports/auto.json                  自动检查后退出
//	./bin/validate -children reports/children.json          启动子进程后保持运行（供崩溃脚本使用）
package main

import (
	"context"
	"embed"
	"encoding/json"
	"errors"
	"flag"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"os"

	"strings"
	"sync"

	"time"

	"github.com/egoist/mygo"
)

//go:embed all:web
var webFS embed.FS

//go:embed all:panels
var panelFS embed.FS

// ---- 事件记录 ----

type logEntry struct {
	T     string `json:"t"`
	Event string `json:"event"`
	Data  any    `json:"data,omitempty"`
}

var (
	logMu   sync.Mutex
	entries []logEntry
	start   = time.Now()
)

func record(event string, data any) {
	logMu.Lock()
	entries = append(entries, logEntry{T: fmt.Sprintf("%.3f", time.Since(start).Seconds()), Event: event, Data: data})
	logMu.Unlock()
	b, _ := json.Marshal(data)
	log.Printf("%s %s", event, b)
}

// ---- 绑定给页面的服务 ----

// Lab 是实验页调用的 Go 服务。
type Lab struct {
	mu       sync.Mutex
	lastBeat time.Time
}

// Msg 是 Channel 推送的消息。
type Msg struct {
	Seq    int    `json:"seq"`
	SentAt int64  `json:"sentAt"` // Unix 纳秒
	Body   string `json:"body"`
}

// StreamResult 记录宿主侧的发送情况。
type StreamResult struct {
	Sent         int     `json:"sent"`
	TotalMs      float64 `json:"totalMs"`
	MaxSendMs    float64 `json:"maxSendMs"`
	BlockedSends int     `json:"blockedSends"` // Send 超过 5ms 的次数，视为背压
	Canceled     bool    `json:"canceled"`
}

// Stream 推送 n 条 size 字节的消息，每条间隔 intervalMs（0 为尽快发送）。
func (l *Lab) Stream(ctx context.Context, n, size, intervalMs int, ch *mygo.Channel[Msg]) (StreamResult, error) {
	body := strings.Repeat("x", size)
	res := StreamResult{}
	t0 := time.Now()
	for i := 0; i < n; i++ {
		s := time.Now()
		if err := ch.Send(Msg{Seq: i, SentAt: time.Now().UnixNano(), Body: body}); err != nil {
			res.Canceled = errors.Is(err, mygo.ErrChannelClosed) || ctx.Err() != nil
			break
		}
		d := time.Since(s)
		if ms := float64(d.Microseconds()) / 1000; ms > res.MaxSendMs {
			res.MaxSendMs = ms
		}
		if d > 5*time.Millisecond {
			res.BlockedSends++
		}
		res.Sent++
		if intervalMs > 0 {
			select {
			case <-time.After(time.Duration(intervalMs) * time.Millisecond):
			case <-ctx.Done():
				res.Canceled = true
				res.TotalMs = float64(time.Since(t0).Milliseconds())
				record("stream", res)
				return res, nil
			}
		}
	}
	res.TotalMs = float64(time.Since(t0).Milliseconds())
	record("stream", res)
	return res, nil
}

// Echo 原样返回，用于测量往返延迟。
func (l *Lab) Echo(v any) any { return v }

// Beat 是页面心跳。
func (l *Lab) Beat() {
	l.mu.Lock()
	l.lastBeat = time.Now()
	l.mu.Unlock()
}

// Log 让页面写入事件记录。
func (l *Lab) Log(event string, data any) { record("page:"+event, data) }

func (l *Lab) sinceBeat() time.Duration {
	l.mu.Lock()
	defer l.mu.Unlock()
	if l.lastBeat.IsZero() {
		return 0
	}
	return time.Since(l.lastBeat)
}

// ---- 窗口与恢复入口 ----

var (
	lab      = &Lab{}
	mainWin  *mygo.Window
	winMu    sync.Mutex
	domReady = make(chan struct{}, 16)
	gone     = make(chan string, 4)
)

func page() *mygo.Page {
	winMu.Lock()
	defer winMu.Unlock()
	if mainWin == nil {
		return nil
	}
	return mainWin.Page()
}

func openMain(url string) {
	winMu.Lock()
	defer winMu.Unlock()
	if mainWin != nil {
		mainWin.Show()
		mainWin.Focus()
		return
	}
	w := mygo.NewWindow(mygo.WindowOptions{
		Title: "Aibo MyGO 验证", Width: 1280, Height: 820, URL: url,
		Page: mygo.PageOptions{DevTools: mygo.DevToolsEnabled, PreloadScript: heartbeatJS},
	})
	p := w.Page()
	p.OnDOMReady(func() {
		record("dom-ready", map[string]any{"url": p.URL(), "pid": p.ProcessID()})
		select {
		case domReady <- struct{}{}:
		default:
		}
	})
	p.OnDidFailLoad(func(err *mygo.LoadError) { record("did-fail-load", fmt.Sprint(err)) })
	p.OnRenderProcessGone(func(reason string) {
		record("render-process-gone", reason)
		select {
		case gone <- reason:
		default:
		}
	})
	w.OnClosed(func() {
		winMu.Lock()
		mainWin = nil
		winMu.Unlock()
		record("window-closed", nil)
	})
	mainWin = w
}

// recoverPage 实现“重新载入界面 / 切回默认界面”：页面仍有心跳时直接载入；
// 心跳中断时先结束 WebContent 进程再载入（实测死循环中的页面 Reload/LoadURL 无效）。
func recoverPage(load func(p *mygo.Page)) {
	p := page()
	if p == nil {
		openMain("/")
		return
	}
	if lab.sinceBeat() > 2*time.Second {
		if pid := p.ProcessID(); pid > 0 {
			for len(gone) > 0 {
				<-gone
			}
			_ = killProcess(pid)
			go func() {
				select {
				case <-gone:
				case <-time.After(3 * time.Second):
				}
				load(p)
			}()
			return
		}
	}
	load(p)
}

func reloadUI() {
	record("menu:reload", lab.sinceBeat().Seconds())
	recoverPage(func(p *mygo.Page) { p.Reload() })
}

func defaultUI() {
	record("menu:default-ui", lab.sinceBeat().Seconds())
	recoverPage(func(p *mygo.Page) { _ = p.LoadURL("/") })
}

func hangPage() {
	if p := page(); p != nil {
		record("menu:hang", nil)
		go func() {
			ctx, cancel := context.WithTimeout(context.Background(), 500*time.Millisecond)
			defer cancel()
			_, _ = p.EvalContext(ctx, "setTimeout(() => { for (;;) {} }, 0); return 1")
		}()
	}
}

func killRenderer() {
	if p := page(); p != nil {
		pid := p.ProcessID()
		record("menu:kill-renderer", pid)
		if pid > 0 {
			_ = killProcess(pid)
		}
	}
}

func menuItems() []*mygo.MenuItem {
	return []*mygo.MenuItem{
		{Label: "重新载入界面", Click: func(*mygo.MenuItem, *mygo.Window) { reloadUI() }},
		{Label: "切回默认界面", Click: func(*mygo.MenuItem, *mygo.Window) { defaultUI() }},
		mygo.Separator(),
		{Label: "打开实验页", Click: func(*mygo.MenuItem, *mygo.Window) {
			openMain("/lab/")
			if p := page(); p != nil {
				_ = p.LoadURL("/lab/")
			}
		}},
		{Label: "测试：让页面死循环", Click: func(*mygo.MenuItem, *mygo.Window) { hangPage() }},
		{Label: "测试：结束渲染进程", Click: func(*mygo.MenuItem, *mygo.Window) { killRenderer() }},
		{Label: "显示工作台", Click: func(*mygo.MenuItem, *mygo.Window) { openMain("/") }},
	}
}

var autoMode bool

// heartbeatJS 注入每个页面：每 500ms 向宿主报一次心跳。
const heartbeatJS = `setInterval(() => { try { window.mygo && window.mygo.call("Lab.Beat") } catch (e) {} }, 500);`

// watchHeartbeat 在页面心跳中断时弹出原生提示（Q89）。
func watchHeartbeat() {
	prompted := false
	for range time.Tick(500 * time.Millisecond) {
		gap := lab.sinceBeat()
		if gap > 4*time.Second && !prompted && page() != nil {
			prompted = true
			record("heartbeat-lost", gap.Seconds())
			if autoMode {
				continue // 自动模式只记录，避免模态对话框占住主线程
			}
			res, err := mygo.Dialog.Message(mygo.MessageOptions{
				Type:    mygo.MessageWarning,
				Message: "界面没有响应",
				Detail:  "会话仍在后台运行。可以继续等待，或重新载入界面。",
				Buttons: []string{"继续等待", "重新载入", "切回默认界面"},
			})
			record("heartbeat-prompt", map[string]any{"button": res.Button, "err": fmt.Sprint(err)})
			switch res.Button {
			case 1:
				reloadUI()
			case 2:
				defaultUI()
			}
		}
		if gap < time.Second {
			prompted = false
		}
	}
}

// ---- 子进程（Q87） ----

// ---- main ----

func main() {
	auto := flag.String("auto", "", "自动检查并把报告写到该路径")
	children := flag.String("children", "", "启动子进程并把进程号写到该路径")
	panelsReport := flag.String("panels-report", "", "双面板故障隔离报告（Windows）")
	panelsRecreateLoop := flag.Bool("panels-recreate-loop", false, "双面板实验：直接重建死循环 WebView")
	logPath := flag.String("log", "", "事件日志文件")
	flag.Parse()
	if *children != "" && runChildrenProbe(*children) {
		return
	}
	if *logPath != "" {
		if f, err := os.OpenFile(*logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644); err == nil {
			log.SetOutput(f)
		}
	}

	mygo.App.SetName("Aibo MyGO Validation")
	if !mygo.App.RequestSingleInstanceLock() {
		fmt.Println("second-instance: lock held by another instance, exiting")
		return
	}
	mygo.App.OnSecondInstance(func(args []string, wd string) {
		record("second-instance", map[string]any{"args": args, "wd": wd})
		openMain("/")
	})

	// 17 号票只使用实验页；旧默认呈现的构建产物不再是依赖。
	web, _ := fs.Sub(webFS, "web/lab")
	mygo.SetFrontend(overlay{app: web, root: webFS})
	panels, _ := fs.Sub(panelFS, "panels")
	// 每个包一个 host：aibopkg://pkg-a/、aibopkg://pkg-b/，origin 随 host 区分。
	mygo.Protocol.Handle("aibopkg", http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		record("protocol", map[string]any{"host": r.Host, "path": r.URL.Path})
		http.FileServer(http.FS(panels)).ServeHTTP(w, r)
	}))
	mygo.Bind(lab)

	// Q89：最后窗口关闭后保持运行。
	mygo.App.OnWindowAllClosed(func() { record("all-windows-closed", nil) })
	mygo.App.OnActivate(func(visible bool) {
		record("activate", visible)
		if !visible {
			openMain("/")
		}
	})

	mygo.App.WhenReady(func() {
		if *panelsReport != "" {
			go runPanelsProbe(*panelsReport, *panelsRecreateLoop)
			return
		}
		mygo.App.SetMenu(mygo.NewMenu([]*mygo.MenuItem{
			{Role: mygo.RoleAppMenu},
			{Role: mygo.RoleEditMenu},
			{Label: "恢复", Submenu: menuItems()},
			{Role: mygo.RoleWindowMenu},
		}))
		if _, err := mygo.NewTray(mygo.TrayOptions{Icon: trayIcon(), IconIsTemplate: true, ToolTip: "Aibo 验证", Menu: mygo.NewMenu(menuItems())}); err != nil {
			record("tray-error", err.Error())
		}
		openMain("/")
		go watchHeartbeat()
		if *children != "" {
			spawnChildren(*children)
		}
		if *auto != "" {
			autoMode = true
			go func() {
				rep := runAuto()
				b, _ := json.MarshalIndent(rep, "", "  ")
				_ = os.WriteFile(*auto, b, 0o644)
				mygo.App.Exit(0)
			}()
		}
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}

// overlay 把默认界面放在根路径，实验页放在 /lab/。
type overlay struct {
	app  fs.FS
	root fs.FS
}

func (o overlay) Open(name string) (fs.File, error) {
	if name == "lab" || strings.HasPrefix(name, "lab/") {
		return o.root.Open("web/" + name)
	}
	return o.app.Open(name)
}
