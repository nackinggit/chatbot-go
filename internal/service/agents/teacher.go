package agents

import (
	"io"

	xlog "com.imilair/chatbot/bootstrap/log"
	"com.imilair/chatbot/internal/model"
	"com.imilair/chatbot/internal/service"
	"com.imilair/chatbot/pkg/llm"
	"com.imilair/chatbot/pkg/llm/api/base"
	"com.imilair/chatbot/pkg/util"
	"github.com/gin-gonic/gin"
)

type teacher struct {
	questionAnalyserModel *base.LLMModel
	answererModels        []*base.LLMModel
	judgeModel            *base.LLMModel
}

func (t *teacher) Name() string {
	return "teacher"
}

func (t *teacher) Init() (err error) {

	xlog.Infof("init service `%s`", t.Name())
	t.questionAnalyserModel, err = llm.GetModel("QuestionAnalyser")
	if err != nil {
		return err
	}
	t.answererModels, err = llm.GetModels([]string{""})
	if err != nil {
		return err
	}
	t.judgeModel, err = llm.GetModel("Judge")
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
