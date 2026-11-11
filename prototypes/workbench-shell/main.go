// 工作台外壳原型（一次性实验，不是产品实现）：在原生 UI 原型之上验证同窗口工具面板、
// 区域级 panic 边界、隐藏标题栏、系统通知与关窗后的界面状态恢复。
//
//	go generate && CGO_ENABLED=0 go build -o bin/shell . && ./bin/shell
//	./bin/shell -auto reports/auto.json                 真实窗口自动检查
//	./bin/shell -close hide                             关闭窗口改为隐藏（保留面板页面）
//	CGO_ENABLED=0 go test ./...                         无窗口检查
package main

//go:generate sh -c "cd ../../tools/ui-checks && go run . tokens && cp ../../prototypes/native-ui/tokens_gen.go ../../prototypes/workbench-shell/"

import (
	"bytes"
	"embed"
	"flag"
	"fmt"
	"image"
	"image/color"
	"image/png"
	"io/fs"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/egoist/mygo"
	"github.com/egoist/mygo/ui"
)

//go:embed panel
var panelFS embed.FS

var closeMode = flag.String("close", "destroy", "关闭窗口：destroy 销毁窗口（状态留在宿主），hide 隐藏窗口")

func main() {
	auto := flag.String("auto", "", "自动检查并写出报告")
	logPath := flag.String("log", "", "事件日志文件")
	flag.Parse()
	if *logPath != "" {
		eventLog, _ = os.OpenFile(*logPath, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0o644)
	}
	app := newApp(10000)
	mygo.App.SetName("Aibo Workbench Shell Prototype")

	panels, _ := fs.Sub(panelFS, "panel")
	mygo.Protocol.Handle("aibopkg", http.FileServer(http.FS(panels)))

	// 关闭最后一个窗口后保持运行；Dock 点击或托盘重新打开。
	mygo.App.OnWindowAllClosed(func() { app.event("最后一个窗口已关闭，应用留在后台") })
	mygo.App.OnActivate(func(visible bool) {
		if !visible {
			app.openWindow()
		}
	})
	// 通知点击：ID 为 "<会话序号>/<条目 ID>"。须在 Run 之前注册，应用被通知重新启动时也能收到。
	mygo.App.OnNotificationClick(func(id string) {
		app.event("通知点击：%s", id)
		app.openWindow()
		app.update(func() { app.revealFromNotification(id) })
	})

	mygo.App.WhenReady(func() {
		app.dark = mygo.Theme.IsDark()
		mygo.Theme.OnUpdated(func() {
			app.dark = mygo.Theme.IsDark()
			app.md.Theme = nil
			if app.win != nil {
				app.win.Invalidate()
			}
		})
		mygo.App.SetMenu(mygo.NewMenu([]*mygo.MenuItem{
			{Role: mygo.RoleAppMenu},
			{Role: mygo.RoleEditMenu},
			{Label: "验证", Submenu: app.menuItems()},
			{Role: mygo.RoleWindowMenu},
		}))
		if _, err := mygo.NewTray(mygo.TrayOptions{Icon: trayIcon(), IconIsTemplate: true, ToolTip: "Aibo 外壳原型", Menu: mygo.NewMenu(app.menuItems())}); err != nil {
			app.event("托盘创建失败：%v", err)
		}
		app.openWindow()
		if *auto != "" {
			go runAuto(app, *auto)
		}
	})
	if err := mygo.App.Run(); err != nil {
		log.Fatal(err)
	}
}

func (app *App) menuItems() []*mygo.MenuItem {
	return []*mygo.MenuItem{
		{Label: "显示工作台", Click: func(*mygo.MenuItem, *mygo.Window) { app.openWindow() }},
		{Label: "8 秒后发送审批通知（可先切到别的应用）", Click: func(*mygo.MenuItem, *mygo.Window) {
			go func() {
				time.Sleep(8 * time.Second)
				app.update(func() { app.notifyApproval() })
			}()
		}},
		{Label: "打印事件日志", Click: func(*mygo.MenuItem, *mygo.Window) {
			app.update(func() { fmt.Println(strings.Join(app.sh.events, "\n")) })
		}},
	}
}

func (app *App) openWindow() {
	if app.win != nil {
		app.win.Show()
		app.win.Focus()
		return
	}
	w := mygo.NewWindow(mygo.WindowOptions{
		Title: "Aibo 工作台外壳原型", Width: 1280, Height: 820, MinWidth: SizeWindowMinWidth, MinHeight: SizeWindowMinHeight,
		TitleBarStyle: mygo.TitleBarHiddenInset,
		Content:       ui.View(app.view),
	})
	app.win = w
	w.OnClose(func(e *mygo.CloseEvent) {
		if *closeMode == "hide" {
			e.PreventDefault()
			w.Hide()
			app.event("窗口已隐藏（hide 模式）")
			return
		}
		app.event("窗口已销毁（destroy 模式）")
	})
	w.OnClosed(func() {
		if app.win == w {
			app.win = nil
			app.sh.panel = nil
		}
	})
	app.makePanel()
	w.Invalidate()
	app.event("窗口已创建")
}

// notifyApproval 在会话 0 追加一条审批请求，并发系统通知。
func (app *App) notifyApproval() {
	// 审批总是落在“修复队列暂停逻辑”会话；若当前不在该会话，直接写进它保存的内容。
	it := &Item{Kind: KindApproval, Approval: "pending", Text: "Agent 请求运行命令：go test ./internal/session（来自通知验证）"}
	if app.sh.cur == 0 {
		app.append(it)
	} else {
		app.next++
		it.ID = app.next
		s := app.sh.sessions[0]
		s.items = append(s.items, it)
	}
	id := fmt.Sprintf("0/%d", it.ID)
	n := mygo.NewNotification(mygo.NotificationOptions{ID: id, Group: "session-0", Title: "修复队列暂停逻辑：需要你的批准", Body: it.Text})
	n.OnClick(func() { n.Close() })
	go func() {
		err := n.Show()
		app.update(func() {
			app.event("通知 %s：supported=%v err=%v", id, mygo.NotificationsSupported(), err)
		})
	}()
}

// revealFromNotification 切到通知所属会话并滚到对应条目。
func (app *App) revealFromNotification(id string) {
	si, iid, ok := strings.Cut(id, "/")
	if !ok {
		return
	}
	s, _ := strconv.Atoi(si)
	target, _ := strconv.Atoi(iid)
	app.selectSession(s)
	for i, it := range app.items {
		if it.ID == target {
			app.list.ScrollTo(i, ui.Center)
			app.event("已定位到会话 %d 第 %d 项", s, i)
			return
		}
	}
	app.event("通知对应条目不在已载入内容中：%s", id)
}

func trayIcon() []byte {
	img := image.NewNRGBA(image.Rect(0, 0, 32, 32))
	for y := 6; y < 26; y++ {
		for x := 6; x < 26; x++ {
			img.Set(x, y, color.NRGBA{A: 255})
		}
	}
	var b bytes.Buffer
	_ = png.Encode(&b, img)
	return b.Bytes()
}
