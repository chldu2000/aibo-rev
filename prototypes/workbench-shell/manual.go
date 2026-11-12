package main

import "github.com/egoist/mygo"

// 17 号票的人工夹具入口：原型已有粘贴和拖放，但没有文件选择器入口。
func (app *App) pickAttachmentForValidation() {
	app.openWindow()
	parent := app.win
	go func() {
		paths, err := mygo.Dialog.Open(mygo.OpenDialogOptions{Parent: parent, Title: "选择附件（17 号票验证）", Multiple: true})
		app.update(func() {
			if err != nil {
				app.event("附件文件选择器失败：%v", err)
				return
			}
			for _, path := range paths {
				app.chips = append(app.chips, Chip{Kind: "file", Label: path})
			}
			app.event("附件文件选择器返回 %d 个文件", len(paths))
		})
	}()
}
