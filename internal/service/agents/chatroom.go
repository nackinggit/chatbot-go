package agents

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	xlog "com.imilair/chatbot/bootstrap/log"
	"com.imilair/chatbot/internal/model"
	"com.imilair/chatbot/internal/model/dbmodel"
	"com.imilair/chatbot/internal/service"
	"com.imilair/chatbot/internal/service/config"
	"com.imilair/chatbot/internal/service/imapi"
	"com.imilair/chatbot/internal/service/memory"
	"com.imilair/chatbot/pkg/llm"
	"com.imilair/chatbot/pkg/llm/api/base"
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
	processedJoin *ttlmap.TTLMap
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
	ctx := context.Background()
	util.AsyncGoWithDefault(ctx, func() {
		xlog.Infof("`%s` chatroom message handler started", t.Name())
		for {
			select {
			case <-t.endflag:
				xlog.Infof("`%s` chatroom message handler stopped", t.Name())
				return
			default:
				chatRoomMessages, _ := t.queue.Dequeue(ctx, 10)
				if len(chatRoomMessages) > 0 {
					for _, roomMessage := range chatRoomMessages {
						t.handleRoomMessage(ctx, &roomMessage)
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

func (t *chatroom) replyUser(ctx context.Context, chatroomSetting *imapi.ChatRoomSetting, userInfo *model.ChatRoomUserInfo) {
	// roomId = speak.roomId
	// bot, gptmsg, discards = await chat_room.reply_speak(speak)
	// if bot is None or gptmsg is None:
	//
	//	return
	//
	// reply: ReplyMessage = ReplyMessage(
	//
	//	sender_id=bot.imBotId,
	//	replyTo=ReplyTo(target_id=int(roomId)),
	//	content=ReplyContent(content=gptmsg.getContent()),
	//
	// )
	// logger.info(f"发送消息: {reply.model_dump_json()}")
	// resp = await bot_api.send_chat_message(reply=reply, scene="room")
	// if resp is not None and resp.status_code == 200:
	//
	//	logger.info(f"发送结果: {resp.json()}")
	//	await chat_room.update_session(roomId, speak, reply)
	//
	// else:
	//
	//	logger.info(f"发送消息失败: {reply.model_dump_json()}")
	//
	// if len(discards) > 50:
	//
	//	_summary_topic(speak.roomId, speak.topic)
}

func (t *chatroom) InputRecommend(ctx *gin.Context, req *model.InputRecommendRequest) {
	// content := req.GetContent()

	// his, _ = sessionManager.build_session(speak.roomId).history_without_rag(content)
	// record = "\n".join([f"{m['content']}" for m in his if m["content"]])
	// input := fmt.Sprintf("【话题介绍】\n%s %s\n", req.Topic.Name, req.Topic.Content.Intro)
	// input += f"【聊天记录】\n{record}\n"

}

func (t *chatroom) handleRoomMessage(ctx context.Context, req *model.Room) {
	if req.UserInfo == nil {
		xlog.Warnf("[处理聊天室消息] 用户信息为空")
		return
	}
	roomId := req.RoomId
	chatroomSetting, err := imapi.ImapiService.QueryChatRoomSetting(ctx, roomId)
	if err != nil {
		xlog.Errorf("查询聊天室信息失败：%v", err)
		return
	}
	if req.UserInfo.Action == "join" {
		xlog.Infof("用户 %d 加入聊天室 %d", req.UserInfo.Nickname, roomId)
		t.welcomeUser(ctx, chatroomSetting, req.UserInfo)
	} else if req.UserInfo.Action == "speak" {
		xlog.Infof("用户 %d 发送小纸条到聊天室 %d", req.UserInfo.Nickname, roomId)
		t.replyUser(ctx, chatroomSetting, req.UserInfo)
	} else {
		xlog.Warnf("未知聊天室事件：%v", req.UserInfo.Action)
	}
}

func (t *chatroom) randomPresenter(ctx context.Context, chatroomSetting *imapi.ChatRoomSetting) (*AgentModel, error) {
	presenter := util.RandSelect([]*imapi.ImUser{chatroomSetting.PresenterA, chatroomSetting.PresenterB})
	cfg := presenter.ParseAiConfig()
	mapi, err := llm.GetApi(cfg.ModelApi)
	if err != nil {
		xlog.Warnf("llm api: %v 未注册, err: ", cfg.ModelApi, err)
		return nil, err
	}
	return &AgentModel{
		LLMModel: &base.LLMModel{
			Api:   mapi,
			Name:  presenter.Nickname,
			Model: cfg.ModelCode,
		},
		Cfg: &config.BotConfig{
			BotId:    presenter.UserId.String(),
			Name:     presenter.Nickname,
			ModelKey: cfg.ModelCode,
			Model:    cfg.ModelCode,
			Api:      cfg.ModelApi,
		},
	}, nil
}

func (t *chatroom) welcomeUser(ctx context.Context, chatroomSetting *imapi.ChatRoomSetting, userInfo *model.ChatRoomUserInfo) {
	xlog.Infof("处理欢迎语：: %s", util.JsonString(userInfo))
	if t.processedJoin.Contains(fmt.Sprintf("%s.%s", chatroomSetting.Id.String(), userInfo.Nickname)) {
		xlog.Infof("[聊天室:%s]处理欢迎语：已处理过: %s", chatroomSetting.Id.String(), util.JsonString(userInfo))
		return
	}
	t.processedJoin.Put(fmt.Sprintf("%s.%s", chatroomSetting.Id.String(), userInfo.Nickname), true)
	bot, err := t.randomPresenter(ctx, chatroomSetting)
	if err != nil {
		xlog.Warnf("处理欢迎语失败: %v", err)
	}
	nickname := userInfo.GetNickname()
	intro := userInfo.GetIntro()
	content := fmt.Sprintf("作为一个有经验的主持人，用你丰富多彩的主持经验，欢迎刚来主题`%s`捧场的用户：\n"+
		"```\n用户名称：%s\n用户简介：%s```", chatroomSetting.Topic, nickname, intro)
	input := &dbmodel.LlmChatHistory{
		ID:      util.NewSnowflakeID().Int64(),
		Mid:     util.Md5Object(content),
		ImBotID: bot.Cfg.BotId,
		Role:    string(base.USER),
		Message: content,
	}
	session := memory.GetTempSession(ctx, chatroomSetting.Id.String())
	memories := session.FetchRelatedMemory(ctx, content, 5000-len([]rune(content)))
	messages := cvtMemory(memories, input)
	output, err := bot.Chat(ctx, messages)
	if err != nil {
		xlog.Warnf("处理欢迎语失败: %v", err)
	} else {
		err = imapi.ImapiService.SendMessage(&imapi.ReplyMessage{
			SenderId: input.ImBotID,
			ReplyTo: &imapi.ReplyTo{
				TargetId: util.StringToInt64(string(chatroomSetting.Id)),
			},
			ReplyContent: &imapi.ReplyContent{
				Content: output.Content,
				Type:    imapi.Text,
			},
		}, "room")
		if err != nil {
			xlog.Warnf("发送欢迎语失败: %v", err)
		} else {
			session.AddMemory(ctx, &memory.MemoryItems{
				CreateTime: time.Now().UnixMilli(),
				Memories: []*dbmodel.LlmChatHistory{input, {
					ID:       util.NewSnowflakeID().Int64(),
					Mid:      util.Md5Object(output.Content),
					ImUserID: chatroomSetting.Id.String(),
					ImBotID:  bot.Cfg.BotId,
					Role:     string(base.Assistant),
					Message:  output.Content,
				}},
				Sid: util.Md5Object(fmt.Sprintf("%v\n%v", fmt.Sprintf("%s 进入了聊天室", nickname), output.Content)),
			})
		}
		session.SetSessionActive()
	}
}

func (t *chatroom) activateChat(ctx context.Context, chatroomSetting *imapi.ChatRoomSetting, triggerMsg string) {
	title := chatroomSetting.Topic.Title
	mhis := []*base.MessageInput{base.UserStringMessage(fmt.Sprintf("作为一个资深主持人，为了避免冷场，请用你丰富的经验让话题`%s`活跃起来\n"+"    %s", title, triggerMsg))}
	times := rand.Int31n(5)
	presenter, _ := t.randomPresenter(ctx, chatroomSetting)
	for i := range int(times) {
		output, err := presenter.Chat(ctx, mhis)
		if err != nil {
			xlog.Warnf("Error: %v", err)
			return
		}
		err = imapi.ImapiService.SendMessage(&imapi.ReplyMessage{
			SenderId: presenter.Cfg.BotId,
			ReplyTo: &imapi.ReplyTo{
				TargetId: util.StringToInt64(string(chatroomSetting.Id)),
			},
			ReplyContent: &imapi.ReplyContent{
				Content: output.Content,
				Type:    imapi.Text,
			},
		}, "room")
		if err != nil {
			xlog.Warnf("Error: %v", err)
		}
		if i == 0 {
			mhis = []*base.MessageInput{base.UserStringMessage(output.Content)}
		} else {
			mhis[len(mhis)-1].Role = base.Assistant
			mhis = append(mhis, base.UserStringMessage(output.Content))
		}
	}
}
