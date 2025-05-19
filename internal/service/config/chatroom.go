package config

import "errors"

type ChatRoomConfig struct {
	TopicRecommend *BotConfig `json:"topicRecommend" yaml:"topicRecommend" mapstructure:"topicRecommend"` // 话题推荐
	HostModel1     *BotConfig `json:"host1" yaml:"host1" mapstructure:"host1"`                            // 主持人1
	HostModel2     *BotConfig `json:"host2" yaml:"host2" mapstructure:"host2"`                            // 主持人2
}

func (t *ChatRoomConfig) Validate() error {
	if t == nil {
		return errors.New("config.chatroom is nil")
	}
	if t.TopicRecommend == nil || t.HostModel1 == nil || t.HostModel2 == nil {
		return errors.New("config.chatroom is invalid")
	}
	return nil
}
