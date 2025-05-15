package config

import "time"

type LLMConfig struct {
	RegisterService   string           `json:"registerService" yaml:"registerService" mapstructure:"registerService"`
	Name              string           `json:"name" yaml:"name" mapstructure:"name"`
	BaseUrl           string           `json:"baseUrl" yaml:"baseUrl" mapstructure:"baseUrl"`
	ApiKey            string           `json:"apiKey" yaml:"apiKey" mapstructure:"apiKey"`
	Timeout           time.Duration    `json:"timeout" yaml:"timeout" mapstructure:"timeout"`
	MaxRetries        int              `json:"maxRetries" yaml:"maxRetries" mapstructure:"maxRetries"`
	Models            []LlmModelConfig `json:"models" yaml:"models" mapstructure:"models"`
	OpenaiCompatiable bool             `json:"openaiCompatiable" yaml:"openaiCompatiable" mapstructure:"openaiCompatiable"`
}
type LlmModelConfig struct {
	Name  string `json:"name" yaml:"name" mapstructure:"name"`
	Model string `json:"model" yaml:"model" mapstructure:"model"`
}
