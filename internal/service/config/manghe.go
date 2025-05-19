package config

import "errors"

type MangHeConfig struct {
	ImageAnalyse *BotConfig `json:"imageAnalyse" yaml:"imageAnalyse" mapstructure:"imageAnalyse"` // 图片分析
	Predict      *BotConfig `json:"predict" yaml:"predict" mapstructure:"predict"`                // 预测
}

func (t *MangHeConfig) Validate() error {
	if t == nil {
		return errors.New("config.manghe is nil")
	}
	if t.ImageAnalyse == nil || t.Predict == nil {
		return errors.New("config.manghe is invalid")
	}
	return nil
}
