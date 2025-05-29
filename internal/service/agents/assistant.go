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
	t.defaultReasoningBot, err = initModel(cfg.ReasoningChat)
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
							xlog.Warnf("`%s` unknown action type: %v, ignore message: %s", t.Name(), action.ActionType, util.JsonString(action))
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
	content := fmt.Sprintf("根据以下`用户`和`小助手`的对话, 提取`用户`对`小助手`的昵称\n"+
		"```%s```\n"+
		"请用json的格式输出，如：{\"nickname\": <对小助手的称呼>}", record)
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

func (a *assistant) OutsideList(ctx *gin.Context, req *model.OutsideListRequest) (*model.OutsideListResponse, error) {
	data := `[{"tip":"《莉可丽丝》短篇动画即将推出，颠覆性暗黑童话风格引发期待。【来源：二次元现场】","link":"","tags":["新番","动漫"]},{"tip":"《烈焰之刃》正式公开，黑暗幻想风动作冒险游戏新作将于 5 月 22 日发售。【来源：二次元现场】","tags":["新番","动漫"],"link":""},{"tip":"《指环王：夏尔传说》再次延期，计划于 7 月 29 日发售。【来源：二次元现场】","tags":["新番","动漫"],"link":""},{"tip":"《境・界 刀鸣》始解测试招募开启，改编自《BLEACH》的游戏备受关注。【来源：二次元现场】","tags":["新番","动漫"],"link":""},{"tip":"2025 年春季动画专题上线，萌娘百科推荐多部新番。【来源：萌娘百科】","tags":["新番","动漫"],"link":""},{"tip":"《咩咩启示录》全球销量突破 450 万份，官方将推出重金属音乐专辑。【来源：二次元现场】","tags":["新番","动漫"],"link":""},{"tip":"《阿凡达：潘多拉边境》破空者 DLC 公布，将于 7 月 16 日上线。【来源：二次元现场】","tags":["新番","动漫"],"link":""},{"tip":"《最终幻想 14》“黄金的遗产” 新消息公布，国际服抢先体验 6 月 28 日开启。【来源：二次元现场】","tags":["新番","动漫"],"link":""},{"tip":"横版冒险游戏《Symphonia》将于 12 月 5 日推出。【来源：二次元现场】","tags":["新番","动漫"],"link":""},{"tip":"《卧龙：苍天陨落》全球玩家总数突破 500 万。【来源：二次元现场】","tags":["新番","动漫"],"link":""}]`
	items := []*model.OutsideItem{}
	util.Unmarshal([]byte(data), &items)
	resp := &model.OutsideListResponse{
		Items: items,
	}
	return resp, nil
}

func (a *assistant) UserActionCallback(ctx *gin.Context, req *model.UserAction) (any, error) {
	err := a.queue.Enqueue(ctx, *req)
	return nil, err
}

func cvtMemory(memories []*memory.MemoryItems, input *dbmodel.LlmChatHistory) []*base.MessageInput {
	mis := []*base.MessageInput{}
	for _, mi := range memories {
		for _, m := range mi.Memories {
			mis = append(mis, &base.MessageInput{
				Role:          base.MessageRole(m.Role),
				StringContent: m.Message,
			})
		}
	}
	mis = append(mis, base.UserStringMessage(input.Message))
	return mis
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
	input := &dbmodel.LlmChatHistory{
		ID:       util.NewSnowflakeID().Int64(),
		Mid:      chat.MsgId,
		ImUserID: chat.SenderId,
		ImBotID:  chat.ReceiverId,
		Role:     string(base.USER),
		Message:  content.Text,
	}
	session := memory.GetSession(ctx, chat.ChatSessionId())
	memories := session.FetchRelatedMemory(ctx, content.Text, 5000)
	messages := cvtMemory(memories, input)
	output, err := imBot.Chat(ctx, messages)
	if err != nil {
		xlog.Warnf("imbot聊天失败, err:%v", err)
		imapi.ImapiService.SendMessage(&imapi.ReplyMessage{
			SenderId: chat.ReceiverId,
			ReplyTo: &imapi.ReplyTo{
				TargetId: util.StringToInt64(chat.SenderId),
			},
			ReplyContent: &imapi.ReplyContent{
				Content: "...(沉默一会儿)",
				Type:    imapi.Text,
			},
		}, "chat")
		return
	}
	err = imapi.ImapiService.SendMessage(&imapi.ReplyMessage{
		SenderId: chat.ReceiverId,
		ReplyTo: &imapi.ReplyTo{
			TargetId: util.StringToInt64(chat.SenderId),
		},
		ReplyContent: &imapi.ReplyContent{
			Content: output.Content,
			Type:    imapi.Text,
		},
	}, "chat")
	if err == nil {
		session.AddMemory(ctx, &memory.MemoryItems{
			CreateTime: time.Now().UnixMilli(),
			Memories: []*dbmodel.LlmChatHistory{input, {
				ID:       util.NewSnowflakeID().Int64(),
				Mid:      util.Md5Object(output.Content),
				ImUserID: chat.SenderId,
				ImBotID:  chat.ReceiverId,
				Role:     string(base.Assistant),
				Message:  output.Content,
			}},
			Sid: util.Md5Object(fmt.Sprintf("%v\n%v", input.Message, output.Content)),
		})
	} else {
		xlog.Warnf("imbot聊天失败, err:%v", err)
	}
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

func (a *assistant) handleComment(ctx context.Context, req *model.UserAction) {
	comment, err := model.GetUserActionContent[model.Comment](req)
	if err != nil {
		xlog.Warnf("解析chat异常, err:%v, useraction:%v", err, util.JsonString(req))
		return
	}
	if comment.Type == "post" {
		// 帖子评论
		xlog.DebugC(ctx, "处理帖子评论消息: %v", util.JsonString(comment))
	} else if comment.Type == "comment" {
		// 评论评论
		xlog.DebugC(ctx, "处理评论评论消息: %v", util.JsonString(comment))
	}
}
func (a *assistant) handleGroupChat(ctx context.Context, req *model.UserAction)    {}
func (a *assistant) handleFollow(ctx context.Context, req *model.UserAction)       {}
func (a *assistant) handleCancelFollow(ctx context.Context, req *model.UserAction) {}
func (a *assistant) handleLike(ctx context.Context, req *model.UserAction)         {}
func (a *assistant) handleCancelLike(ctx context.Context, req *model.UserAction)   {}
func (a *assistant) handleJoinGroup(ctx context.Context, req *model.UserAction)    {}
func (a *assistant) handleExistGroup(ctx context.Context, req *model.UserAction)   {}
func (a *assistant) handleReplyComment(ctx context.Context, req *model.UserAction) {}
func (a *assistant) handleCommentPic(ctx context.Context, req *model.UserAction)   {}
