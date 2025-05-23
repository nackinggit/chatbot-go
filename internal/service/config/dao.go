package config

type DaoConfig struct {
	DbName string `json:"db" yaml:"db" mapstructure:"db"` // 数据库配置
}
