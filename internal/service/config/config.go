package config

import "errors"

type ServiceConfig struct {
	Teacher *TeacherConfig `json:"teacher" yaml:"teacher" mapstructure:"teacher"` // 解题助手配置
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

type BotConfig struct {
	Model    string `json:"model" yaml:"model" mapstructure:"model"`          // 模型名称
	Name     string `json:"name" yaml:"name" mapstructure:"name"`             // 机器人名称
	ModelKey string `json:"modelKey" yaml:"modelKey" mapstructure:"modelKey"` // 模型标识
	Api      string `json:"api" yaml:"api" mapstructure:"api"`                // 模型API
}
