package agents

import (
	"fmt"
	"strings"

	xlog "com.imilair/chatbot/bootstrap/log"
	"com.imilair/chatbot/internal/model"
	"com.imilair/chatbot/internal/service"
	"com.imilair/chatbot/pkg/llm/api/base"
	"com.imilair/chatbot/pkg/util"
	"github.com/gin-gonic/gin"
)

var AssistantService *assistant

type assistant struct {
	extractName  *AgentModel
	commentImage *AgentModel
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
	t.extractName, err = initModel(cfg.ExtractName)
	if err != nil {
		return err
	}
	t.commentImage, err = initModel(cfg.CommentImage)
	if err != nil {
		return err
	}

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
