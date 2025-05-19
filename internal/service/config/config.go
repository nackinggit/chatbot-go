package config

type ServiceConfig struct {
	Teacher *TeacherConfig `json:"teacher" yaml:"teacher" mapstructure:"teacher"` // 解题助手配置
}

type BotConfig struct {
	Model    string `json:"model" yaml:"model" mapstructure:"model"`          // 模型名称
	Name     string `json:"name" yaml:"name" mapstructure:"name"`             // 机器人名称
	ModelKey string `json:"modelKey" yaml:"modelKey" mapstructure:"modelKey"` // 模型标识
}
