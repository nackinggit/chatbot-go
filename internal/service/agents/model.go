package agents

import (
	"com.imilair/chatbot/internal/service/config"
	"com.imilair/chatbot/pkg/llm"
	"com.imilair/chatbot/pkg/llm/api/base"
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
