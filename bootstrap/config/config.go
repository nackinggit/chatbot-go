package config

import (
	"time"
)

type Config struct {
	Env        string             `json:"env" yaml:"env" mapstructure:"env"`
	App        *Application       `json:"app" yaml:"app" mapstructure:"app"`
	HttpServer *HttpServerConfig  `json:"httpServer" yaml:"httpServer" mapstructure:"httpServer"`
	Logger     *LoggerConfig      `json:"logger" yaml:"logger" mapstructure:"logger"`
	MySql      []*MySQLConfig     `json:"mysqls" yaml:"mysqls" mapstructure:"mysqls"`
	LLMS       []*LLMConfig       `json:"llms" yaml:"llms" mapstructure:"llms"`
	Embedding  []*EmbeddingConfig `json:"embedding" yaml:"embedding" mapstructure:"embedding"`
	Milvus     *MilvusConfig      `json:"milvus" yaml:"milvus" mapstructure:"milvus"`
	Redis      *RedisConfig       `json:"redis" yaml:"redis" mapstructure:"redis"`
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
