package model

type StreamMessage struct {
	Reasoning string `json:"reasoning"` // 推理
	Content   string `json:"content"`   // 内容
	Endflag   bool   `json:"endflag"`   // 结束标志
	Exception string `json:"exception"` // 错误信息
}
