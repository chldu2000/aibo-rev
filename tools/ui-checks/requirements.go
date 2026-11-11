package main

import (
	"regexp"
	"strings"
)

var (
	levels   = []string{"必须", "不得", "建议"}
	fields   = []string{"条款", "依据", "Checklist", "验收", "验证"}
	headRe   = regexp.MustCompile(`^#### (UI-[A-Z0-9]+-\d{2}) · (\S+) · (.+)$`)
	fieldRe  = regexp.MustCompile(`^- ([^：]+)：(.*)$`)
	caseRe   = regexp.MustCompile(`ID:\s*"(UI-[A-Z0-9]+-\d{2})"`)
	splitRe  = regexp.MustCompile(`[、,，\s]+`)
	checkedR = regexp.MustCompile(`(?m)^- \[x\] (F\d{2}\.\d+|D\d+(?:\.\d+)?)`)
	allR     = regexp.MustCompile(`(?m)^- \[[ x]\] (F\d{2}\.\d+|D\d+(?:\.\d+)?)`)
)

type requirement struct {
	id, level, title string
	line             int
	fields           map[string]string
}

func parseRequirements(text string) []*requirement {
	var reqs []*requirement
	var cur *requirement
	for i, line := range strings.Split(text, "\n") {
		if m := headRe.FindStringSubmatch(line); m != nil {
			cur = &requirement{id: m[1], level: m[2], title: m[3], line: i + 1, fields: map[string]string{}}
			reqs = append(reqs, cur)
			continue
		}
		if strings.HasPrefix(line, "#") {
			cur = nil
			continue
		}
		if m := fieldRe.FindStringSubmatch(line); cur != nil && m != nil {
			cur.fields[strings.TrimSpace(m[1])] = strings.TrimSpace(m[2])
		}
	}
	return reqs
}

// acceptanceCases 返回验收套件中登记的要求编号。
func acceptanceCases() map[string]bool {
	out := map[string]bool{}
	for _, p := range walk([]string{cfg.Acceptance}, ".go") {
		for _, m := range caseRe.FindAllStringSubmatch(read(p), -1) {
			out[m[1]] = true
		}
	}
	return out
}

func set(re *regexp.Regexp, s string) map[string]bool {
	out := map[string]bool{}
	for _, m := range re.FindAllStringSubmatch(s, -1) {
		out[m[1]] = true
	}
	return out
}

func checkRequirements(static map[string]func() *Report) *Report {
	r := newReport("requirements")
	reqs := parseRequirements(read(cfg.Requirements))
	inv := read(cfg.Inventory)
	checked, all := set(checkedR, inv), set(allR, inv)
	cases := acceptanceCases()
	seen := map[string]int{}
	referenced := map[string]bool{}
	counts := map[string]int{}
	var pending []string
	for _, q := range reqs {
		at := cfg.Requirements + ":" + itoa(q.line) + " " + q.id
		if l, ok := seen[q.id]; ok {
			r.Error(at, "编号重复（首次在第 %d 行）", l)
		}
		seen[q.id] = q.line
		if !contains(levels, q.level) {
			r.Error(at, "等级“%s”不是 %s", q.level, strings.Join(levels, "/"))
		} else {
			counts[q.level]++
		}
		for _, f := range fields {
			if q.fields[f] == "" {
				r.Error(at, "缺少字段“%s”", f)
			}
		}
		for _, ref := range splitRe.Split(q.fields["Checklist"], -1) {
			if ref == "" {
				continue
			}
			switch {
			case !all[ref]:
				r.Error(at, "Checklist 中没有 %s", ref)
			case !checked[ref]:
				r.Error(at, "%s 未勾选，不能作为必需功能的依据", ref)
			}
			referenced[ref] = true
		}
		auto := false
		for _, m := range splitRe.Split(q.fields["验证"], -1) {
			switch {
			case m == "" || m == "manual":
			case m == "uitest" || m == "perf":
				auto = true
				if !cases[q.id] {
					if cfg.RequireAllUITests {
						r.Error(at, "声明了 %s，但 %s 中没有 ID 为 %s 的验收用例", m, cfg.Acceptance, q.id)
					} else {
						pending = append(pending, q.id)
					}
				}
			case strings.HasPrefix(m, "static:"):
				auto = true
				if static[strings.TrimPrefix(m, "static:")] == nil {
					r.Error(at, "未知的静态检查 %s", strings.TrimPrefix(m, "static:"))
				}
			default:
				r.Error(at, "未知的验证方式“%s”", m)
			}
		}
		if auto {
			counts["auto"]++
		} else {
			counts["manual"]++
		}
	}
	for id := range cases {
		if _, ok := seen[id]; !ok {
			r.Error(cfg.Acceptance, "验收用例引用了不存在的要求 %s", id)
		}
	}
	if len(pending) > 0 {
		r.Warn("", "%d 条要求的验收用例尚未编写（产品实现成为验收目标时补齐，并把 requireAllUITests 改为 true）：%s", len(pending), strings.Join(pending, "、"))
	}
	var uncovered []string
	for _, m := range checkedR.FindAllStringSubmatch(inv, -1) {
		if !referenced[m[1]] {
			uncovered = append(uncovered, m[1])
		}
	}
	if len(uncovered) > 0 {
		r.Warn("", "以下已勾选的 Checklist 项尚无界面要求覆盖（可能不涉及界面，或待补）：%s", strings.Join(uncovered, "、"))
	}
	r.Note("%d 条要求：必须 %d、不得 %d、建议 %d；可自动验证 %d，仅人工 %d；验收用例 %d；覆盖已勾选 Checklist %d/%d",
		len(reqs), counts["必须"], counts["不得"], counts["建议"], counts["auto"], counts["manual"], len(cases), len(referenced), len(checked))
	return r
}

func contains(s []string, v string) bool {
	for _, x := range s {
		if x == v {
			return true
		}
	}
	return false
}
