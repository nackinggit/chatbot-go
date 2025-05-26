package imapi

import "encoding/json"

type ContentType string

const (
	Text ContentType = "text"
)

type ReplyMessage struct {
	SenderId     string        `json:"sender_id"`
	ReplyTo      *ReplyTo      `json:"replyTo"`
	ReplyContent *ReplyContent `json:"content"`
}

func (r *ReplyMessage) getContent() string {
	return r.ReplyContent.Content
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

type ImUser struct {
	UserId   json.Number `json:"userId"`
	Nickname string      `json:"nickname"`
}

type ImPost struct {
	Title string  `json:"title"`
	User  *ImUser `json:"user"`
}

type PostComments struct {
	Comments []*ImComment `json:"list"`
}
