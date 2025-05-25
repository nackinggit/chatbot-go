package config

import "errors"

type AssistantConfig struct {
	ExtractName    *BotConfig `json:"extractName" yaml:"extractName" mapstructure:"extractName"`
	CommentImage   *BotConfig `json:"commentImage" yaml:"commentImage" mapstructure:"commentImage"`
	Chat           *BotConfig `json:"chat" yaml:"chat" mapstructure:"chat"`
	ComicTranslate string     `json:"comicTranslate" yaml:"comicTranslate" mapstructure:"comicTranslate"`
}

func (a *AssistantConfig) Validate() error {
	if a == nil || a.ExtractName == nil || a.CommentImage == nil || a.ComicTranslate == "" {
		return errors.New("config: AssistantConfig is not initialized")
	}
	return nil
}
