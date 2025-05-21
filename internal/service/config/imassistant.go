package config

import "errors"

type AssistantConfig struct {
	ExtractName  *BotConfig `json:"extractName" yaml:"extractName" mapstructure:"extractName"`
	CommentImage *BotConfig `json:"commentImage" yaml:"commentImage" mapstructure:"commentImage"`
}

func (a *AssistantConfig) Validate() error {
	if a == nil || a.ExtractName == nil || a.CommentImage == nil {
		return errors.New("config: AssistantConfig is not initialized")
	}
	return nil
}
