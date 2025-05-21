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
