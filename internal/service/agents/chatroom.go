package agents

import (
	"context"
	"time"

	xlog "com.imilair/chatbot/bootstrap/log"
	"com.imilair/chatbot/internal/model"
	"com.imilair/chatbot/internal/service"
	"com.imilair/chatbot/pkg/queue"
	"com.imilair/chatbot/pkg/util"
	"github.com/gin-gonic/gin"
)

var ChatRoomService *chatroom

type chatroom struct {
	topicRecModel *AgentModel
	hostModel1    *AgentModel
	hostModel2    *AgentModel
	queue         *queue.Queue[model.Room]
	endflag       chan bool
}

func (t *chatroom) Name() string {
	return "chatroom"
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
	t.hostModel1, err = initModel(chatroomCfg.HostModel1)
	if err != nil {
		return err
	}
	t.hostModel2, err = initModel(chatroomCfg.HostModel2)
	if err != nil {
		return err
	}
	t.queue = queue.NewQueue[model.Room]("chatroom_message")
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
	return nil, nil
}

func (t *chatroom) ReplySpeak() string {
	return ""
}

func (t *chatroom) InputRecommend(ctx *gin.Context, req *model.InputRecommendRequest) {
	// content := req.GetContent()

	// his, _ = sessionManager.build_session(speak.roomId).history_without_rag(content)
	// record = "\n".join([f"{m['content']}" for m in his if m["content"]])
	// input := fmt.Sprintf("【话题介绍】\n%s %s\n", req.Topic.Name, req.Topic.Content.Intro)
	// input += f"【聊天记录】\n{record}\n"

}

func (t *chatroom) handleRoomMessage(req *model.Room) {

}
