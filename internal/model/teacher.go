package model

type QuestionAnalyseRequest struct {
	ImageUrl string `json:"imgUrl" binding:"required" err:"imageUrl is required"` // 图片url
}

type QuestionAnalyseStreamChunk struct {
	StreamMessage
}

type QARequest struct {
	Question string   `json:"question" binding:"required" err:"question is required"` // 题目
	Models   []string `json:"modelNames"`                                             // 模型名称
}

type QAStreamChunk struct {
	*StreamMessage
	Name       *string          `json:"name,omitempty"`       // 模型名称
	Model      *string          `json:"model,omitempty"`      // 模型
	AllAnswers []*QAStreamChunk `json:"allAnswers,omitempty"` // 多个模型返回的答案
	AllEndflag bool             `json:"allEndflag,omitempty"` // 多个模型返回的结束标志
}

type JudgeAnswerRequest struct {
	Question string   `json:"question" binding:"required" err:"question is required"`   // 题目
	Answers  []Answer `json:"answers" binding:"required,dive" err:"answer is required"` // 模型返回的答案
}

type Answer struct {
	Model   string `json:"model" binding:"required" err:"model is required"`     // 模型名称
	Name    string `json:"name" binding:"required" err:"name is required"`       // 模型
	Content string `json:"content" binding:"required" err:"content is required"` // 模型返回的答案内容
}
