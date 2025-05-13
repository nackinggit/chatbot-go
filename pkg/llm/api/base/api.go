package base

import (
	"context"

	"com.imilair/chatbot/bootstrap/config"
)

type InitApi func(cfg *config.LLMConfig) LLMApi

type LLMApi interface {
	Chat(ctx context.Context, model *LLMModel, messages []MessageInput, extras map[string]any) (MessageOutput, error)
	StreamChat(ctx context.Context, model *LLMModel, messages []MessageInput, extras map[string]any) (<-chan MessageOutput, error)
}

type LLMModel struct {
	Name  string `json:"name"`  // 模型名称
	Model string `json:"model"` // 模型代号

	Api LLMApi
}

func (m *LLMModel) Chat(ctx context.Context, messages []MessageInput, extras map[string]any) (MessageOutput, error) {
	return m.Api.Chat(ctx, m, messages, extras)
}

func (m *LLMModel) StreamChat(ctx context.Context, messages []MessageInput, extras map[string]any) (<-chan MessageOutput, error) {
	return m.Api.StreamChat(ctx, m, messages, extras)
}
