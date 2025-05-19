package agents

import (
	"io"

	xlog "com.imilair/chatbot/bootstrap/log"
	"com.imilair/chatbot/internal/model"
	"com.imilair/chatbot/internal/service"
	"com.imilair/chatbot/internal/service/config"
	"com.imilair/chatbot/pkg/llm"
	"com.imilair/chatbot/pkg/llm/api/base"
	"com.imilair/chatbot/pkg/util"
	"github.com/gin-gonic/gin"
)

type teacher struct {
	questionAnalyserModel *AgentModel
	answererModels        []*AgentModel
	judgeModel            *AgentModel
}

func (t *teacher) Name() string {
	return "teacher"
}

func (t *teacher) Init() (err error) {
	initModel := func(cfg *config.BotConfig) (*AgentModel, error) {
		api, err := llm.GetApi(cfg.Api)
		if err != nil {
			return nil, err
		}
		return &AgentModel{
			LLMModel: &base.LLMModel{
				Name:  cfg.Name,
				Model: cfg.ModelKey,
				Api:   api,
			},
			Cfg: cfg,
		}, nil
	}

	xlog.Infof("init service `%s`", t.Name())
	teacherCfg := service.Config.Teacher
	err = teacherCfg.Validate()
	if err != nil {
		return err
	}
	t.questionAnalyserModel, err = initModel(teacherCfg.QuestionAnalyse)
	if err != nil {
		return err
	}
	for _, am := range teacherCfg.AnswerModels {
		m, e := initModel(am)
		if e != nil {
			return e
		}
		t.answererModels = append(t.answererModels, m)
	}

	t.judgeModel, err = initModel(teacherCfg.JudgeModel)
	if err != nil {
		return err
	}
	xlog.Info("Teacher inited")
	return nil
}

func init() {
	service.Register(&teacher{})
}

func Teacher() *teacher {
	return service.Service[teacher]("teacher")
}

func (t *teacher) QuestionAnalyse(ctx *gin.Context, req *model.QuestionAnalyseRequest) {
	mi := base.MessageInput{
		Role: base.USER,
		MultiModelContents: []base.InputContent{
			{Type: base.Image, Content: req.ImageUrl},
		},
	}
	messages := []*base.MessageInput{&mi}
	stream := t.questionAnalyserModel.StreamChat(ctx, messages)
	util.SSEHeader(ctx)

	finalChunk := &model.QuestionAnalyseStreamChunk{}
	ctx.Stream(func(w io.Writer) bool {
		for stream.Next() {
			chunk := stream.Current()
			xlog.Infof("data: %v", chunk)
			sc := &model.QuestionAnalyseStreamChunk{
				StreamMessage: model.StreamMessage{
					Reasoning: chunk.ReasoningContent,
					Content:   chunk.Content,
				},
			}
			ctx.SSEvent("data", util.JsonString(sc))
			finalChunk.Content += sc.Content
			finalChunk.Reasoning += sc.Reasoning
			return true
		}
		if stream.Err() != nil {
			finalChunk.Content = ""
			finalChunk.Reasoning = ""
			finalChunk.Exception = stream.Err().Error()
		}
		finalChunk.Endflag = true
		ctx.SSEvent("data", util.JsonString(finalChunk))
		return false
	})
}
