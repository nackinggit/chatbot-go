package agents

import (
	"context"
	"errors"
	"fmt"
	"maps"
	"math/rand"
	"slices"
	"strings"
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
	roomMaps      *ttlmap.TTLMap
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
	t.processedJoin = ttlmap.New(1000, chatroomCfg.WelcomeTTL)
	t.roomMaps = ttlmap.New(1000, chatroomCfg.RoomTTL)
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

func (t *chatroom) replyUser(ctx context.Context, chatroomSetting *imapi.ChatRoomSetting, req *model.Room) {
	presenter, err := t.randomPresenter(ctx, chatroomSetting)
	if err != nil {
		xlog.WarnC(ctx, "chatroom.randomPresenter err: %v", err)
		return
	}
	userInfo := req.UserInfo
	chatTopic := req.Topic
	nickname := userInfo.GetNickname()
	intro := userInfo.GetIntro()
	ctype := userInfo.Action
	if userInfo.Content == nil {
		return
	}
	xlog.DebugC(ctx, "[聊天室 %s]用户小纸条: %v", chatroomSetting.Id.String(), util.JsonString(userInfo))
	var input *dbmodel.LlmChatHistory
	content := ""
	if userInfo.Content.Text != "" {
		content = userInfo.Content.Text
		input = &dbmodel.LlmChatHistory{
			ID:      util.NewSnowflakeID().Int64(),
			Mid:     util.Md5Object(userInfo.Content.Text),
			ImBotID: presenter.Cfg.BotId,
			Role:    string(base.USER),
			Message: fmt.Sprintf("作为一个有经验的主持人，用你丰富多彩的主持经验，针对主题`%s`，回答以下观点：\n"+
				"```\n%s（%s）发表了观点：%s\n```", chatTopic.Name, nickname, intro, userInfo.Content.Text),
		}
	} else if ctype == "vote" {
		input = &dbmodel.LlmChatHistory{
			ID:      util.NewSnowflakeID().Int64(),
			Mid:     util.Md5Object(userInfo.Content.Text),
			ImBotID: presenter.Cfg.BotId,
			Role:    string(base.USER),
			Message: fmt.Sprintf("作为一个有经验的主持人，用你丰富多彩的主持经验，总结以下主题`%s`的阶段性投票结果：\n"+
				"```\n%s\n---最新投票信息\n%s 为 `%s` 投了一票\n```", chatTopic.Name, chatTopic.GetVoteOpts(), nickname, userInfo.GetVote()),
		}
	}
	session := memory.GetTempSession(ctx, req.RoomId)
	memories := []*memory.MemoryItems{}
	if content != "" {
		memories = session.FetchRelatedMemory(ctx, content, 5000-len([]rune(input.Message)))
	}
	messages := cvtMemory(memories, input)
	output, err := presenter.Chat(ctx, messages)
	if err != nil {
		xlog.Warnf("处理小纸条失败: %v", err)
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
			xlog.Warnf("回复小纸条失败: %v", err)
		} else {
			if content != "" {
				input.Message = content
				session.AddMemory(ctx, &memory.MemoryItems{
					CreateTime: time.Now().UnixMilli(),
					Memories: []*dbmodel.LlmChatHistory{input, {
						ID:       util.NewSnowflakeID().Int64(),
						Mid:      util.Md5Object(output.Content),
						ImUserID: chatroomSetting.Id.String(),
						ImBotID:  presenter.Cfg.BotId,
						Role:     string(base.Assistant),
						Message:  output.Content,
					}},
					Sid: util.Md5Object(fmt.Sprintf("%v\n%v", content, output.Content)),
				})
			}
		}
		session.SetSessionActive()
	}
}

func (t *chatroom) InputRecommend(ctx *gin.Context, req *model.InputRecommendRequest) ([]any, error) {
	parseRec := func(resp string) ([]any, error) {
		ret := []any{}
		var respArr []map[string]any
		err := util.TryParseJsonArray(resp, &respArr)
		if err != nil {
			return ret, err
		}
		for _, a := range respArr {
			ret = append(ret, slices.Collect(maps.Values(a))...)
		}
		return ret, nil
	}
	contet := req.GetContent()
	memories := memory.GetTempSession(ctx, req.RoomId).FetchRelatedMemory(ctx, contet, 3000-len(contet))
	records := []string{}
	for _, memory := range memories {
		records = append(records, memory.ToString())
	}
	input := fmt.Sprintf("【话题介绍】\n%s %s", req.Topic.Name, req.Topic.GetIntro())
	input += fmt.Sprintf("【聊天记录】\n%s", strings.Join(records, "\n"))
	messages := []*base.MessageInput{base.UserStringMessage(input)}
	output, err := t.topicRecModel.Chat(ctx, messages)
	if err != nil {
		return nil, err
	}
	return parseRec(output.Content)
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
		xlog.Infof("用户 %s 加入聊天室 %s", req.UserInfo.Nickname, roomId)
		t.welcomeUser(ctx, chatroomSetting, req.UserInfo)
	} else if req.UserInfo.Action == "speak" {
		xlog.Infof("用户 %s 发送小纸条到聊天室 %s", req.UserInfo.Nickname, roomId)
		t.replyUser(ctx, chatroomSetting, req)
	} else {
		xlog.Warnf("未知聊天室事件：%v", req.UserInfo.Action)
	}
}

func (t *chatroom) getRoomPresenters(ctx context.Context, chatroomSetting *imapi.ChatRoomSetting) ([]*AgentModel, error) {
	ret := []*AgentModel{}
	if presenters, ok := t.roomMaps.Get(chatroomSetting.Id.String()); ok {
		return presenters.([]*AgentModel), nil
	}
	imbots := []*imapi.ImUser{chatroomSetting.PresenterA, chatroomSetting.PresenterB}
	for _, imbot := range imbots {
		cfg := imbot.ParseAiConfig()
		mapi, err := llm.GetApi(cfg.ModelApi)
		if err != nil {
			xlog.WarnC(ctx, "llm api: %v 未注册, err: ", cfg.ModelApi, err)
			return nil, err
		}
		ret = append(ret, &AgentModel{
			LLMModel: &base.LLMModel{
				Api:   mapi,
				Name:  imbot.Nickname,
				Model: cfg.ModelCode,
			},
			Cfg: &config.BotConfig{
				BotId:    imbot.UserId.String(),
				Name:     imbot.Nickname,
				ModelKey: cfg.ModelCode,
				Model:    cfg.ModelCode,
				Api:      cfg.ModelApi,
			},
		})
	}

	return ret, nil
}

func (t *chatroom) randomPresenter(ctx context.Context, chatroomSetting *imapi.ChatRoomSetting) (*AgentModel, error) {
	presenters, err := t.getRoomPresenters(ctx, chatroomSetting)
	if err != nil {
		return nil, err
	}
	return util.RandSelect(presenters), nil
}

func (t *chatroom) nextPresenter(ctx context.Context, prev *AgentModel, chatroomSetting *imapi.ChatRoomSetting) (*AgentModel, error) {
	presenters, err := t.getRoomPresenters(ctx, chatroomSetting)
	if err != nil {
		return nil, err
	}
	for _, p := range presenters {
		if p.Cfg.BotId != prev.Cfg.BotId {
			return p, nil
		}
	}
	return nil, errors.New("no presenter found")
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
		"```\n用户名称：%s\n用户简介：%s```", chatroomSetting.GetTopicTitle(), nickname, intro)
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
			input.Message = fmt.Sprintf("%s 进入了聊天室", nickname)
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
				Sid: util.Md5Object(fmt.Sprintf("%v\n%v", input.Message, output.Content)),
			})
		}
		if !session.IsActiveBefore(5 * time.Minute) {
			session.SetSessionActive()
			t.activateChat(ctx, chatroomSetting, bot, output.Content)
		}
	}
}

func (t *chatroom) activateChat(ctx context.Context, chatroomSetting *imapi.ChatRoomSetting, bot *AgentModel, triggerMsg string) {
	xlog.DebugC(ctx, "开始活跃聊天室: %s", util.JsonString(chatroomSetting))
	title := chatroomSetting.Topic.Title
	mhis := []*base.MessageInput{base.UserStringMessage(fmt.Sprintf("作为一个资深主持人，为了避免冷场，请用你丰富的经验让话题`%s`活跃起来\n"+"    %s", title, triggerMsg))}
	times := 1 + rand.Int31n(5)
	presenter, _ := t.nextPresenter(ctx, bot, chatroomSetting)
	for i := range int(times) {
		output, err := presenter.Chat(ctx, mhis)
		if err != nil {
			xlog.WarnC(ctx, "Error: %v", err)
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
			xlog.WarnC(ctx, "Error: %v", err)
			return
		}
		if i == 0 {
			mhis = []*base.MessageInput{base.UserStringMessage(output.Content)}
		} else {
			mhis[len(mhis)-1].Role = base.Assistant
			mhis = append(mhis, base.UserStringMessage(output.Content))
		}
	}
}
