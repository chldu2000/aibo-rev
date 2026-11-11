package main

// 一次性实验：模型 × 推理强度选择。
//
// Agent 提供完整的“模型 → 支持的推理强度”映射时，以矩阵呈现，一次选定一个组合；否则退回两个下拉，
// 推理强度只列出当前模型支持的选项（这是标准 ACP 的形态：codex-acp 2.6.0 把 model 与
// reasoning_effort 作为两个 select 配置项公布，只给出当前模型的推理强度；完整映射只在适配器内部）。
//
// 示例数据是虚构的，只用于呈现。

import (
	"fmt"
	"slices"

	"github.com/egoist/mygo/plugins/glass"
	"github.com/egoist/mygo/ui"
)

type modelCap struct {
	Name    string
	Efforts []string // 空：该模型不提供推理强度
	Default string
}

// catalog 是一个会话从 Agent 得到的模型目录。FullMap 表示每个模型支持的推理强度都已知。
type catalog struct {
	Agent   string
	FullMap bool
	Models  []modelCap
}

// effortAxis 是推理强度的固定顺序；矩阵的列取目录中出现过的强度。
var effortAxis = []string{"最低", "低", "中", "高", "超高"}

var (
	catalogMatrix = catalog{Agent: "Claude Code", FullMap: true, Models: []modelCap{
		{Name: "Opus 5.5", Efforts: []string{"低", "中", "高", "超高"}, Default: "高"},
		{Name: "Sonnet 5.5", Efforts: []string{"低", "中", "高"}, Default: "中"},
		{Name: "Haiku 5.5", Efforts: nil},
	}}
	catalogPerModel = catalog{Agent: "Codex", FullMap: false, Models: []modelCap{
		{Name: "示例模型 Large", Efforts: []string{"最低", "低", "中", "高"}, Default: "中"},
		{Name: "示例模型 Mini", Efforts: []string{"低", "中"}, Default: "低"},
	}}
)

// catalogFor：会话 0、1 用提供完整映射的 Agent，会话 2、3 用只提供当前模型推理强度的 Agent。
func catalogFor(session int) *catalog {
	if session >= 2 {
		return &catalogPerModel
	}
	return &catalogMatrix
}

func (app *App) catalog() *catalog { return catalogFor(app.sh.cur) }

func (cat *catalog) model(name string) *modelCap {
	for i := range cat.Models {
		if cat.Models[i].Name == name {
			return &cat.Models[i]
		}
	}
	return &cat.Models[0]
}

func (cat *catalog) columns() []string {
	var cols []string
	for _, e := range effortAxis {
		for _, m := range cat.Models {
			if slices.Contains(m.Efforts, e) {
				cols = append(cols, e)
				break
			}
		}
	}
	return cols
}

// defaults 返回会话的初始模型与推理强度。
func (cat *catalog) defaults() (string, string) {
	m := cat.Models[0]
	return m.Name, m.Default
}

// setModel：只换模型时（两个下拉的形态），当前推理强度不受支持就改用该模型的默认值，与 codex-acp 的行为一致。
func (app *App) setModel(name string) {
	m := app.catalog().model(name)
	app.model = m.Name
	if !slices.Contains(m.Efforts, app.effort) {
		app.effort = m.Default
	}
	app.changeModel()
}

func (app *App) pickerLabel() string {
	m := app.catalog().model(app.model)
	if len(m.Efforts) == 0 {
		return m.Name
	}
	return m.Name + " · " + app.effort
}

// modelPicker 在输入面板的设置行中呈现模型与推理强度选择。
func (app *App) modelPicker(c *ui.Context, p Palette) {
	cat := app.catalog()
	if !cat.FullMap {
		names := make([]string, len(cat.Models))
		for i, m := range cat.Models {
			names[i] = m.Name
		}
		model := app.model
		if app.kselect(c, p, &model, names, "模型") {
			app.setModel(model)
		}
		if m := cat.model(app.model); len(m.Efforts) > 0 {
			prev := app.effort
			if app.kselect(c, p, &app.effort, m.Efforts, "推理强度") && app.effort != prev {
				app.changeModel()
			}
		} else {
			app.unsupportedSetting(c, p, "推理强度", "当前模型不提供推理强度")
		}
		return
	}
	app.matrixPicker(c, p, cat)
}

type matrixState struct {
	open      bool
	row, col  int  // 键盘位置
	focusCell bool // 下一帧把焦点交给 (row, col)
	wasOpen   bool
}

func (app *App) matrixPicker(c *ui.Context, p Palette, cat *catalog) {
	st := &app.sh.matrix
	cols := cat.columns()
	tr := ui.ButtonBase(c.Key("matrix-trigger")).Label("模型与推理强度："+app.pickerLabel()).Height(SizeControl).Padding(0, 10, 0, 12).Gap(4).
		AlignItems(ui.Center).Radius(RadiusPill).Material(glass.Glass{Interactive: true})
	tr.Children(func() {
		ui.Text(c, app.pickerLabel()).FontSize(FontSizeSmall + 1)
		ui.Icon(c, icChevron).Size(12, 12)
	})
	if tr.Clicked() {
		st.open = !st.open
	}
	if st.open && !st.wasOpen {
		// 打开时键盘位置落在当前选择上
		for r, m := range cat.Models {
			if m.Name == app.model {
				st.row = r
				st.col = max(0, slices.Index(cols, app.effort))
			}
		}
		st.focusCell = true
	}
	st.wasOpen = st.open

	ui.PopoverBase(c, tr, &st.open, func(panel ui.Element) {
		panel.AttachTo(tr, ui.AnchorTopLeft, ui.AnchorBottomLeft).Margin(0, 0, 8, 0)
		app.glassPanel(panel)
		panel.Padding(Space2).Gap(2).Role(ui.RoleGrid).Label("模型与推理强度")
		app.matrixKeys(panel, cat, cols)
		const nameW, cellW, cellH = 128, 52, 30
		// 列标题
		ui.Row(c).Role(ui.RoleRow).Gap(4).Children(func() {
			ui.Box(c).Width(nameW).AccessHidden(true)
			for i, e := range cols {
				ui.Text(c, e).Width(cellW).TextAlign(ui.Center).FontSize(FontSizeSmall).TextColor(p.Muted).Role(ui.RoleColumnHeader).ColumnIndex(i + 1)
			}
		})
		for r, m := range cat.Models {
			ui.Row(c.Key("row:" + m.Name)).Role(ui.RoleRow).RowIndex(r + 1).Gap(4).AlignItems(ui.Center).Children(func() {
				ui.Text(c, m.Name).Width(nameW).Padding(0, 6).Role(ui.RoleRowHeader).TextColor(map[bool]ui.Color{true: p.Text, false: p.Muted}[m.Name == app.model])
				if len(m.Efforts) == 0 {
					// 不提供推理强度的模型：整行一个格子
					w := float32(len(cols))*cellW + float32(len(cols)-1)*4
					app.matrixCell(c, p, r, 0, m.Name, "", "不使用推理强度", w, cellH)
					return
				}
				for ci, e := range cols {
					if !slices.Contains(m.Efforts, e) {
						ui.Text(c, "—").Width(cellW).TextAlign(ui.Center).TextColor(p.Muted.Alpha(0.6)).Role(ui.RoleCell).Label(m.Name + " 不支持 " + e).ColumnIndex(ci + 1)
						continue
					}
					label := ""
					if e == m.Default {
						label = "默认"
					}
					app.matrixCell(c, p, r, ci, m.Name, e, label, cellW, cellH)
				}
			})
		}
		ui.Text(c, "由 "+cat.Agent+" 提供的模型与推理强度").FontSize(FontSizeSmall).TextColor(p.Muted).Padding(Space1, 6, 0, 6)
	})
}

// matrixCell：可选的组合。选中为强调色实心；键盘位置为强调色描边；“默认”以小字标出。
func (app *App) matrixCell(c *ui.Context, p Palette, r, ci int, model, effort, note string, w, h float32) {
	st := &app.sh.matrix
	selected := app.model == model && (effort == "" || app.effort == effort)
	name := model
	if effort != "" {
		name += " · " + effort
	}
	cell := ui.ButtonBase(c.Key(fmt.Sprint("cell:", model, effort))).Label(name).Role(ui.RoleCell).Selected(selected).ColumnIndex(ci+1).
		Size(w, h).Center().Radius(10)
	if note != "" {
		cell.Description(note)
	}
	hi := st.row == r && st.col == ci
	switch {
	case selected:
		cell.Background(p.Accent).TextColor(p.OnAccent)
	case cell.Hovered():
		cell.Background(p.Accent.Alpha(0.14))
	default:
		cell.Background(p.Hover)
	}
	if hi && !selected {
		cell.Border(1.5, p.Accent)
	}
	if hi && st.focusCell {
		cell.Focus()
		st.focusCell = false
	}
	cell.Children(func() {
		switch {
		case selected:
			ui.Icon(c, icCheck).Size(13, 13)
		case note != "":
			ui.Text(c, note).FontSize(FontSizeSmall).TextColor(p.Muted)
		default:
			ui.Box(c).Size(6, 6).Radius(3).Background(p.Muted.Alpha(0.6)).AccessHidden(true)
		}
	})
	if cell.Clicked() {
		st.row, st.col = r, ci
		app.chooseCell(model, effort)
	}
}

func (app *App) chooseCell(model, effort string) {
	if app.model != model || (effort != "" && app.effort != effort) {
		app.model = model
		if effort != "" {
			app.effort = effort
		}
		app.changeModel()
	}
	app.sh.matrix.open = false
}

// matrixKeys：方向键在可选格子之间移动（跳过不支持的组合），Enter/Space 选择，Escape 由 PopoverBase 关闭。
func (app *App) matrixKeys(panel ui.Element, cat *catalog, cols []string) {
	st := &app.sh.matrix
	ok := func(r, ci int) bool {
		m := cat.Models[r]
		if len(m.Efforts) == 0 {
			return ci == 0
		}
		return slices.Contains(m.Efforts, cols[ci])
	}
	move := func(dr, dc int) {
		r, ci := st.row, st.col
		for i := 0; i < len(cat.Models)*len(cols); i++ {
			if dc != 0 {
				ci += dc
				if ci < 0 || ci >= len(cols) {
					return
				}
			} else {
				r += dr
				if r < 0 || r >= len(cat.Models) {
					return
				}
				// 换行时落在同一列，或该行最近的可选列
				if !ok(r, ci) {
					best := -1
					for d := 0; d < len(cols); d++ {
						if ci-d >= 0 && ok(r, ci-d) {
							best = ci - d
							break
						}
						if ci+d < len(cols) && ok(r, ci+d) {
							best = ci + d
							break
						}
					}
					if best < 0 {
						continue
					}
					st.row, st.col, st.focusCell = r, best, true
					return
				}
			}
			if ok(r, ci) {
				st.row, st.col, st.focusCell = r, ci, true
				return
			}
		}
	}
	switch {
	case panel.Shortcut(0, ui.KeyUp):
		move(-1, 0)
	case panel.Shortcut(0, ui.KeyDown):
		move(1, 0)
	case panel.Shortcut(0, ui.KeyLeft):
		move(0, -1)
	case panel.Shortcut(0, ui.KeyRight):
		move(0, 1)
	}
}

// unsupportedSetting：Agent 不支持的执行设置保留为禁用控件并说明原因，不隐藏（UI-COMP-10）。
func (app *App) unsupportedSetting(c *ui.Context, p Palette, label, reason string) {
	b := ui.ButtonBase(c.Key("unsupported:"+label)).Label(label).AccessDisabled(true).Description(reason).Tooltip(reason).
		Height(SizeControl).Padding(0, 12).Gap(4).AlignItems(ui.Center).Radius(RadiusPill).Border(1, p.Border).TextColor(p.Muted)
	b.Children(func() { ui.Text(c, label+" —").FontSize(FontSizeSmall + 1) })
}
