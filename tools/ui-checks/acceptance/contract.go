package acceptance

// 测试契约（要求文档 §11.1）：中文界面下的无障碍名称。只用于定位，不限制布局与外观。
const (
	NameStream     = "会话内容"     // 会话流
	NameUser       = "用户消息"     // 用户消息（RoleGroup）
	NameAgent      = "Agent 回复" // Agent 回复（RoleGroup）
	NameComposer   = "消息"       // 输入框（RoleTextField）
	NamePicker     = "候选"       // @ # / 候选列表
	NameSend       = "发送"       // 主按钮（空闲时）
	NameQueue      = "排队"       // 主按钮（执行中）
	PrefixFileChip = "文件 · "    // 文件附件标签的可见文字前缀
)
