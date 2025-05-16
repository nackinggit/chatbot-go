package base

import (
	"context"
	"fmt"
	"strings"

	"com.imilair/chatbot/bootstrap/config"
	"github.com/openai/openai-go/packages/ssestream"
)

type InitApi func(cfg *config.LLMConfig) LLMApi

type Output struct {
	ReasoningContent string      `json:"reasoning"`
	Content          string      `json:"content"`
	Role             MessageRole `json:"role"`
	Exception        string      `json:"exception"`
	RawJson          string      `json:"rawJson"`
}

func (o *Output) Trim() {
	o.ReasoningContent = strings.TrimSpace(o.ReasoningContent)
	o.Content = strings.TrimSpace(o.Content)
}

type OutputChunk struct {
	ReasoningContent string      `json:"reasoning"`
	Content          string      `json:"content"`
	Role             MessageRole `json:"role"`
	RawJSON          string      `json:"rawJson"`
}

func (o *OutputChunk) HumanText() string {
	return fmt.Sprintf("%s: %s", o.Role, o.Content)
}

type LLMApi interface {
	Cfg() *config.LLMConfig
	Chat(ctx context.Context, model string, messages []MessageInput) (Output, error)
	StreamChat(ctx context.Context, model string, messages []MessageInput) *ssestream.Stream[OutputChunk]
}

type LLMModel struct {
	Name  string `json:"name"`  // 模型名称
	Model string `json:"model"` // 模型代号

	Api LLMApi
}

func (m *LLMModel) Chat(ctx context.Context, messages []MessageInput) (Output, error) {
	return m.Api.Chat(ctx, m.Model, messages)
}

func (m *LLMModel) StreamChat(ctx context.Context, messages []MessageInput) *ssestream.Stream[OutputChunk] {
	return m.Api.StreamChat(ctx, m.Model, messages)
}
