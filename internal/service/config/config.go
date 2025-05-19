package config

type ServiceConfig struct {
	Teacher  *TeacherConfig  `json:"teacher" yaml:"teacher" mapstructure:"teacher"`    // 解题助手配置
	ChatRoom *ChatRoomConfig `json:"chatroom" yaml:"chatroom" mapstructure:"chatroom"` // 聊天室配置
	MangHe   *MangHeConfig   `json:"manghe" yaml:"manghe" mapstructure:"manghe"`       // 盲盒配置
}

type BotConfig struct {
	Model    string `json:"model" yaml:"model" mapstructure:"model"`          // 模型名称
	Name     string `json:"name" yaml:"name" mapstructure:"name"`             // 机器人名称
	ModelKey string `json:"modelKey" yaml:"modelKey" mapstructure:"modelKey"` // 模型标识
	Api      string `json:"api" yaml:"api" mapstructure:"api"`                // 模型API
	BotId    int    `json:"botId" yaml:"botId" mapstructure:"botId"`          // 机器人ID
}
