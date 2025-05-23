package agents

import (
	"bytes"
	"context"
	"fmt"
	"net/http"
	"strings"

	xlog "com.imilair/chatbot/bootstrap/log"
	"com.imilair/chatbot/internal/bcode"
	"com.imilair/chatbot/internal/model"
	"com.imilair/chatbot/internal/service"
	"com.imilair/chatbot/internal/service/config"
	"com.imilair/chatbot/pkg/llm/api/base"
	"com.imilair/chatbot/pkg/queue"
	"com.imilair/chatbot/pkg/util"
	"com.imilair/chatbot/pkg/xhttp"
	"github.com/gin-gonic/gin"
)

var AssistantService *assistant

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
	cfg          *config.AssistantConfig
	extractName  *AgentModel
	commentImage *AgentModel
	queue        *queue.Queue[model.UserAction]
	endflag      chan bool
}

func (t *assistant) Name() string {
	return "agents.assistant"
}

func (t *assistant) Init() (err error) {
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
				actions, _ := t.queue.Dequeue(context.Background(), 10)
				for _, action := range actions {
					fn := actionFuncs[action.ActionType]
					if fn != nil {
						fn(&action)
					}
				}
			}
		}
	})
	xlog.Infof("`%s` inited", t.Name())
	AssistantService = t
	return nil
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

func (a *assistant) handleChat(req *model.UserAction)         {}
func (a *assistant) handleGroupChat(req *model.UserAction)    {}
func (a *assistant) handleFollow(req *model.UserAction)       {}
func (a *assistant) handleCancelFollow(req *model.UserAction) {}
func (a *assistant) handleLike(req *model.UserAction)         {}
func (a *assistant) handleCancelLike(req *model.UserAction)   {}
func (a *assistant) handleJoinGroup(req *model.UserAction)    {}
func (a *assistant) handleExistGroup(req *model.UserAction)   {}
func (a *assistant) handleComment(req *model.UserAction)      {}
func (a *assistant) handleReplyComment(req *model.UserAction) {}
func (a *assistant) handleCommentPic(req *model.UserAction)   {}
