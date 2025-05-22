package agents

import (
	xlog "com.imilair/chatbot/bootstrap/log"
	"com.imilair/chatbot/internal/model"
	"com.imilair/chatbot/internal/service"
	"github.com/gin-gonic/gin"
)

var ChatRoomService *chatroom

type chatroom struct {
	topicRecModel *AgentModel
	hostModel1    *AgentModel
	hostModel2    *AgentModel
}

func (t *chatroom) Name() string {
	return "chatroom"
}

func (t *chatroom) Init() (err error) {
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
	xlog.Infof("`%s` inited", t.Name())
	ChatRoomService = t
	return nil
}

func init() {
	service.Register(&teacher{})
}

func ChatRoom() *chatroom {
	return service.Service[chatroom]("chatroom")
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
