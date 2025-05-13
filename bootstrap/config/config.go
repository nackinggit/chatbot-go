package config

import (
	"errors"
	"fmt"
	"time"
)

type Config struct {
	Env        string                  `json:"env" yaml:"env" mapstructure:"env"`
	App        *Application            `json:"app" yaml:"app" mapstructure:"app"`
	HttpServer *HttpServerConfig       `json:"httpServer" yaml:"httpServer" mapstructure:"httpServer"`
	Logger     *LoggerConfig           `json:"logger" yaml:"logger" mapstructure:"logger"`
	MySql      map[string]*MySQLConfig `json:"mysql" yaml:"mysql" mapstructure:"mysql"`
	LLM        map[string]*LLMConfig   `json:"llm" yaml:"llm" mapstructure:"llm"`
}

type LoggerConfig struct {
	LogFileDir    string `json:"logFileDir" yaml:"logFileDir" mapstructure:"logFileDir"`          //文件保存地方
	LogFilename   string `json:"logFilename" yaml:"logFilename" mapstructure:"logFilename"`       //日志文件前缀
	ErrorFileName string `json:"errorFileName" yaml:"errorFileName" mapstructure:"errorFileName"` //错误日志文件名
	WarnFileName  string `json:"warnFileName" yaml:"warnFileName" mapstructure:"warnFileName"`    //警告日志文件名
	InfoFileName  string `json:"infoFileName" yaml:"infoFileName" mapstructure:"infoFileName"`    //信息日志文件名
	DebugFileName string `json:"debugFileName" yaml:"debugFileName" mapstructure:"debugFileName"` //调试日志文件名
	Level         string `json:"level" yaml:"level" mapstructure:"level"`                         //日志等级
	MaxSize       int    `json:"maxSize" yaml:"maxSize" mapstructure:"maxSize"`                   //日志文件小大（M）
	MaxBackups    int    `json:"maxBackups" yaml:"maxBackups" mapstructure:"maxBackups"`          //最多存在多少个切片文件
	MaxAge        int    `json:"maxAge" yaml:"maxAge" mapstructure:"maxAge"`                      //保存的最大天数
	Console       bool   `json:"console" yaml:"console" mapstructure:"console"`                   //是否打印控制台
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

type LLMConfig struct {
	Name    string           `json:"name" yaml:"name" mapstructure:"name"`
	BaseUrl string           `json:"baseUrl" yaml:"baseUrl" mapstructure:"baseUrl"`
	ApiKey  string           `json:"apiKey" yaml:"apiKey" mapstructure:"apiKey"`
	Timeout time.Duration    `json:"timeout" yaml:"timeout" mapstructure:"timeout"`
	Models  []LlmModelConfig `json:"models" yaml:"models" mapstructure:"models"`
}
type LlmModelConfig struct {
	Name    string `json:"name" yaml:"name" mapstructure:"name"`
	BaseUrl string `json:"baseUrl" yaml:"baseUrl" mapstructure:"baseUrl"`
}
