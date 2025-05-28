package agents

import (
	xlog "com.imilair/chatbot/bootstrap/log"
	"com.imilair/chatbot/internal/model"
	"com.imilair/chatbot/internal/service"
	"com.imilair/chatbot/pkg/llm/api/base"
	"com.imilair/chatbot/pkg/util"
	"github.com/gin-gonic/gin"
)

var MangHeService *manghe

type manghe struct {
	imageAnalyseModel *AgentModel
	predictModel      *AgentModel
}

func (t *manghe) Name() string {
	return "agens.manghe"
}

func (t *manghe) InitAndStart() (err error) {
	xlog.Infof("init service `%s`", t.Name())
	mangheCfg := service.Config.MangHe
	err = mangheCfg.Validate()
	if err != nil {
		xlog.Warnf("check manghe config failed: %v", err)
		return err
	}
	t.imageAnalyseModel, err = initModel(mangheCfg.ImageAnalyse)
	if err != nil {
		return err
	}
	t.predictModel, err = initModel(mangheCfg.Predict)
	if err != nil {
		return err
	}

	MangHeService = t
	xlog.Infof("`%s` inited", t.Name())
	return nil
}

func (t *manghe) Stop() {}

func init() {
	service.Register(&manghe{})
}

func (m *manghe) ImageAnalyse(ctx *gin.Context, imgUrl string) {
	messages := []*base.MessageInput{
		{
			Role: base.USER,
			MultiModelContents: []base.InputContent{
				{
					Type:    base.Image,
					Content: imgUrl,
				},
			},
		},
	}
	stream := m.imageAnalyseModel.StreamChat(ctx, messages)
	util.SSEHeader(ctx)
	sseResponse(ctx, &sseStream[model.StreamMessage]{
		stream:      stream,
		dataHandler: streamMessageHandlerfunc,
	})
}

func (m *manghe) Predict(ctx *gin.Context, req *model.MangHePredictRequest) {
	ms := []*base.MessageInput{
		base.UserStringMessage(req.ToString()),
	}
	stream := m.predictModel.StreamChat(ctx, ms)
	util.SSEHeader(ctx)
	sseResponse(ctx, &sseStream[model.StreamMessage]{
		stream:      stream,
		dataHandler: streamMessageHandlerfunc,
	})
}
