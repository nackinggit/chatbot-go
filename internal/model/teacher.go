package model

import (
	"fmt"
	"strings"
)

type ImageRequest struct {
	ImageUrl string `json:"imgUrl" binding:"required" err:"imgUrl is required"` // 图片url
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

type MangHePredictRequest struct {
	Time        string   `json:"time"`                                                         // 时间
	Addr        string   `json:"addr"`                                                         // 地址
	Direction   string   `json:"direction"`                                                    // 方向
	Series      string   `json:"series"`                                                       // 系列
	Roles       []string `json:"roles"`                                                        // 角色
	GoalRole    string   `json:"goalRole" binding:"required" err:"goalRole is required"`       // 目标角色
	Description string   `json:"description" binding:"required" err:"description is required"` // 描述
}

func (mpr *MangHePredictRequest) ToString() string {
	res := ""
	if mpr.Time != "" {
		res += fmt.Sprintf("【当前时间】%s\n", mpr.Time)
	}
	if mpr.Addr != "" {
		res += fmt.Sprintf("【所处地址】%s\n", mpr.Addr)
	}
	if mpr.Direction != "" {
		res += fmt.Sprintf("【朝向】%s\n", mpr.Direction)
	}
	if mpr.Series != "" {
		res += fmt.Sprintf("【盲盒系列】%s\n", mpr.Series)
	}
	if len(mpr.Roles) > 0 {
		res += fmt.Sprintf("【该系列主要角色】%s\n", strings.Join(mpr.Roles, ", "))
	}
	res += fmt.Sprintf("【目标款式】%s\n", mpr.GoalRole)
	res += fmt.Sprintf("【描述】%s", mpr.Description)
	return res
}
