package base

import (
	"context"
	"fmt"
	"strings"

	"com.imilair/chatbot/bootstrap/config"
	xlog "com.imilair/chatbot/bootstrap/log"
	"com.imilair/chatbot/pkg/util"
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
	ReasoningContent string      `json:"reasoning_content"`
	Content          string      `json:"content"`
	Role             MessageRole `json:"role"`
	RawJSON          string      `json:"rawJson"`
	IsLastChunk      bool        `json:"isLastChunk"`
}

func (o *OutputChunk) HumanText() string {
	return fmt.Sprintf("%s: %s", o.Role, o.Content)
}

type LLMApi interface {
	Cfg() *config.LLMConfig
	Chat(ctx context.Context, model string, messages []*MessageInput) (Output, error)
	StreamChat(ctx context.Context, model string, messages []*MessageInput) *ssestream.Stream[OutputChunk]
}

type LLMModel struct {
	Name  string `json:"name"`  // 模型名称
	Model string `json:"model"` // 模型代号

	Api LLMApi
}

func (m *LLMModel) Chat(ctx context.Context, messages []*MessageInput) (Output, error) {
	xlog.DebugC(ctx, "[%s] chat: %v", m.Model, util.BeautifulJson(messages))
	o, e := m.Api.Chat(ctx, m.Model, messages)
	xlog.DebugC(ctx, "[%s] chat end", m.Model)
	return o, e
}

func (m *LLMModel) StreamChat(ctx context.Context, messages []*MessageInput) *ssestream.Stream[OutputChunk] {
	xlog.DebugC(ctx, "[%s], start streamchat: %v", m.Model, util.BeautifulJson(messages))
	stream := m.Api.StreamChat(ctx, m.Model, messages)
	xlog.Debugf("model: %v, end streamchat", m.Model)
	return stream
}
