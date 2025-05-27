package agents

import (
	"context"
	"time"

	xlog "com.imilair/chatbot/bootstrap/log"
	"com.imilair/chatbot/internal/model"
	"com.imilair/chatbot/internal/service"
	"com.imilair/chatbot/internal/service/imapi"
	"com.imilair/chatbot/pkg/queue"
	"com.imilair/chatbot/pkg/util"
	"com.imilair/chatbot/pkg/util/ttlmap"
	"github.com/gin-gonic/gin"
)

var ChatRoomService *chatroom

type chatroom struct {
	topicRecModel *AgentModel
	queue         *queue.Queue[model.Room]
	endflag       chan bool
	tmpvalMap     *ttlmap.TTLMap
}

func (t *chatroom) Name() string {
	return "agents.chatroom"
}

func (t *chatroom) InitAndStart() (err error) {
	xlog.Infof("init service `%s`", t.Name())
	chatroomCfg := service.Config.ChatRoom
	err = chatroomCfg.Validate()
	if err != nil {
		return err
	}
	t.topicRecModel, err = initModel(chatroomCfg.TopicRecommend)
	if err != nil {
		return err
	}
	t.queue = queue.NewQueue[model.Room]("chatroom:message")
	t.endflag = make(chan bool)
	util.AsyncGoWithDefault(context.Background(), func() {
		xlog.Infof("`%s` chatroom message handler started", t.Name())
		for {
			select {
			case <-t.endflag:
				xlog.Infof("`%s` chatroom message handler stopped", t.Name())
				return
			default:
				chatRoomMessages, _ := t.queue.Dequeue(context.Background(), 10)
				if len(chatRoomMessages) > 0 {
					for _, roomMessage := range chatRoomMessages {
						t.handleRoomMessage(&roomMessage)
					}
				} else {
					time.Sleep(100 * time.Millisecond)
				}
			}
		}
	})
	xlog.Infof("`%s` inited", t.Name())
	ChatRoomService = t
	return nil
}

func (t *chatroom) Stop() {
	t.endflag <- true
}

func init() {
	service.Register(&chatroom{})
}

func ChatRoom() *chatroom {
	return service.Service[chatroom]("chatroom")
}

func (t *chatroom) RoomActionCallback(ctx *gin.Context, req *model.Room) (any, error) {
	t.queue.Enqueue(ctx, *req)
	return nil, nil
}

func (t *chatroom) replyUser(chatroomSetting *imapi.ChatRoomSetting, userInfo *model.ChatRoomUserInfo) {

}

func (t *chatroom) InputRecommend(ctx *gin.Context, req *model.InputRecommendRequest) {
	// content := req.GetContent()

	// his, _ = sessionManager.build_session(speak.roomId).history_without_rag(content)
	// record = "\n".join([f"{m['content']}" for m in his if m["content"]])
	// input := fmt.Sprintf("【话题介绍】\n%s %s\n", req.Topic.Name, req.Topic.Content.Intro)
	// input += f"【聊天记录】\n{record}\n"

}

func (t *chatroom) handleRoomMessage(req *model.Room) {
	if req.UserInfo == nil {
		xlog.Warnf("[处理聊天室消息] 用户信息为空")
		return
	}
	roomId := req.RoomId
	chatroomSetting, err := imapi.ImapiService.QueryChatRoomSetting(roomId)
	if err != nil {
		xlog.Errorf("查询聊天室信息失败：%v", err)
		return
	}
	if req.UserInfo.Action == "join" {
		xlog.Infof("用户 %d 加入聊天室 %d", req.UserInfo.Nickname, roomId)
		t.welcomeUser(chatroomSetting, req.UserInfo)
	} else if req.UserInfo.Action == "speak" {
		xlog.Infof("用户 %d 发送小纸条到聊天室 %d", req.UserInfo.Nickname, roomId)
		t.replyUser(chatroomSetting, req.UserInfo)
	} else {
		xlog.Warnf("未知聊天室事件：%v", req.UserInfo.Action)
	}
}

func (t *chatroom) welcomeUser(chatroomSetting *imapi.ChatRoomSetting, userInfo *model.ChatRoomUserInfo) {

}
