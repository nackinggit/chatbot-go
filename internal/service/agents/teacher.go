package agents

import (
	"io"
	"net/http"
	"sync"

	xlog "com.imilair/chatbot/bootstrap/log"
	"com.imilair/chatbot/internal/model"
	"com.imilair/chatbot/internal/service"
	"com.imilair/chatbot/internal/service/config"
	"com.imilair/chatbot/pkg/llm"
	"com.imilair/chatbot/pkg/llm/api/base"
	"com.imilair/chatbot/pkg/util"
	"github.com/gin-gonic/gin"
)

var TeacherService *teacher

type teacher struct {
	questionAnalyserModel *AgentModel
	qaModels              []*AgentModel
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
		t.qaModels = append(t.qaModels, m)
	}

	t.judgeModel, err = initModel(teacherCfg.JudgeModel)
	if err != nil {
		return err
	}
	TeacherService = t
	xlog.Infof("`%s` inited", t.Name())
	return nil
}

func init() {
	service.Register(&teacher{})
}

func Teacher() *teacher {
	return service.Service[teacher]("teacher")
}

func (t *teacher) QuestionAnalyse(ctx *gin.Context, req *model.ImageRequest) {
	mi := base.MessageInput{
		Role: base.USER,
		MultiModelContents: []base.InputContent{
			{Type: base.Image, Content: req.ImageUrl},
		},
	}
	messages := []*base.MessageInput{&mi}
	stream := t.questionAnalyserModel.StreamChat(ctx, messages)
	sseResponse(ctx, &sseStream[model.StreamMessage]{
		stream:      stream,
		dataHandler: streamMessageHandlerfunc,
	})
}

func (t *teacher) AnswerQuestion(ctx *gin.Context, req *model.QARequest) {
	models := []*AgentModel{}
	if len(req.Models) > 0 {
		for _, m := range req.Models {
			var model *AgentModel
			for _, mm := range t.qaModels {
				if mm.Cfg.Name == m {
					model = mm
					break
				}
			}
			if model != nil {
				models = append(models, model)
			} else {
				xlog.Warnf("Model not found: %s", m)
				ctx.JSON(http.StatusBadRequest, gin.H{"error": "Model not found"})
				return
			}
		}
	} else {
		models = t.qaModels
	}
	mi := base.UserStringMessage(req.Question)
	messages := []*base.MessageInput{mi}
	wg := sync.WaitGroup{}
	wg.Add(len(models))
	lock := sync.RWMutex{}
	finalChunks := []*model.QAStreamChunk{}
	for _, amodel := range models {
		util.AsyncGoWithDefault(ctx, func() {
			defer wg.Done()
			stream := amodel.StreamChat(ctx, messages)
			sseResponse(ctx, &sseStream[model.QAStreamChunk]{
				stream: stream,
				lock:   &lock,
				dataHandler: func(output *base.OutputChunk, err error) model.QAStreamChunk {
					chunk := model.QAStreamChunk{
						StreamMessage: &model.StreamMessage{},
						Model:         &amodel.Cfg.Model,
						Name:          &amodel.Cfg.Name,
					}
					if err != nil {
						chunk.Exception = err.Error()
						chunk.Endflag = true
					} else if output != nil {
						chunk.Reasoning = output.ReasoningContent
						chunk.Content = output.Content
						chunk.Endflag = output.IsLastChunk
						if output.IsLastChunk {
							finalChunks = append(finalChunks, &chunk)
						}
					}
					return chunk
				},
			})
		})
	}
	wg.Wait()
	ctx.Stream(func(w io.Writer) bool {
		chunk := &model.QAStreamChunk{
			AllAnswers: finalChunks,
			AllEndflag: true,
		}
		ctx.SSEvent("data", util.JsonString(chunk))
		return false
	})
}

func (t *teacher) JudgeAnswer(ctx *gin.Context, req *model.JudgeAnswerRequest) {
	ancontents := []map[string]string{}
	for _, answer := range req.Answers {
		ancontents = append(ancontents, map[string]string{"模型名称": answer.Name, "回答内容": answer.Content})
	}

	contentDict := map[string]any{"原题": req.Question, "答案": map[string]any{"答案信息": ancontents}}
	messages := []*base.MessageInput{base.UserStringMessage(util.BeautifulJson(contentDict))}
	stream := t.judgeModel.StreamChat(ctx, messages)
	sseResponse(ctx, &sseStream[model.StreamMessage]{
		stream:      stream,
		dataHandler: streamMessageHandlerfunc,
	})
}
