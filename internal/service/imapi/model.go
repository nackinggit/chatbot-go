package imapi

import (
	"encoding/json"

	xlog "com.imilair/chatbot/bootstrap/log"
	"com.imilair/chatbot/pkg/util"
)

type ContentType string

const (
	Text ContentType = "text"
)

type ReplyMessage struct {
	SenderId     string        `json:"sender_id"`
	ReplyTo      *ReplyTo      `json:"replyTo"`
	ReplyContent *ReplyContent `json:"content"`
}

type ReplyTo struct {
	TargetId int64 `json:"target_id"`
}

type ReplyContent struct {
	SrcContentId int         `json:"srcContentId"`
	Content      string      `json:"content"`
	Type         ContentType `json:"type"`
}

type ChatContent struct {
	Model    string `json:"model"`
	ChatMode string `json:"chatMode"`
	Text     string `json:"text"`
	Type     string `json:"type"`
}

// im评论
type ImComment struct {
	PostId    string  `json:"postId"`
	Content   string  `json:"content"`
	Commentor *ImUser `json:"user"`
}

// im用户
type ImUser struct {
	UserId   json.Number `json:"userId"`
	Nickname string      `json:"nickname"`
	AiConfig string      `json:"aiConf"`
}

func (imuser *ImUser) ParseAiConfig() *AiConfig {
	if imuser == nil || imuser.AiConfig == "" {
		return nil
	}
	var aiCfg AiConfig
	err := util.Unmarshal([]byte(imuser.AiConfig), &aiCfg)
	if err != nil {
		xlog.Warnf("imuser.ParseAiConfig(%v) error: %v", imuser.AiConfig, err)
		return nil
	}
	return &aiCfg
}

type AiConfig struct {
	ModelApi  string `json:"modelApi"`  // 模型api名称
	ModelCode string `json:"modelCode"` // 模型代码
}

// im帖子
type ImPost struct {
	Title string  `json:"title"`
	User  *ImUser `json:"user"`
}

// im帖子下的评论
type PostComments struct {
	Comments []*ImComment `json:"list"`
}

// im聊天室设置
type ChatRoomSetting struct {
	Id         json.Number `json:"id"`
	PresenterA *ImUser     `json:"presenterA"`
	PresenterB *ImUser     `json:"presenterB"`
	Topic      *ChatTopic  `json:"topic"`
}

func (c *ChatRoomSetting) GetTopicTitle() string {
	if c.Topic != nil {
		return c.Topic.Title
	} else {
		return "未知聊天室"
	}
}

type ChatTopic struct {
	Image string `json:"image"`
	Title string `json:"title"`
}
