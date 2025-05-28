package config

import "errors"

type ServiceConfig struct {
	Teacher   *TeacherConfig   `json:"teacher" yaml:"teacher" mapstructure:"teacher"`       // 解题助手配置
	ChatRoom  *ChatRoomConfig  `json:"chatroom" yaml:"chatroom" mapstructure:"chatroom"`    // 聊天室配置
	MangHe    *MangHeConfig    `json:"manghe" yaml:"manghe" mapstructure:"manghe"`          // 盲盒配置
	Assistant *AssistantConfig `json:"assistant" yaml:"assistant" mapstructure:"assistant"` // IM助手配置
	Dao       *DaoConfig       `json:"dao" yaml:"dao" mapstructure:"dao"`                   // 数据库配置
	ImApi     *ImApi           `json:"imApi" yaml:"imApi" mapstructure:"imApi"`             // IM API配置
	Memory    *MemoryConfig    `json:"memory" yaml:"memory" mapstructure:"memory"`          // 记忆库配置
}

type MemoryConfig struct {
	LongMemory  *LongMemoryConfig  `json:"longMemory" yaml:"vectlongMemory" mapstructure:"longMemory"` // 长记忆配置
	ShortMemory *ShortMemoryConfig `json:"shortMemory" yaml:"shortMemory" mapstructure:"shortMemory"`  // 短记忆配置
}

func (m *MemoryConfig) Validate() error {
	if m == nil || m.LongMemory == nil {
		return errors.New("long memory config is not valid")
	}
	if m.ShortMemory == nil {
		m.ShortMemory = &ShortMemoryConfig{
			TTL:     7200,
			MaxSize: 4000,
		}
	}
	return nil
}

type LongMemoryConfig struct {
	VectorDim int    `json:"vectorDim" yaml:"vectorDim" mapstructure:"vectorDim"` // 向量维度
	EmbApi    string `json:"embApi" yaml:"embApi" mapstructure:"embApi"`          // 长记忆API配置
	EmbModel  string `json:"embModel" yaml:"embModel" mapstructure:"embModel"`    // 长记忆模型配置
}

type ShortMemoryConfig struct {
	TTL     int `json:"ttl" yaml:"ttl" mapstructure:"ttl"`             // 短记忆TTL
	MaxSize int `json:"maxSize" yaml:"maxSize" mapstructure:"maxSize"` // 短记忆最大容量
}

type BotConfig struct {
	Model    string `json:"model" yaml:"model" mapstructure:"model"`          // 模型名称
	Name     string `json:"name" yaml:"name" mapstructure:"name"`             // 机器人名称
	ModelKey string `json:"modelKey" yaml:"modelKey" mapstructure:"modelKey"` // 模型标识
	Api      string `json:"api" yaml:"api" mapstructure:"api"`                // 模型API
	BotId    string `json:"botId" yaml:"botId" mapstructure:"botId"`          // 机器人ID
}

type ImApi struct {
	BaseUrl string `json:"baseUrl" yaml:"baseUrl" mapstructure:"baseUrl"` // 接口地址
}

func (i *ImApi) Validate() error {
	if i == nil || i.BaseUrl == "" {
		return errors.New("imapi config is invalid")
	}
	return nil
}
