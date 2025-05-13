package base

import (
	"context"
	"net/http"

	"com.imilair/chatbot/bootstrap/config"
)

type OpenaiCompatiableApi struct {
	client *http.Client
	cfg    *config.LLMConfig
}

var client = http.Client{}

func NewOpenaiCompatibleApi(cfg *config.LLMConfig) *OpenaiCompatiableApi {
	return &OpenaiCompatiableApi{
		client: &client,
		cfg:    cfg,
	}
}

func (o *OpenaiCompatiableApi) Chat(ctx context.Context, model *LLMModel, messages []MessageInput, extras map[string]any) (output MessageOutput, err error) {
	return output, err
}

func (o *OpenaiCompatiableApi) StreamChat(ctx context.Context, model *LLMModel, messages []MessageInput, extras map[string]any) (<-chan MessageOutput, error) {
	c := make(chan MessageOutput)
	for {
		c <- MessageOutput{}
	}
}
