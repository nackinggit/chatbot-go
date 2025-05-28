package config

type LoggerConfig struct {
	LogFileDir    string   `json:"logFileDir" yaml:"logFileDir" mapstructure:"logFileDir"`          //文件保存地方
	LogFilename   string   `json:"logFilename" yaml:"logFilename" mapstructure:"logFilename"`       //日志文件前缀
	ErrorFileName string   `json:"errorFileName" yaml:"errorFileName" mapstructure:"errorFileName"` //错误日志文件名
	WarnFileName  string   `json:"warnFileName" yaml:"warnFileName" mapstructure:"warnFileName"`    //警告日志文件名
	InfoFileName  string   `json:"infoFileName" yaml:"infoFileName" mapstructure:"infoFileName"`    //信息日志文件名
	DebugFileName string   `json:"debugFileName" yaml:"debugFileName" mapstructure:"debugFileName"` //调试日志文件名
	Level         string   `json:"level" yaml:"level" mapstructure:"level"`                         //日志等级
	MaxSize       int      `json:"maxSize" yaml:"maxSize" mapstructure:"maxSize"`                   //日志文件小大（M）
	MaxBackups    int      `json:"maxBackups" yaml:"maxBackups" mapstructure:"maxBackups"`          //最多存在多少个切片文件
	MaxAge        int      `json:"maxAge" yaml:"maxAge" mapstructure:"maxAge"`                      //保存的最大天数
	Console       bool     `json:"console" yaml:"console" mapstructure:"console"`                   //是否打印控制台
	CtxFields     []string `json:"ctxFields" yaml:"ctxFields" mapstructure:"ctxFields"`             //需要添加到上下文的字段
}
