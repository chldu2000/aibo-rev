//go:build ignore

// 合规样例：不得误报。
package good

import (
	"github.com/egoist/mygo/ui"
	"github.com/egoist/mygo/ui/markdown"
)

func view(c *ui.Context, p Palette) {
	ui.Text(c, "已发送，等待确认").TextColor(p.Text).FontSize(FontSizeBody)
	ui.Column(c).Background(p.Surface).Border(1, p.Border).FontSize(12)
	ui.Column(c).BorderEndWidth(1).BorderColor(p.Border) // 侧栏与内容之间的分隔线
	ui.Column(c).BorderWidth(0, 0, 1, 0)
	ui.Column(c).BorderWidth(2, 2, 2, 2)
	ui.ButtonBase(c).FocusRing(true)
	ui.Column(c).Shadow(0, 8, 28, 0, ui.RGBA(0, 0, 0, 0.12)) // ui-check-allow: color 阴影令牌是复合值，尚未生成
	_ = markdown.Theme{QuoteBarWidth: 0}
	_ = ui.Animation{NonMotion: true} // ui-check-allow: motion 只改变透明度
	_ = ui.Theme{FontSize: 13, FocusRingWidth: 2}
}
