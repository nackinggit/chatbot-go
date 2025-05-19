package agents

import (
	"com.imilair/chatbot/internal/service/config"
	"com.imilair/chatbot/pkg/llm/api/base"
)

type AgentModel struct {
	*base.LLMModel
	Cfg *config.BotConfig
}
