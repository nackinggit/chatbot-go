package config

import (
	"errors"
	"fmt"
	"time"

	"github.com/gin-gonic/gin"
)

type Server interface {
	Start() error
	Stop() error
	HealthCheck() error
	Engine() *gin.Engine
}

type Application struct {
	Name string `json:"name" yaml:"name" mapstructure:"name"`
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
	CORS              *CORSConfig   `json:"cors" yaml:"cors" mapstructure:"cors"`
}

type CORSConfig struct {
	AllowAllOrigins bool          `json:"allowAllOrigins" yaml:"allowAllOrigins" mapstructure:"allowAllOrigins"`
	AllowOrigins    []string      `json:"allowOrigins" yaml:"allowOrigins" mapstructure:"allowOrigins"`
	AllHeaders      []string      `json:"allHeaders" yaml:"allHeaders" mapstructure:"allHeaders"`
	MaxAge          time.Duration `json:"maxAge" yaml:"maxAge" mapstructure:"maxAge"`
	Domains         []string      `json:"domains" yaml:"domains" mapstructure:"domains"`
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
	return fmt.Sprintf("%s:%d", hsc.Host, hsc.Port), nil
}
