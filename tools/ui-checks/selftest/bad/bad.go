//go:build ignore

// 违规样例：每条源码规则至少命中一处。
package bad

import (
	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/markdown"
)

func view(c *ui.Context, p Palette) {
	ui.Text(c, "发送成功").TextColor(ui.Hex("#ff0000"))
	ui.Column(c).Background(ui.RGBA(0, 0, 0, 0.1)).Border(1, ui.Color{R: 1, A: 255})
	ui.Text(c, "小字").FontSize(11)
	ui.Column(c).BorderStartWidth(3).BorderColor(p.Accent)
	ui.Column(c).BorderWidth(0, 0, 0, 4)
	ui.ButtonBase(c).FocusRing(false)
	_ = markdown.Theme{QuoteBarWidth: 3}
	_ = ui.Animation{NonMotion: true}
	_ = ui.Theme{FontSize: 10, FocusRingWidth: 1}
	var th markdown.Theme
	th.QuoteBarWidth = 2
}
