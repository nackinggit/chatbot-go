package model

import (
	"errors"

	"com.imilair/chatbot/pkg/util"
)

type HandleCallback func(useraction *UserAction)

type ActionType string

const (
	CHAT          ActionType = "chat"
	GROUPCHAT     ActionType = "groupChat"
	FOLLOW        ActionType = "follow"
	CANCEL_FOLLOW ActionType = "cancelFollow"
	LIKE          ActionType = "like"
	CANCEL_LIKE   ActionType = "cancelLike"
	JOIN_GROUP    ActionType = "joinGroup"
	EXIST_GROUP   ActionType = "existGroup"
	COMMENT       ActionType = "comment"
	REPLY_COMMENT ActionType = "replyComment"
	COMMENT_PIC   ActionType = "commentPic"
	ROOM          ActionType = "room"
)

type UserAction struct {
	UserId        string     `json:"userId"`
	ActionType    ActionType `json:"actionType"`
	ActionContent any        `json:"actionContent"`
}

type Chat struct {
	ReceiverId  string   `json:"receiverId"`
	MsgId       string   `json:"msgId"`
	SceneIds    []string `json:"sceneIds"`
	BotNickname string   `json:"botNickname"`
}

type GroupChat struct {
	GroupId  string   `json:"groupId"`
	MsgId    string   `json:"msgId"`
	Mentions []string `json:"mentions"`
}

type Follow struct {
	FollowUserId string `json:"followUserId"`
}

type Like struct {
	Id   string `json:"likeUserId"`
	Type string `json:"type"` // post-帖子 comment-评论
}

type Group struct {
	GroupId string `json:"groupId"`
}

type Comment struct {
	Id   string `json:"id"`
	Type string `json:"type"` // post-帖子 comment-评论
}

type Room struct {
	RoomId     string            `json:"roomId" binding:"required" err:"roomId is required"`
	Topic      *ChatRoomTopic    `json:"topic" binding:"required" err:"topic is required"`
	UserInfo   *ChatRoomUserInfo `json:"user"`
	CreateTime int64             `json:"createTime"`
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
	Id           string `json:"id"`
	Text         string `json:"text"`
	VoteOptionId int    `json:"voteOptionId"`
}

func GetUserActionContent[T any](ua *UserAction) (*T, error) {
	if ua == nil || ua.ActionContent == nil {
		return nil, errors.New("nil user action")
	}
	var t T
	bs, err := util.Marshal(ua.ActionContent)
	if err != nil {
		return nil, err
	}
	err = util.Unmarshal(bs, &t)
	return &t, err
}
