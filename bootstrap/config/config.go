package config

import (
	"errors"
	"fmt"
	"time"

	xlog "com.imilair/chatbot/bootstrap/log"
)

type Config struct {
	Env    string             `json:"env" yaml:"env" mapstructure:"env"`
	App    *Application       `json:"app" yaml:"app" mapstructure:"app"`
	Logger *xlog.LoggerConfig `json:"logger" yaml:"logger" mapstructure:"logger"`
}

func (c *Config) GetGracefulTimeout() time.Duration {
	if c == nil {
		return 0 * time.Second
	}
	if c.App == nil || c.App.HttpServer == nil || c.App.HttpServer.Http == nil {
		return 0 * time.Second
	}
	return c.App.HttpServer.Http.GracefulTimeout
}

type Application struct {
	Name       string            `json:"name" yaml:"name" mapstructure:"name"`
	HttpServer *HttpServerConfig `json:"httpServer" yaml:"httpServer" mapstructure:"httpServer"`
}

type HttpServerConfig struct {
	Http      *HttpConfig      `json:"http" yaml:"http" mapstructure:"http"`
	GinServer *GinServerConfig `json:"gin" yaml:"gin" mapstructure:"gin"`
}

type HttpConfig struct {
	Port              int           `json:"port" yaml:"port" mapstructure:"port"`
	Host              string        `json:"host" yaml:"host" mapstructure:"host"`
	MaxListenLimit    int           `json:"maxListenLimit" yaml:"maxListenLimit" mapstructure:"maxListenLimit"`
	ReadTimeout       time.Duration `json:"readTimeout" yaml:"readTimeout" mapstructure:"readTimeout"`
	ReadHeaderTimeout time.Duration `json:"readHeaderTimeout" yaml:"readHeaderTimeout" mapstructure:"readHeaderTimeout"`
	WriteTimeout      time.Duration `json:"writeTimeout" yaml:"writeTimeout" mapstructure:"writeTimeout"`
	IdleTimeout       time.Duration `json:"idleTimeout" yaml:"idleTimeout" mapstructure:"idleTimeout"`
	MaxHeaderBytes    int           `json:"maxHeaderBytes" yaml:"maxHeaderBytes" mapstructure:"maxHeaderBytes"`
	GracefulTimeout   time.Duration `json:"gracefulTimeout" yaml:"gracefulTimeout" mapstructure:"gracefulTimeout"`
	Tls               *TlsConfig    `json:"tls" yaml:"tls" mapstructure:"tls"`
}

type TlsConfig struct {
	CertFile string `json:"certFile" yaml:"certFile" mapstructure:"certFile"`
	KeyFile  string `json:"keyFile" yaml:"keyFile" mapstructure:"keyFile"`
}

type GinServerConfig struct {
	Mode string `json:"mode" yaml:"mode" mapstructure:"mode"`
}

func (cfg *HttpServerConfig) GetListen() (string, error) {
	if cfg.Http == nil {
		return "", errors.New("empty http config")
	}
	hsc := cfg.Http
	if hsc.Port == 0 {
		return "", errors.New("empty port")
	}
	if hsc.Host == "" {
		return "", errors.New("empty host")
	}
	return fmt.Sprintf("%s:%d", hsc.Host, hsc.Port), nil
}
