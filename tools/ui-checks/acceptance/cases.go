package acceptance

import (
	"errors"
	"strings"
	"testing"
	"time"

	"github.com/egoist/mygo/ui"
)

// Cases 是全部验收用例。新增用例时 ID 用要求编号，要求的“验证”字段写 uitest 或 perf。
var Cases = []Case{
	{ID: "UI-GEN-04", Title: "验收接口：夹具可加载、核心元素存在、未知夹具报错", Run: func(t *testing.T, x *X) {
		opened := 0
		for _, f := range Fixtures {
			fx, err := x.Target.Open(f)
			if errors.Is(err, ErrUnsupported) {
				continue
			}
			if err != nil {
				t.Errorf("%s: %v", f, err)
				continue
			}
			opened++
			tt := ui.NewTester(fx.View, width, height)
			tt.SetLocale("zh-CN")
			tt.Frame()
			if _, ok := tt.Access(NameStream); !ok {
				t.Errorf("%s: 缺少会话流（%s）", f, NameStream)
			}
			if _, ok := tt.AccessByRole(ui.RoleTextField, NameComposer); !ok {
				t.Errorf("%s: 缺少输入框（%s）", f, NameComposer)
			}
		}
		if opened == 0 {
			t.Error("目标没有提供任何夹具")
		}
		if _, err := x.Target.Open("__no_such_fixture__"); err == nil || errors.Is(err, ErrUnsupported) {
			t.Errorf("未知夹具必须报错而不是忽略：%v", err)
		}
	}},

	{ID: "UI-STREAM-04", Title: "一万条历史虚拟化", Fixture: "history-10k", Run: func(t *testing.T, x *X) {
		rows := x.Count(ui.RoleGroup, NameUser) + x.Count(ui.RoleGroup, NameAgent)
		if rows == 0 || rows > 60 {
			t.Fatalf("构建了 %d 条消息，应在 1–60 之间", rows)
		}
		t.Logf("10,000 条历史中构建的消息：%d", rows)
	}},

	{ID: "UI-STREAM-05", Title: "流式输出不阻塞界面", Fixture: "history-10k", Run: func(t *testing.T, x *X) {
		if x.Fixture.Stream == nil {
			t.Fatal("history-10k 必须提供 Stream")
		}
		var worst, total time.Duration
		n := 0
		for x.Fixture.Stream() {
			s := time.Now()
			x.Frame()
			d := time.Since(s)
			total += d
			worst = max(worst, d)
			n++
		}
		if n == 0 {
			t.Fatal("Stream 没有发布任何内容")
		}
		t.Logf("流式 %d 帧：平均 %.2f ms，最长 %.2f ms（无窗口，含 CPU 绘制）", n, float64(total.Microseconds())/1000/float64(n), float64(worst.Microseconds())/1000)
		if worst > 50*time.Millisecond {
			t.Fatalf("最长一帧 %v，超过 50ms", worst)
		}
	}},

	{ID: "UI-COMP-01", Title: "组字期间按 Enter 不发送", Fixture: "idle-empty", Run: func(t *testing.T, x *X) {
		x.FocusComposer(t)
		before := x.Count(ui.RoleGroup, NameUser)
		x.Compose("zhong wen", 9)
		x.Key(0, ui.KeyEnter)
		if got := x.Count(ui.RoleGroup, NameUser); got != before {
			t.Fatal("组字中的 Enter 发送了消息")
		}
		x.Type("中文")
		if v := x.Composer(t).Value; v != "中文" {
			t.Fatalf("组字确认后输入框为 %q", v)
		}
		x.Key(0, ui.KeyEnter)
		x.Settle()
		if got := x.Count(ui.RoleGroup, NameUser); got != before+1 {
			t.Fatalf("组字结束后的 Enter 应发送一条消息，用户消息 %d → %d", before, got)
		}
	}},

	{ID: "UI-COMP-04", Title: "触发符候选可含路径，焦点留在输入框", Fixture: "idle-empty", Run: func(t *testing.T, x *X) {
		x.FocusComposer(t)
		x.Type("看一下 @src/co")
		if _, ok := x.Access(NamePicker); !ok {
			t.Fatal("候选列表未显示")
		}
		if !x.Focused(NameComposer) {
			t.Fatal("焦点离开了输入框")
		}
		x.Key(0, ui.KeyEnter)
		chip, ok := x.TextWithPrefix(PrefixFileChip)
		if !ok || !strings.Contains(chip, "src/co") {
			t.Fatalf("没有生成文件附件标签：%q", x.Texts())
		}
		if v := x.Composer(t).Value; strings.Contains(v, "@") {
			t.Fatalf("触发文字未从草稿中去掉：%q", v)
		}
		if x.Count(ui.RoleGroup, NameUser) != 0 {
			t.Fatal("选择候选的 Enter 发送了消息")
		}
	}},
}
