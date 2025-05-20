package agents

import (
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

	doWrite := func(t T) {
		if sseStream.lock != nil {
			sseStream.lock.Lock()
		}
		ctx.SSEvent("data", util.JsonString(t))
		ctx.Writer.Flush()
		if sseStream.lock != nil {
			sseStream.lock.Unlock()
		}
	}

	ctx.Writer.Header().Set("Content-Type", "text/event-stream")
	ctx.Writer.Header().Set("Cache-Control", "no-cache")
	ctx.Writer.Header().Set("Connection", "keep-alive")
	ctx.Writer.Header().Set("Transfer-Encoding", "chunked")

	for stream.Next() {
		output := stream.Current()
		t := dataHandler(&output, nil)
		xlog.Infof("data: %v", util.JsonString(t))
		doWrite(t)
	}

	if stream.Err() != nil {
		t := dataHandler(nil, stream.Err())
		doWrite(t)
	}
	ctx.Writer.Flush()
}
