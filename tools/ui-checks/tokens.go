package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"go/format"
	"math"
	"os"
	"regexp"
	"sort"
	"strconv"
	"strings"
)

// kv 是保持 tokens.json 中原有顺序的对象。
type kv struct {
	keys []string
	vals map[string]any
}

func (o *kv) str(k string) string { s, _ := o.vals[k].(string); return s }
func (o *kv) obj(k string) *kv    { v, _ := o.vals[k].(*kv); return v }

func decodeOrdered(dec *json.Decoder) (any, error) {
	tok, err := dec.Token()
	if err != nil {
		return nil, err
	}
	switch t := tok.(type) {
	case json.Delim:
		switch t {
		case '{':
			o := &kv{vals: map[string]any{}}
			for dec.More() {
				kt, err := dec.Token()
				if err != nil {
					return nil, err
				}
				k := kt.(string)
				v, err := decodeOrdered(dec)
				if err != nil {
					return nil, err
				}
				o.keys = append(o.keys, k)
				o.vals[k] = v
			}
			_, err := dec.Token()
			return o, err
		case '[':
			var a []any
			for dec.More() {
				v, err := decodeOrdered(dec)
				if err != nil {
					return nil, err
				}
				a = append(a, v)
			}
			_, err := dec.Token()
			return a, err
		}
	}
	return tok, nil
}

func loadTokens() (*kv, error) {
	dec := json.NewDecoder(strings.NewReader(read(cfg.Tokens)))
	dec.UseNumber()
	v, err := decodeOrdered(dec)
	if err != nil {
		return nil, fmt.Errorf("%s: %v", cfg.Tokens, err)
	}
	o, ok := v.(*kv)
	if !ok {
		return nil, fmt.Errorf("%s 不是对象", cfg.Tokens)
	}
	return o, nil
}

var (
	hexRe  = regexp.MustCompile(`^#([0-9a-fA-F]{3}|[0-9a-fA-F]{6})$`)
	rgbaRe = regexp.MustCompile(`^rgba?\(\s*(\d+)\s*,\s*(\d+)\s*,\s*(\d+)\s*(?:,\s*([\d.]+)\s*)?\)$`)
)

type rgba struct {
	r, g, b uint8
	a       float64
}

func parseColor(v string) (rgba, bool) {
	s := strings.TrimSpace(v)
	if m := hexRe.FindStringSubmatch(s); m != nil {
		h := m[1]
		if len(h) == 3 {
			h = string([]byte{h[0], h[0], h[1], h[1], h[2], h[2]})
		}
		n, _ := strconv.ParseUint(h, 16, 32)
		return rgba{uint8(n >> 16), uint8(n >> 8), uint8(n), 1}, true
	}
	if m := rgbaRe.FindStringSubmatch(s); m != nil {
		c := rgba{a: 1}
		for i, p := range []*uint8{&c.r, &c.g, &c.b} {
			n, _ := strconv.Atoi(m[i+1])
			*p = uint8(n)
		}
		if m[4] != "" {
			c.a, _ = strconv.ParseFloat(m[4], 64)
		}
		return c, true
	}
	return rgba{}, false
}

func luminance(c rgba) float64 {
	ch := func(v uint8) float64 {
		x := float64(v) / 255
		if x <= 0.03928 {
			return x / 12.92
		}
		return math.Pow((x+0.055)/1.055, 2.4)
	}
	return 0.2126*ch(c.r) + 0.7152*ch(c.g) + 0.0722*ch(c.b)
}

func contrastRatio(a, b rgba) float64 {
	x, y := luminance(a), luminance(b)
	return (math.Max(x, y) + 0.05) / (math.Min(x, y) + 0.05)
}

func checkTokensComplete() *Report {
	r := newReport("tokens-complete")
	t, err := loadTokens()
	if err != nil {
		r.Error("", "%v", err)
		return r
	}
	light, dark := t.obj("color").obj("light"), t.obj("color").obj("dark")
	for _, k := range light.keys {
		if _, ok := dark.vals[k]; !ok {
			r.Error("", "深色缺少令牌 %s", k)
		}
	}
	for _, k := range dark.keys {
		if _, ok := light.vals[k]; !ok {
			r.Error("", "浅色缺少令牌 %s", k)
		}
	}
	r.Note("%d 个颜色令牌", len(light.keys))
	return r
}

func checkContrast() *Report {
	r := newReport("contrast")
	t, err := loadTokens()
	if err != nil {
		r.Error("", "%v", err)
		return r
	}
	pairs, _ := t.obj("contrast").vals["pairs"].([]any)
	n := 0
	for _, theme := range []string{"light", "dark"} {
		colors := t.obj("color").obj(theme)
		for _, p := range pairs {
			pp, _ := p.([]any)
			if len(pp) != 3 {
				r.Error("", "contrast.pairs 中的项应为 [前景, 背景, 对比度]")
				continue
			}
			fg, bg := fmt.Sprint(pp[0]), fmt.Sprint(pp[1])
			min, _ := strconv.ParseFloat(fmt.Sprint(pp[2]), 64)
			a, okA := parseColor(colors.str(fg))
			b, okB := parseColor(colors.str(bg))
			if !okA || !okB || a.a != 1 || b.a != 1 {
				r.Error("", "%s: %s/%s 不是不透明颜色，无法计算", theme, fg, bg)
				continue
			}
			n++
			if v := contrastRatio(a, b); v < min {
				r.Error("", "%s: %s 在 %s 上为 %.2f:1，要求 ≥ %g:1", theme, fg, bg, v, min)
			}
		}
	}
	r.Note("%d 组对比度", n)
	return r
}

// renderTokensCSS 生成 tokens.css，供静态原型与能力包工具面板使用。
func renderTokensCSS(t *kv) string {
	block := func(sel string, o *kv, prefix string) string {
		var b strings.Builder
		fmt.Fprintf(&b, "%s {\n", sel)
		for _, k := range o.keys {
			if strings.HasPrefix(k, "$") {
				continue
			}
			fmt.Fprintf(&b, "  --%s%s: %s;\n", prefix, k, o.str(k))
		}
		b.WriteString("}\n")
		return b.String()
	}
	var misc strings.Builder
	misc.WriteString(":root {\n")
	for _, g := range []string{"font", "space", "radius", "size", "z", "motion", "focus"} {
		o := t.obj(g)
		if o == nil {
			continue
		}
		for _, k := range o.keys {
			if !strings.HasPrefix(k, "$") {
				fmt.Fprintf(&misc, "  --%s-%s: %s;\n", g, k, o.str(k))
			}
		}
	}
	misc.WriteString("}\n")
	return strings.Join([]string{
		"/* 由 tools/ui-checks 从 tokens.json 生成，请勿手改（go -C tools/ui-checks run . tokens）。 */",
		block(`:root, [data-theme="light"]`, t.obj("color").obj("light"), ""),
		block(`[data-theme="dark"]`, t.obj("color").obj("dark"), ""),
		misc.String(),
	}, "\n")
}

func goName(k string) string {
	parts := strings.Split(k, "-")
	for i, p := range parts {
		if p != "" {
			parts[i] = strings.ToUpper(p[:1]) + p[1:]
		}
	}
	return strings.Join(parts, "")
}

func goColor(v string) (string, bool) {
	c, ok := parseColor(v)
	if !ok {
		return "", false // 阴影等复合值不进调色板
	}
	if c.a == 1 && strings.HasPrefix(strings.TrimSpace(v), "#") {
		return fmt.Sprintf("ui.Hex(%q)", strings.TrimSpace(v)), true
	}
	return fmt.Sprintf("ui.RGBA(%d, %d, %d, %s)", c.r, c.g, c.b, strconv.FormatFloat(float64(float32(c.a)), 'g', -1, 32)), true
}

// renderTokensGo 生成 MyGO 原生 UI 使用的 Go 令牌：Palette（Light/Dark）与字号、间距、圆角、尺寸常量。
func renderTokensGo(t *kv, pkg string) ([]byte, error) {
	var b strings.Builder
	fmt.Fprintf(&b, "// Code generated by tools/ui-checks from %s. DO NOT EDIT.\n\npackage %s\n\nimport \"github.com/egoist/mygo/ui\"\n\n", cfg.Tokens, pkg)
	light, dark := t.obj("color").obj("light"), t.obj("color").obj("dark")
	var keys []string
	for _, k := range light.keys {
		if _, ok := goColor(light.str(k)); ok {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)
	b.WriteString("// Palette 是一套主题的颜色令牌。\ntype Palette struct {\n")
	for _, k := range keys {
		fmt.Fprintf(&b, "\t%s ui.Color\n", goName(k))
	}
	b.WriteString("}\n\n")
	for _, th := range []struct {
		name string
		o    *kv
	}{{"light", light}, {"dark", dark}} {
		fmt.Fprintf(&b, "var %s = Palette{\n", goName(th.name))
		for _, k := range keys {
			c, ok := goColor(th.o.str(k))
			if !ok {
				return nil, fmt.Errorf("%s 中 %s 不是颜色", th.name, k)
			}
			fmt.Fprintf(&b, "\t%s: %s,\n", goName(k), c)
		}
		b.WriteString("}\n\n")
	}
	group := func(prefix string, o *kv, skip func(string) bool) {
		ks := append([]string(nil), o.keys...)
		sort.Strings(ks)
		b.WriteString("const (\n")
		for _, k := range ks {
			if strings.HasPrefix(k, "$") || (skip != nil && skip(k)) {
				continue
			}
			fmt.Fprintf(&b, "\t%s%s = %s\n", prefix, goName(k), strings.TrimSuffix(o.str(k), "px"))
		}
		b.WriteString(")\n\n")
	}
	group("Font", t.obj("font"), func(k string) bool { return k == "ui" || k == "mono" })
	group("Space", t.obj("space"), nil)
	group("Radius", t.obj("radius"), nil)
	group("Size", t.obj("size"), nil)
	out, err := format.Source([]byte(b.String()))
	if err != nil {
		return nil, fmt.Errorf("%v\n%s", err, b.String())
	}
	return out, nil
}

// generateTokens 写出 tokens.css 与所有 Go 令牌文件。
func generateTokens() error {
	t, err := loadTokens()
	if err != nil {
		return err
	}
	if err := os.WriteFile(abs(cfg.TokensCSS), []byte(renderTokensCSS(t)), 0o644); err != nil {
		return err
	}
	fmt.Println("已生成", cfg.TokensCSS)
	for _, g := range cfg.TokensGo {
		out, err := renderTokensGo(t, g.Package)
		if err != nil {
			return err
		}
		if err := os.WriteFile(abs(g.Path), out, 0o644); err != nil {
			return err
		}
		fmt.Println("已生成", g.Path)
	}
	return nil
}

func checkTokenSync() *Report {
	r := newReport("token-sync")
	t, err := loadTokens()
	if err != nil {
		r.Error("", "%v", err)
		return r
	}
	hint := "运行 go -C tools/ui-checks run . tokens"
	if got := read(cfg.TokensCSS); got != renderTokensCSS(t) {
		r.Error(cfg.TokensCSS, "与 tokens.json 不一致，%s", hint)
	}
	for _, g := range cfg.TokensGo {
		want, err := renderTokensGo(t, g.Package)
		if err != nil {
			r.Error(g.Path, "%v", err)
			continue
		}
		if got, err := os.ReadFile(abs(g.Path)); err != nil || !bytes.Equal(got, want) {
			r.Error(g.Path, "与 tokens.json 不一致，%s", hint)
		}
	}
	r.Note("核对了 tokens.css 与 %d 个 Go 令牌文件", len(cfg.TokensGo))
	return r
}
