package config

import "errors"

type DaoConfig struct {
	DbName string `json:"db" yaml:"db" mapstructure:"db"` // 数据库配置
}

func (c *DaoConfig) Validate() error {
	if c == nil || c.DbName == "" {
		return errors.New("config is nil")
	}
	return nil
}
