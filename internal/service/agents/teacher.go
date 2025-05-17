package agents

import (
	"context"

	xlog "com.imilair/chatbot/bootstrap/log"
	"com.imilair/chatbot/internal/model"
	"com.imilair/chatbot/internal/service"
	"com.imilair/chatbot/pkg/llm"
	"com.imilair/chatbot/pkg/llm/api/base"
	"github.com/openai/openai-go/packages/ssestream"
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

func (t *teacher) QuestionAnalyse(ctx context.Context, req *model.QuestionAnalyseRequest) *ssestream.Stream[base.OutputChunk] {
	mi := base.MessageInput{
		Role: base.USER,
		MultiModelContents: []base.InputContent{
			{Type: base.Image, Content: req.ImageUrl},
		},
	}
	messages := []*base.MessageInput{&mi}
	return t.questionAnalyserModel.StreamChat(ctx, messages)
}
