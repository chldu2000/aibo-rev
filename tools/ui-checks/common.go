package main

import (
	"encoding/json"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Config 是 ui-checks.config.json。路径都相对仓库根。
type Config struct {
	Requirements  string     `json:"requirements"`
	Inventory     string     `json:"inventory"`
	Tokens        string     `json:"tokens"`
	TokensCSS     string     `json:"tokensCss"`
	TokensGo      []TokensGo `json:"tokensGo"`
	GoRoots       []string   `json:"goRoots"`
	CopyRoots     []string   `json:"copyRoots"`
	ForbiddenCopy []string   `json:"forbiddenCopy"`
	Acceptance    string     `json:"acceptance"`
	// RequireAllUITests 为 false 时，声明了 uitest/perf 但验收套件中尚无测试的要求只给出提示。
	// 产品实现成为验收目标后改为 true。
	RequireAllUITests bool     `json:"requireAllUITests"`
	Targets           []Target `json:"targets"`
}

// TokensGo 是一处生成的 Go 令牌文件。
type TokensGo struct {
	Path    string `json:"path"`
	Package string `json:"package"`
}

// Target 是一个验收目标：在 Dir 中运行 go test -run Run。
type Target struct {
	Name string `json:"name"`
	Dir  string `json:"dir"`
	Run  string `json:"run"`
}

var (
	root string
	cfg  Config
)

const configPath = "tools/ui-checks/ui-checks.config.json"

func load() error {
	dir, err := os.Getwd()
	if err != nil {
		return err
	}
	for {
		if _, err := os.Stat(filepath.Join(dir, configPath)); err == nil {
			root = dir
			break
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			return fmt.Errorf("找不到 %s，请在仓库内运行", configPath)
		}
		dir = parent
	}
	raw, err := os.ReadFile(abs(configPath))
	if err != nil {
		return err
	}
	return json.Unmarshal(raw, &cfg)
}

func abs(p string) string { return filepath.Join(root, filepath.FromSlash(p)) }

func rel(p string) string {
	r, err := filepath.Rel(root, p)
	if err != nil {
		return p
	}
	return filepath.ToSlash(r)
}

func read(p string) string {
	b, err := os.ReadFile(abs(p))
	if err != nil {
		return ""
	}
	return string(b)
}

var skipDirs = map[string]bool{"node_modules": true, "dist": true, ".git": true, "bin": true, "reports": true}

// walk 列出 roots 下扩展名匹配的文件（相对仓库根）。
func walk(roots []string, exts ...string) []string {
	var out []string
	for _, r := range roots {
		_ = filepath.WalkDir(abs(r), func(p string, d fs.DirEntry, err error) error {
			if err != nil {
				return nil
			}
			if d.IsDir() {
				if skipDirs[d.Name()] {
					return filepath.SkipDir
				}
				return nil
			}
			for _, e := range exts {
				if strings.HasSuffix(p, e) {
					out = append(out, rel(p))
					break
				}
			}
			return nil
		})
	}
	sort.Strings(out)
	return out
}

// allowed 报告第 line 行（从 1 起）的违规是否已说明：
// 同一行或上一行有 `ui-check-allow: <rule> <理由>`，
// 或处在 `ui-check-allow-begin: <rule|all> <理由>` 与 `ui-check-allow-end` 之间。没有理由不算说明。
func allowed(lines []string, line int, rule string) bool {
	re := regexp.MustCompile(`ui-check-allow:\s*` + regexp.QuoteMeta(rule) + `\s+\S`)
	at := func(i int) string {
		if i >= 1 && i <= len(lines) {
			return lines[i-1]
		}
		return ""
	}
	if re.MatchString(at(line)) || re.MatchString(at(line-1)) {
		return true
	}
	begin := regexp.MustCompile(`ui-check-allow-begin:\s*(` + regexp.QuoteMeta(rule) + `|all)\s+\S`)
	for i := line - 1; i >= 1; i-- {
		if strings.Contains(at(i), "ui-check-allow-end") {
			return false
		}
		if begin.MatchString(at(i)) {
			return true
		}
	}
	return false
}

// Report 是一项检查的结果。
type Report struct {
	Name                    string
	Errors, Warnings, Notes []string
}

func newReport(name string) *Report { return &Report{Name: name} }

func (r *Report) Error(where, format string, a ...any) {
	r.Errors = append(r.Errors, join(where, fmt.Sprintf(format, a...)))
}

func (r *Report) Warn(where, format string, a ...any) {
	r.Warnings = append(r.Warnings, join(where, fmt.Sprintf(format, a...)))
}

func (r *Report) Note(format string, a ...any) { r.Notes = append(r.Notes, fmt.Sprintf(format, a...)) }

func (r *Report) OK() bool { return len(r.Errors) == 0 }

func join(where, msg string) string {
	if where == "" {
		return msg
	}
	return where + "  " + msg
}

func (r *Report) Print() bool {
	mark := "✓"
	if !r.OK() {
		mark = "✗"
	} else if len(r.Warnings) > 0 {
		mark = "△"
	}
	notes := ""
	if len(r.Notes) > 0 {
		notes = "  — " + strings.Join(r.Notes, "；")
	}
	fmt.Printf("%s %s%s\n", mark, r.Name, notes)
	for _, e := range r.Errors {
		fmt.Printf("    ✗ %s\n", e)
	}
	for _, w := range r.Warnings {
		fmt.Printf("    △ %s\n", w)
	}
	return r.OK()
}
