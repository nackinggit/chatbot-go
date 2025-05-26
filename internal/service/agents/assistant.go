package agents

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"
	"sync"
	"time"

	xlog "com.imilair/chatbot/bootstrap/log"
	"com.imilair/chatbot/internal/bcode"
	"com.imilair/chatbot/internal/model"
	"com.imilair/chatbot/internal/model/dbmodel"
	"com.imilair/chatbot/internal/service"
	"com.imilair/chatbot/internal/service/config"
	"com.imilair/chatbot/internal/service/dao"
	"com.imilair/chatbot/internal/service/imapi"
	"com.imilair/chatbot/internal/service/memory"
	"com.imilair/chatbot/pkg/llm/api/base"
	"com.imilair/chatbot/pkg/queue"
	"com.imilair/chatbot/pkg/util"
	"com.imilair/chatbot/pkg/xhttp"
	"github.com/gin-gonic/gin"
)

var AssistantService *assistant
var registeredImBot = sync.Map{}

var actionFuncs = map[model.ActionType]model.HandleCallback{
	model.CHAT:          AssistantService.handleChat,
	model.GROUPCHAT:     AssistantService.handleGroupChat,
	model.FOLLOW:        AssistantService.handleFollow,
	model.CANCEL_FOLLOW: AssistantService.handleCancelFollow,
	model.LIKE:          AssistantService.handleLike,
	model.CANCEL_LIKE:   AssistantService.handleCancelLike,
	model.JOIN_GROUP:    AssistantService.handleJoinGroup,
	model.EXIST_GROUP:   AssistantService.handleExistGroup,
	model.COMMENT:       AssistantService.handleComment,
	model.REPLY_COMMENT: AssistantService.handleReplyComment,
	model.COMMENT_PIC:   AssistantService.handleCommentPic,
}

type assistant struct {
	cfg                 *config.AssistantConfig
	extractName         *AgentModel
	commentImage        *AgentModel
	defaultChatBot      *AgentModel
	defaultReasoningBot *AgentModel
	queue               *queue.Queue[model.UserAction]
	endflag             chan bool
}

func (t *assistant) Name() string {
	return "agents.assistant"
}

func (t *assistant) InitAndStart() (err error) {
	xlog.Infof("init service `%s`", t.Name())
	cfg := service.Config.Assistant
	err = cfg.Validate()
	if err != nil {
		return err
	}
	t.cfg = cfg
	t.extractName, err = initModel(cfg.ExtractName)
	if err != nil {
		return err
	}
	t.commentImage, err = initModel(cfg.CommentImage)
	if err != nil {
		return err
	}
	t.defaultChatBot, err = initModel(cfg.Chat)
	if err != nil {
		return err
	}
	t.queue = queue.NewQueue[model.UserAction]("user_action")
	t.endflag = make(chan bool)
	util.AsyncGoWithDefault(context.Background(), func() {
		xlog.Infof("`%s` useraction handler started", t.Name())
		for {
			select {
			case <-t.endflag:
				xlog.Infof("`%s` stopped", t.Name())
				return
			default:
				ctx := context.Background()
				actions, _ := t.queue.Dequeue(ctx, 10)
				if len(actions) > 0 {
					for _, action := range actions {
						fn := actionFuncs[action.ActionType]
						if fn != nil {
							fn(ctx, &action)
						} else {
							xlog.Warnf("`%s` unknown action type: %d, ignore message: %s", t.Name(), action.ActionType, util.JsonString(action))
						}
					}
				} else {
					time.Sleep(100 * time.Millisecond)
				}
			}
		}
	})
	xlog.Infof("`%s` inited", t.Name())
	AssistantService = t
	return nil
}

func (t *assistant) Stop() {
	t.endflag <- true
}

func init() {
	service.Register(&assistant{})
}

func (a *assistant) ExtractName(ctx *gin.Context, req *model.ExtractNameRequest) (*model.ExtractNameResponse, error) {
	record := strings.Join(req.Content, "\n")
	content := fmt.Sprintf("根据以下`用户`和`小助手`的对话, 提取`用户`对`小助手`的昵称\n```%s```\n请用json的格式输出，如：{\"nickname\": <对小助手的称呼>}", record)
	ms := []*base.MessageInput{
		base.UserStringMessage(content),
	}
	output, err := a.extractName.Chat(ctx, ms)
	if err != nil {
		return nil, err
	}
	var resp model.ExtractNameResponse
	err = util.TryParseJson(output.Content, &resp)
	return &resp, err
}

func (a *assistant) CommentPic(ctx *gin.Context, req *model.CommentPicRequest) (*model.CommentPicResponse, error) {
	ms := []*base.MessageInput{
		base.UserMultiModalMessage([]base.InputContent{
			base.ImagePart(req.PicUrl),
		}),
	}
	output, err := a.commentImage.Chat(ctx, ms)
	if err != nil {
		return nil, err
	}
	return &model.CommentPicResponse{
		Comment: output.Content,
	}, nil
}

func (a *assistant) ComicTranslate(ctx *gin.Context, req *model.ImageRequest) (*model.ComicTranslateResponse, error) {
	bs, err := util.Marshal(req)
	if err != nil {
		return nil, err
	}
	request, err := http.NewRequestWithContext(ctx, http.MethodPost, a.cfg.ComicTranslate, bytes.NewBuffer(bs))
	if err != nil {
		return nil, err
	}

	var resp struct {
		Code    int                           `json:"code"`
		Message string                        `json:"msg"`
		Data    *model.ComicTranslateResponse `json:"data"`
	}
	err = xhttp.DoAndBind(request, &resp)
	if err != nil {
		return nil, err
	}
	if resp.Data == nil {
		return nil, bcode.New(500, "翻译失败")
	}
	return resp.Data, nil
}

func (a *assistant) UserActionCallback(ctx *gin.Context, req *model.UserAction) (any, error) {
	err := a.queue.Enqueue(ctx, *req)
	return nil, err
}

func (a *assistant) handleChat(ctx context.Context, req *model.UserAction) {
	chat, err := model.GetUserActionContent[model.Chat](req)
	if err != nil {
		xlog.Warnf("解析chat异常, err:%v, useraction:%v", err, util.JsonString(req))
		return
	}
	content, err := imapi.ImapiService.QueryChatContent(chat.ReceiverId, chat.MsgId)
	if err != nil {
		xlog.Warnf("获取chat内容失败, err:%v", err)
		return
	}
	val, _ := registeredImBot.LoadOrStore(chat.ReceiverId, a.loadImBot(chat.ReceiverId, chat.BotNickname, content))
	imBot := val.(*AgentModel)
	input := dbmodel.LlmChatHistory{
		ID:       util.NewSnowflakeID().Int64(),
		Mid:      chat.MsgId,
		Sid:      chat.ChatSessionId(),
		ImUserID: chat.SenderId,
		ImBotID:  chat.ReceiverId,
		Role:     string(base.USER),
		Message:  content.Text,
	}
	memories := memory.FetchRelatedMemory(ctx, chat.ChatSessionId(), content.Text, 5000)
}

func (a *assistant) loadImBot(imBotId string, imBotName string, imcontent *imapi.ChatContent) *AgentModel {
	defaultChatAgent := &AgentModel{
		LLMModel: a.defaultChatBot.LLMModel,
		Cfg: &config.BotConfig{
			ModelKey: a.defaultChatBot.Model,
			Model:    a.defaultChatBot.Name,
			Name:     imBotName,
			BotId:    imBotId,
			Api:      a.defaultChatBot.Api.Cfg().RegisterService,
		},
	}

	defaultReasoningAgent := &AgentModel{
		LLMModel: a.defaultReasoningBot.LLMModel,
		Cfg: &config.BotConfig{
			ModelKey: a.defaultReasoningBot.Model,
			Model:    a.defaultReasoningBot.Name,
			Name:     imBotName,
			BotId:    imBotId,
			Api:      a.defaultReasoningBot.Api.Cfg().RegisterService,
		},
	}
	if imcontent.ChatMode == "reasoning" && imcontent.Model == "deepseek" {
		return defaultReasoningAgent
	}

	imbot, err := dao.QueryById(context.Background(), dbmodel.LlmModel{}, imBotId)
	if err != nil {
		xlog.Errorf("获取imbot失败, err:%v, 使用默认bot", err)
		return defaultChatAgent
	}
	m, err := initModel(&config.BotConfig{
		Model:    imbot.ModelName,
		ModelKey: imbot.ModelKey,
		Name:     imbot.BindImBotName,
		Api:      imbot.API,
		BotId:    imbot.BindImBotID,
	})
	if err != nil {
		xlog.Errorf("初始化bot失败, err:%v, 使用默认bot", err)
		return &AgentModel{
			LLMModel: a.defaultChatBot.LLMModel,
			Cfg: &config.BotConfig{
				ModelKey: a.defaultChatBot.Model,
				Model:    a.defaultChatBot.Name,
				Name:     imBotName,
				BotId:    imBotId,
				Api:      a.defaultChatBot.Api.Cfg().RegisterService,
			},
		}
	}
	return m
}
func (a *assistant) handleGroupChat(ctx context.Context, req *model.UserAction)    {}
func (a *assistant) handleFollow(ctx context.Context, req *model.UserAction)       {}
func (a *assistant) handleCancelFollow(ctx context.Context, req *model.UserAction) {}
func (a *assistant) handleLike(ctx context.Context, req *model.UserAction)         {}
func (a *assistant) handleCancelLike(ctx context.Context, req *model.UserAction)   {}
func (a *assistant) handleJoinGroup(ctx context.Context, req *model.UserAction)    {}
func (a *assistant) handleExistGroup(ctx context.Context, req *model.UserAction)   {}
func (a *assistant) handleComment(ctx context.Context, req *model.UserAction)      {}
func (a *assistant) handleReplyComment(ctx context.Context, req *model.UserAction) {}
func (a *assistant) handleCommentPic(ctx context.Context, req *model.UserAction)   {}
