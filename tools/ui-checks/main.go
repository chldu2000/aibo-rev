// ui-checks：界面硬性要求（docs/design/ui-requirements.md）的本地验证。在仓库内运行：
//
//	go -C tools/ui-checks run . static        令牌、对比度、原生 UI 源码规则、禁用文案
//	go -C tools/ui-checks run . requirements  要求文档自身的完整性与验收用例映射
//	go -C tools/ui-checks run . selftest      用 selftest/ 中的样例确认各源码规则能命中违规、不误报
//	go -C tools/ui-checks run . uitest        在各验收目标中运行验收套件（go test）
//	go -C tools/ui-checks run . all           依次运行 selftest、static、requirements、uitest
//	go -C tools/ui-checks run . tokens        从 tokens.json 重新生成 tokens.css 与 Go 令牌文件
//
// 任一错误时以非零状态退出。
package main

import (
	"fmt"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

type rootsRule = func(roots []string) *Report

// 源码规则：既检查 goRoots，也参与规则自检。
var sourceRules = []struct {
	name string
	fn   rootsRule
}{
	{"no-hardcoded-color", checkNoHardcodedColor},
	{"min-font-size", checkMinFontSize},
	{"no-left-accent", checkNoLeftAccent},
	{"focus-ring", checkFocusRing},
	{"reduced-motion", checkReducedMotion},
	{"forbidden-copy", checkForbiddenCopy},
}

func staticChecks() (names []string, fns map[string]func() *Report) {
	fns = map[string]func() *Report{
		"tokens-complete": checkTokensComplete,
		"contrast":        checkContrast,
		"token-sync":      checkTokenSync,
	}
	names = []string{"tokens-complete", "contrast", "token-sync"}
	for _, s := range sourceRules {
		fn := s.fn
		fns[s.name] = func() *Report { return fn(cfg.GoRoots) }
		names = append(names, s.name)
	}
	return names, fns
}

func runStatic() bool {
	fmt.Println("\n静态检查")
	names, fns := staticChecks()
	ok := true
	for _, n := range names {
		ok = fns[n]().Print() && ok
	}
	return ok
}

func runRequirements() bool {
	fmt.Println("\n要求文档")
	_, fns := staticChecks()
	return checkRequirements(fns).Print()
}

// runSelftest：selftest/bad 中每条源码规则都必须报错，selftest/good 中不得报错。
func runSelftest() bool {
	fmt.Println("\n规则自检")
	ok := true
	bad, good := []string{"tools/ui-checks/selftest/bad"}, []string{"tools/ui-checks/selftest/good"}
	for _, s := range sourceRules {
		b, g := s.fn(bad), s.fn(good)
		// min-font-size 同时检查令牌，自检只看样例文件中的命中
		be, ge := onlySelftest(b.Errors), onlySelftest(g.Errors)
		pass := len(be) > 0 && len(ge) == 0
		ok = ok && pass
		fmt.Printf("%s %s  — 违规样例命中 %d 处，合规样例误报 %d 处\n", mark(pass), s.name, len(be), len(ge))
		for _, e := range ge {
			fmt.Printf("    ✗ 误报：%s\n", e)
		}
	}
	la := strings.Join(checkNoLeftAccent(bad).Errors, "\n")
	for _, f := range []string{"BorderStartWidth", "BorderWidth", "QuoteBarWidth"} {
		hit := strings.Contains(la, f)
		ok = ok && hit
		fmt.Printf("  %s no-left-accent 识别 %s\n", mark(hit), f)
	}
	return ok
}

func onlySelftest(errs []string) []string {
	var out []string
	for _, e := range errs {
		if strings.Contains(e, "selftest/") {
			out = append(out, e)
		}
	}
	return out
}

func mark(ok bool) string {
	if ok {
		return "✓"
	}
	return "✗"
}

// runUITest 在每个验收目标的目录中运行 go test。
func runUITest(args []string) bool {
	fmt.Println("\n验收测试（uitest）")
	if len(cfg.Targets) == 0 {
		fmt.Println("△ 没有验收目标")
		return true
	}
	ok := true
	for _, t := range cfg.Targets {
		fmt.Printf("— 目标 %s（%s）\n", t.Name, t.Dir)
		cmdArgs := append([]string{"test", "-count=1", "-run", t.Run, "-v"}, args...)
		cmd := exec.Command("go", cmdArgs...)
		cmd.Dir = abs(t.Dir)
		cmd.Env = append(os.Environ(), "CGO_ENABLED=0")
		cmd.Stdout, cmd.Stderr = os.Stdout, os.Stderr
		if err := cmd.Run(); err != nil {
			ok = false
		}
	}
	return ok
}

func main() {
	if err := load(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(2)
	}
	cmd := "all"
	if len(os.Args) > 1 {
		cmd = os.Args[1]
	}
	rest := os.Args[min(2, len(os.Args)):]
	var ok bool
	switch cmd {
	case "static":
		ok = runStatic()
	case "requirements":
		ok = runRequirements()
	case "selftest":
		ok = runSelftest()
	case "uitest":
		ok = runUITest(rest)
	case "all":
		ok = runSelftest()
		ok = runStatic() && ok
		ok = runRequirements() && ok
		ok = runUITest(rest) && ok
	case "tokens":
		if err := generateTokens(); err != nil {
			fmt.Fprintln(os.Stderr, err)
			os.Exit(1)
		}
		ok = true
	default:
		fmt.Fprintf(os.Stderr, "未知命令 %q；可用：all static requirements selftest uitest tokens\n", cmd)
		os.Exit(2)
	}
	if !ok {
		fmt.Println("\n未通过")
		os.Exit(1)
	}
	fmt.Println("\n全部通过")
}

func itoa(n int) string { return strconv.Itoa(n) }
