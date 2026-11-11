package main

import (
	"go/ast"
	"go/constant"
	"go/parser"
	"go/token"
	"strconv"
	"strings"
)

// 原生 UI 源码（Go）的静态规则。只看 goRoots 下的非测试、非生成文件。

type goFile struct {
	path  string
	fset  *token.FileSet
	file  *ast.File
	lines []string
}

func parseGoFiles(roots []string) ([]goFile, []string) {
	var out []goFile
	var errs []string
	for _, p := range walk(roots, ".go") {
		if strings.HasSuffix(p, "_test.go") {
			continue
		}
		src := read(p)
		if strings.Contains(src, "// Code generated ") && strings.Contains(src, "DO NOT EDIT.") {
			continue
		}
		fset := token.NewFileSet()
		f, err := parser.ParseFile(fset, p, src, parser.ParseComments|parser.SkipObjectResolution)
		if err != nil {
			errs = append(errs, err.Error())
			continue
		}
		out = append(out, goFile{p, fset, f, strings.Split(src, "\n")})
	}
	return out, errs
}

func (g goFile) line(n ast.Node) int { return g.fset.Position(n.Pos()).Line }

func (g goFile) where(n ast.Node) string { return g.path + ":" + strconv.Itoa(g.line(n)) }

// number 返回字面量（可带一元负号）表达式的数值。
func number(e ast.Expr) (float64, bool) {
	switch x := e.(type) {
	case *ast.BasicLit:
		if x.Kind != token.INT && x.Kind != token.FLOAT {
			return 0, false
		}
		v, ok := constant.Float64Val(constant.MakeFromLiteral(x.Value, x.Kind, 0))
		return v, ok
	case *ast.ParenExpr:
		return number(x.X)
	case *ast.UnaryExpr:
		if x.Op == token.SUB {
			v, ok := number(x.X)
			return -v, ok
		}
	}
	return 0, false
}

// method 返回 x.Name(...) 调用的方法名与接收者表达式。
func method(c *ast.CallExpr) (string, ast.Expr) {
	if s, ok := c.Fun.(*ast.SelectorExpr); ok {
		return s.Sel.Name, s.X
	}
	return "", nil
}

func isUI(e ast.Expr) bool { id, ok := e.(*ast.Ident); return ok && id.Name == "ui" }

func exprText(g goFile, e ast.Expr) string {
	s, end := g.fset.Position(e.Pos()), g.fset.Position(e.End())
	if s.Line != end.Line || s.Line > len(g.lines) {
		return ""
	}
	return g.lines[s.Line-1][s.Column-1 : end.Column-1]
}

// goRule 对每个文件逐节点检查，report 只在违规未被 ui-check-allow 说明时记录。
func goRule(name, allow string, roots []string, visit func(g goFile, n ast.Node, report func(n ast.Node, format string, a ...any))) *Report {
	r := newReport(name)
	files, errs := parseGoFiles(roots)
	for _, e := range errs {
		r.Error("", "无法解析：%s", e)
	}
	for _, g := range files {
		ast.Inspect(g.file, func(n ast.Node) bool {
			if n == nil {
				return false
			}
			visit(g, n, func(at ast.Node, format string, a ...any) {
				if !allowed(g.lines, g.line(at), allow) {
					r.Error(g.where(at), format, a...)
				}
			})
			return true
		})
	}
	r.Note("扫描 %d 个 Go 文件", len(files))
	return r
}

// no-hardcoded-color：颜色只来自生成的令牌（UI-VIS-01）。
func checkNoHardcodedColor(roots []string) *Report {
	return goRule("no-hardcoded-color", "color", roots, func(g goFile, n ast.Node, report func(ast.Node, string, ...any)) {
		switch x := n.(type) {
		case *ast.CallExpr:
			if name, recv := method(x); isUI(recv) && (name == "RGB" || name == "RGBA" || name == "Hex") {
				report(x, "硬写颜色 ui.%s(...)，应使用令牌（UI-VIS-01）", name)
			}
		case *ast.CompositeLit:
			if s, ok := x.Type.(*ast.SelectorExpr); ok && isUI(s.X) && s.Sel.Name == "Color" {
				report(x, "硬写颜色 ui.Color{...}，应使用令牌（UI-VIS-01）")
			}
		}
	})
}

// min-font-size：字号不小于 12（UI-VIS-04）。源码中的字面量与令牌中的 font.size-* 都检查。
func checkMinFontSize(roots []string) *Report {
	r := goRule("min-font-size", "font-size", roots, func(g goFile, n ast.Node, report func(ast.Node, string, ...any)) {
		var arg ast.Expr
		switch x := n.(type) {
		case *ast.CallExpr:
			if name, _ := method(x); name == "FontSize" && len(x.Args) == 1 {
				arg = x.Args[0]
			}
		case *ast.KeyValueExpr:
			if k, ok := x.Key.(*ast.Ident); ok && strings.HasSuffix(k.Name, "FontSize") {
				arg = x.Value
			}
		}
		if v, ok := number(arg); arg != nil && ok && v < 12 {
			report(arg, "字号 %g 小于 12（UI-VIS-04）", v)
		}
	})
	if t, err := loadTokens(); err == nil && t.obj("font") != nil {
		f := t.obj("font")
		for _, k := range f.keys {
			if !strings.HasPrefix(k, "size-") {
				continue
			}
			if v, err := strconv.ParseFloat(strings.TrimSuffix(f.str(k), "px"), 64); err == nil && v < 12 {
				r.Error(cfg.Tokens, "令牌 font.%s 为 %s，小于 12px（UI-VIS-04）", k, f.str(k))
			}
		}
	}
	return r
}

// no-left-accent：不用一侧加粗的边条表达状态或选中（UI-VIS-05）。
func checkNoLeftAccent(roots []string) *Report {
	return goRule("no-left-accent", "left-accent", roots, func(g goFile, n ast.Node, report func(ast.Node, string, ...any)) {
		switch x := n.(type) {
		case *ast.CallExpr:
			name, _ := method(x)
			switch {
			case name == "BorderStartWidth" && len(x.Args) == 1:
				if v, ok := number(x.Args[0]); !ok || v > 1 {
					report(x, "BorderStartWidth(%s) 形成左侧信号条（UI-VIS-05）", exprText(g, x.Args[0]))
				}
			case name == "BorderWidth" && len(x.Args) == 4:
				left, right := x.Args[3], x.Args[1]
				lv, lok := number(left)
				if lok && lv <= 1 {
					return
				}
				if rv, rok := number(right); (lok && rok && lv == rv) || (!lok && exprText(g, left) == exprText(g, right)) {
					return
				}
				report(x, "BorderWidth 的左边 %s 宽于 1 且与右边不同，形成左侧信号条（UI-VIS-05）", exprText(g, left))
			}
		case *ast.KeyValueExpr:
			if k, ok := x.Key.(*ast.Ident); ok && k.Name == "QuoteBarWidth" {
				if v, ok := number(x.Value); !ok || v > 1 {
					report(x, "QuoteBarWidth = %s 形成引用块左侧竖条，应设为 0（UI-VIS-05）", exprText(g, x.Value))
				}
			}
		case *ast.AssignStmt:
			for i, lhs := range x.Lhs {
				if s, ok := lhs.(*ast.SelectorExpr); ok && s.Sel.Name == "QuoteBarWidth" && i < len(x.Rhs) {
					if v, ok := number(x.Rhs[i]); !ok || v > 1 {
						report(x, "QuoteBarWidth = %s 形成引用块左侧竖条，应设为 0（UI-VIS-05）", exprText(g, x.Rhs[i]))
					}
				}
			}
		}
	})
}

// focus-ring：不得关掉焦点环（UI-VIS-06）。
func checkFocusRing(roots []string) *Report {
	return goRule("focus-ring", "focus-ring", roots, func(g goFile, n ast.Node, report func(ast.Node, string, ...any)) {
		if x, ok := n.(*ast.CallExpr); ok {
			if name, _ := method(x); name == "FocusRing" && len(x.Args) == 1 {
				if id, ok := x.Args[0].(*ast.Ident); !ok || id.Name != "true" {
					report(x, "FocusRing(%s) 可能关掉焦点环，须注明替代样式（UI-VIS-06）", exprText(g, x.Args[0]))
				}
			}
		}
		if x, ok := n.(*ast.KeyValueExpr); ok {
			if k, ok := x.Key.(*ast.Ident); ok && k.Name == "FocusRingWidth" {
				if v, ok := number(x.Value); ok && v > 0 && v < 2 {
					report(x, "FocusRingWidth %g 小于 2（UI-VIS-06）", v)
				}
			}
		}
	})
}

// reduced-motion：动画遵循系统“减少动态效果”（UI-VIS-08）。NonMotion 会绕过它，须说明只用于不产生位移的效果。
func checkReducedMotion(roots []string) *Report {
	return goRule("reduced-motion", "motion", roots, func(g goFile, n ast.Node, report func(ast.Node, string, ...any)) {
		if x, ok := n.(*ast.KeyValueExpr); ok {
			if k, ok := x.Key.(*ast.Ident); ok && k.Name == "NonMotion" {
				if v, ok := x.Value.(*ast.Ident); !ok || v.Name != "false" {
					report(x, "NonMotion 让动画在“减少动态效果”下照常播放，须注明理由（UI-VIS-08）")
				}
			}
		}
	})
}

// forbidden-copy：Go 字符串字面量与语言资源中不得出现禁用文案（UI-GEN-03、UI-COMP-08）。
func checkForbiddenCopy(roots []string) *Report {
	r := newReport("forbidden-copy")
	files, errs := parseGoFiles(roots)
	for _, e := range errs {
		r.Error("", "无法解析：%s", e)
	}
	hit := func(where, s string) {
		for _, p := range cfg.ForbiddenCopy {
			if strings.Contains(s, p) {
				r.Error(where, "出现禁用文案“%s”（UI-GEN-03 / UI-COMP-08）", p)
			}
		}
	}
	for _, g := range files {
		ast.Inspect(g.file, func(n ast.Node) bool {
			if lit, ok := n.(*ast.BasicLit); ok && lit.Kind == token.STRING {
				if s, err := strconv.Unquote(lit.Value); err == nil {
					hit(g.where(lit), s)
				}
			}
			return true
		})
	}
	res := walk(roots, ".json", ".toml")
	for _, p := range res {
		for i, line := range strings.Split(read(p), "\n") {
			hit(p+":"+strconv.Itoa(i+1), line)
		}
	}
	r.Note("扫描 %d 个 Go 文件、%d 个资源文件", len(files), len(res))
	return r
}
