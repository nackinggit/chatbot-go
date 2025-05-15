package config

import "time"

type EmbeddingConfig struct {
	Company    string                 `json:"company" yaml:"company" mapstructure:"company"`
	BaseUrl    string                 `json:"baseUrl" yaml:"baseUrl" mapstructure:"baseUrl"`
	ApiKey     string                 `json:"apiKey" yaml:"apiKey" mapstructure:"apiKey"`
	Timeout    time.Duration          `json:"timeout" yaml:"timeout" mapstructure:"timeout"`
	MaxRetries int                    `json:"maxRetries" yaml:"maxRetries" mapstructure:"maxRetries"`
	Models     []EmbeddingModelConfig `json:"models" yaml:"models" mapstructure:"models"`
}

type EmbeddingModelConfig struct {
	Name  string `json:"name" yaml:"name" mapstructure:"name"`
	Model string `json:"model" yaml:"model" mapstructure:"model"`
	Dims  []int  `json:"dims" yaml:"dims" mapstructure:"dims"`
}
