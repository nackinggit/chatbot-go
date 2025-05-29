package config

import "errors"

type AssistantConfig struct {
	ExtractName    *BotConfig `json:"extractName" yaml:"extractName" mapstructure:"extractName"`
	CommentImage   *BotConfig `json:"commentImage" yaml:"commentImage" mapstructure:"commentImage"`
	Chat           *BotConfig `json:"chat" yaml:"chat" mapstructure:"chat"`
	ReasoningChat  *BotConfig `json:"reasoningChat" yaml:"reasoningChat" mapstructure:"reasoningChat"`
	ComicTranslate string     `json:"comicTranslate" yaml:"comicTranslate" mapstructure:"comicTranslate"`
}

func (a *AssistantConfig) Validate() error {
	if a == nil || a.ExtractName == nil || a.CommentImage == nil || a.ComicTranslate == "" || a.Chat == nil {
		return errors.New("config: AssistantConfig is not initialized")
	}
	return nil
}
