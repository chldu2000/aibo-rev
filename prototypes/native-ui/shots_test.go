package main

import (
	"image/png"
	"os"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

// 写出关键状态的截图（测试器的 CPU 绘制），用于检查样式。SHOTS=1 时运行。
func TestShots(t *testing.T) {
	if os.Getenv("SHOTS") == "" {
		t.Skip("SHOTS=1 时运行")
	}
	_ = os.MkdirAll("reports", 0o755)
	for _, dark := range []bool{false, true} {
		app := newApp(40)
		app.dark = dark
		app.startLong()
		app.queue = []string{"补充测试用例", "更新文档"}
		app.running = true
		app.chips = []Chip{{Kind: "file", Label: "src/core/session.ts"}, {Kind: "image", Label: "粘贴的图片 568 KB"}}
		tt := ui.NewTesterWith(app.view, 1280, 820, ui.TesterOptions{Dark: dark})
		tt.SetScale(2)
		_ = tt.Click("消息")
		tt.Type("看一下 @src/c")
		tt.Advance(time.Second)
		name := map[bool]string{false: "light", true: "dark"}[dark]
		save(t, tt, "reports/shot-"+name+".png")
	}
}

func save(t *testing.T, tt *ui.Tester, path string) {
	f, err := os.Create(path)
	if err != nil {
		t.Fatal(err)
	}
	defer f.Close()
	if err := png.Encode(f, tt.Image()); err != nil {
		t.Fatal(err)
	}
}
