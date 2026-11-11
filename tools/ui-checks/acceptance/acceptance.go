// Package acceptance 是界面硬性要求（docs/design/ui-requirements.md）的验收套件，
// 用 MyGO 的无窗口测试器 ui.Tester 驱动界面，只按测试契约（要求文档 §11）中的无障碍角色与名称定位元素。
//
// 实现在自己的测试中调用：
//
//	func TestAcceptance(t *testing.T) { acceptance.Run(t, myTarget{}) }
package acceptance

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

// Fixtures 是要求文档 §11.2 登记的夹具。
var Fixtures = []string{
	"running-approval", "delivery-unknown", "pending-question", "turn-failed",
	"untrusted", "idle-empty", "restart", "save-failure", "history-10k",
}

// ErrUnsupported 表示目标不提供这个已登记的夹具；对应用例显示为跳过，不算通过。
var ErrUnsupported = errors.New("目标不提供此夹具")

// Fixture 是某个夹具初始状态下的界面。
type Fixture struct {
	View func(c *ui.Context)
	// Stream 模拟运行核心发布下一段流式正文，结束时返回 false。history-10k 必须提供（UI-STREAM-05）。
	Stream func() bool
}

// Target 是一个被验收的界面实现。
type Target interface {
	// Open 返回夹具的初始界面。不认识的名称必须返回错误，不得忽略；
	// 已登记但不提供的夹具返回 ErrUnsupported。
	Open(fixture string) (Fixture, error)
}

// Case 是一条验收用例。ID 必须是要求文档中的编号，requirements 检查据此核对映射。
type Case struct {
	ID, Title string
	Fixture   string // 空表示用例自己打开夹具
	Run       func(t *testing.T, x *X)
}

// X 是用例的执行环境。
type X struct {
	*ui.Tester
	Target  Target
	Fixture Fixture
}

const width, height = 1280, 820

// Open 打开夹具并返回测试器；夹具不受支持时跳过当前用例。
func Open(t *testing.T, target Target, fixture string) *X {
	t.Helper()
	f, err := target.Open(fixture)
	if errors.Is(err, ErrUnsupported) {
		t.Skipf("目标不提供夹具 %s", fixture)
	}
	if err != nil {
		t.Fatalf("打开夹具 %s: %v", fixture, err)
	}
	tt := ui.NewTester(f.View, width, height)
	tt.SetLocale("zh-CN")
	tt.Frame()
	return &X{Tester: tt, Target: target, Fixture: f}
}

// Run 运行全部验收用例。
func Run(t *testing.T, target Target) {
	for _, c := range Cases {
		t.Run(c.ID+" "+c.Title, func(t *testing.T) {
			var x *X
			if c.Fixture != "" {
				x = Open(t, target, c.Fixture)
			} else {
				x = &X{Target: target}
			}
			c.Run(t, x)
		})
	}
}

// Count 返回无障碍名称为 name 的元素个数。
func (x *X) Count(role ui.Role, name string) int {
	n := 0
	for _, a := range x.AccessAll() {
		if a.Role == role && a.Name == name {
			n++
		}
	}
	return n
}

// TextWithPrefix 返回第一段以 prefix 开头的可见文字。
func (x *X) TextWithPrefix(prefix string) (string, bool) {
	for _, s := range x.Texts() {
		if strings.HasPrefix(s, prefix) {
			return s, true
		}
	}
	return "", false
}

// Composer 返回输入框的无障碍信息。
func (x *X) Composer(t *testing.T) ui.AccessInfo {
	t.Helper()
	a, ok := x.AccessByRole(ui.RoleTextField, NameComposer)
	if !ok {
		t.Fatalf("找不到输入框（%s）", NameComposer)
	}
	return a
}

// FocusComposer 点击输入框使其获得焦点。
func (x *X) FocusComposer(t *testing.T) {
	t.Helper()
	if err := x.Click(NameComposer); err != nil {
		t.Fatalf("聚焦输入框: %v", err)
	}
}

// Settle 推进虚拟时间，让滚动与过渡动画结束。
func (x *X) Settle() { x.Advance(time.Second) }
