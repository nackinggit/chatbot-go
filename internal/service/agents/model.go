package agents

import (
	"io"
	"sync"

	xlog "com.imilair/chatbot/bootstrap/log"
	"com.imilair/chatbot/internal/service/config"
	"com.imilair/chatbot/pkg/llm"
	"com.imilair/chatbot/pkg/llm/api/base"
	"com.imilair/chatbot/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/openai/openai-go/packages/ssestream"
)

type AgentModel struct {
	*base.LLMModel
	Cfg *config.BotConfig
}

func initModel(cfg *config.BotConfig) (*AgentModel, error) {
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

type sseStream[T any] struct {
	lock *sync.RWMutex

	stream      *ssestream.Stream[base.OutputChunk]
	dataHandler func(output *base.OutputChunk, err error) T
}

func sseResponse[T any](ctx *gin.Context, sseStream *sseStream[T]) {
	stream := sseStream.stream
	dataHandler := sseStream.dataHandler

	doWrite := func(t T) bool {
		if sseStream.lock != nil {
			sseStream.lock.Lock()
		}
		clientGone := ctx.Stream(func(w io.Writer) bool {
			ctx.SSEvent("data", util.JsonString(t))
			return false
		})
		if sseStream.lock != nil {
			sseStream.lock.Unlock()
		}
		return clientGone
	}

	for stream.Next() {
		output := stream.Current()
		t := dataHandler(&output, nil)
		xlog.Infof("data: %v", util.JsonString(t))
		if doWrite(t) {
			return
		}
	}

	if stream.Err() != nil {
		t := dataHandler(nil, stream.Err())
		if doWrite(t) {
			return
		}
	}
	ctx.Writer.Flush()
}
