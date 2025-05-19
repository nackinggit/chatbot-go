package config

import "time"

type LLMConfig struct {
	RegisterService   string        `json:"registerService" yaml:"registerService" mapstructure:"registerService"`
	Name              string        `json:"name" yaml:"name" mapstructure:"name"`
	BaseUrl           string        `json:"baseUrl" yaml:"baseUrl" mapstructure:"baseUrl"`
	ApiKey            string        `json:"apiKey" yaml:"apiKey" mapstructure:"apiKey"`
	Timeout           time.Duration `json:"timeout" yaml:"timeout" mapstructure:"timeout"`
	MaxRetries        int           `json:"maxRetries" yaml:"maxRetries" mapstructure:"maxRetries"`
	OpenaiCompatiable bool          `json:"openaiCompatiable" yaml:"openaiCompatiable" mapstructure:"openaiCompatiable"`
}
