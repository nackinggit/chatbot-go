package model

type QuestionAnalyseRequest struct {
	ImageUrl string `json:"imageUrl" binding:"required" err:"imageUrl is required"` // 图片url
}

type QuestionAnalyseStreamChunk struct {
	StreamMessage
}

type QAStreamChunk struct {
	StreamMessage
	Name       string           `json:"name"`       // 模型名称
	AllAnswers []*StreamMessage `json:"allAnswers"` // 多个模型返回的答案
	AllEndflag bool             `json:"allEndflag"` // 多个模型返回的结束标志
}

type QARequest struct {
	Question string   `json:"question" binding:"required" err:"question is required"` // 题目
	Models   []string `json:"modelNames"`                                             // 模型名称
}
