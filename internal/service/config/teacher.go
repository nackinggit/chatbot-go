package config

import "errors"

type TeacherConfig struct {
	QuestionAnalyse *BotConfig   `json:"questionAnalyse" yaml:"questionAnalyse" mapstructure:"questionAnalyse"` // 解析题目配置
	AnswerModels    []*BotConfig `json:"answerModels" yaml:"answerModels" mapstructure:"answerModels"`          // 解析答案模型配置
	JudgeModel      *BotConfig   `json:"judgeModel" yaml:"judgeModel" mapstructure:"judgeModel"`                // 判断模型配置
}

func (t *TeacherConfig) Validate() error {
	if t == nil {
		return errors.New("config.teacher is nil")
	}
	if len(t.AnswerModels) == 0 || t.JudgeModel == nil || t.QuestionAnalyse == nil {
		return errors.New("config.teacher is invalid")
	}
	return nil
}
