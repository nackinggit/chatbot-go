package config

import "errors"

type ChatRoomConfig struct {
	TopicRecommend *BotConfig `json:"topicRecommend" yaml:"topicRecommend" mapstructure:"topicRecommend"` // 话题推荐
	WelcomeTTL     int        `json:"welcomeTTL" yaml:"welcomeTTL" mapstructure:"welcomeTTL"`             // 欢迎语缓存时间
	RoomTTL        int        `json:"roomTTL" yaml:"roomTTL" mapstructure:"roomTTL"`                      // 房间缓存时间
}

func (t *ChatRoomConfig) Validate() error {
	if t == nil {
		return errors.New("config.chatroom is nil")
	}
	if t.TopicRecommend == nil {
		return errors.New("config.chatroom is invalid")
	}
	return nil
}
