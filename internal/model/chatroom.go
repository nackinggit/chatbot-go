package model

type InputRecommendRequest struct {
	RoomId     string            `json:"roomId" binding:"required" err:"roomId is required"`
	Topic      *ChatRoomTopic    `json:"topic" binding:"required" err:"topic is required"`
	UserInfo   *ChatRoomUserInfo `json:"user"`
	CreateTime int64             `json:"createTime"`
}

func (r *InputRecommendRequest) GetContent() string {
	if r.UserInfo == nil || r.UserInfo.Content == nil {
		return ""
	}
	return r.UserInfo.Content.Text
}

type ChatRoomTopic struct {
	Name    string       `json:"name" binding:"required" err:"topic name is required"`
	Type    string       `json:"type" binding:"required" err:"topic type is required"`
	Content TopicContent `json:"content"`
}

type TopicContent struct {
	Intro       string        `json:"intro"`
	VoteOptions []*VoteOption `json:"voteOptions"`
}

type VoteOption struct {
	Id    int    `json:"id"`
	Title string `json:"title"`
	Count int    `json:"count"`
}

type ChatRoomUserInfo struct {
	Nickname string       `json:"nickname"`
	Intro    string       `json:"intro"`
	Action   string       `json:"action"`
	Content  *UserContent `json:"content"`
}

type UserContent struct {
	Text         string `json:"text"`
	VoteOptionId int    `json:"voteOptionId"`
}
