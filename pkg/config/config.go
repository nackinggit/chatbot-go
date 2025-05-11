package config

import "com.imilair/chatbot/pkg/log"

type Config struct {
	Env       string       `json:"env" yaml:"env" mapstructure:"env"`
	App       *Application `json:"app" yaml:"app" mapstructure:"app"`
	LoggerOpt *log.Options `json:"logger" yaml:"logger" mapstructure:"logger"`
}

type Application struct {
	Name string `json:"name" yaml:"name" mapstructure:"name"`
}
