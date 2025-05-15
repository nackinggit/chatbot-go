package config

import (
	"time"
)

type Config struct {
	Env        string                      `json:"env" yaml:"env" mapstructure:"env"`
	App        *Application                `json:"app" yaml:"app" mapstructure:"app"`
	HttpServer *HttpServerConfig           `json:"httpServer" yaml:"httpServer" mapstructure:"httpServer"`
	Logger     *LoggerConfig               `json:"logger" yaml:"logger" mapstructure:"logger"`
	MySql      map[string]*MySQLConfig     `json:"mysql" yaml:"mysql" mapstructure:"mysql"`
	LLM        map[string]*LLMConfig       `json:"llm" yaml:"llm" mapstructure:"llm"`
	Embedding  map[string]*EmbeddingConfig `json:"embedding" yaml:"embedding" mapstructure:"embedding"`
}

func (c *Config) GetGracefulTimeout() time.Duration {
	if c == nil {
		return 0 * time.Second
	}
	if c.HttpServer == nil || c.HttpServer.Http == nil {
		return 0 * time.Second
	}
	return c.HttpServer.Http.GracefulTimeout
}
