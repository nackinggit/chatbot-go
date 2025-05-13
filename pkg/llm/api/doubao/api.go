package doubao

import (
	"com.imilair/chatbot/bootstrap/config"
	"com.imilair/chatbot/pkg/llm/api/base"
)

type DouBao struct {
	*base.OpenaiCompatiableApi
}

func InitApi(cfg *config.LLMConfig) base.LLMApi {
	o := base.NewOpenaiCompatibleApi(cfg)
	return &DouBao{
		OpenaiCompatiableApi: o,
	}
}
