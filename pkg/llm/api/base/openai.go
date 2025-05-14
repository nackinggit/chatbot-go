package base

import (
	"context"
	"net/http"

	"com.imilair/chatbot/bootstrap/config"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/option"
	"github.com/openai/openai-go/packages/ssestream"
)

type OpenaiCompatiableApi struct {
	client *openai.Client
	cfg    *config.LLMConfig
}

var httpclient = http.Client{}

func NewOpenaiCompatibleApi(cfg *config.LLMConfig) *OpenaiCompatiableApi {
	toOpts := func(cfg *config.LLMConfig) []option.RequestOption {
		opts := []option.RequestOption{
			option.WithHTTPClient(&httpclient),
		}
		if cfg.ApiKey != "" {
			opts = append(opts, option.WithAPIKey(cfg.ApiKey))
		}
		if cfg.BaseUrl != "" {
			opts = append(opts, option.WithBaseURL(cfg.BaseUrl))
		}
		if cfg.Timeout > 0 {
			opts = append(opts, option.WithRequestTimeout(cfg.Timeout))
		}
		if cfg.MaxRetries > 0 {
			opts = append(opts, option.WithMaxRetries(cfg.MaxRetries))
		}

		// 拦截器
		middlewares := []option.Middleware{}
		middlewares = append(middlewares, func(request *http.Request, nextfn option.MiddlewareNext) (*http.Response, error) {
			return nextfn(request)
		})
		opts = append(opts, option.WithMiddleware(middlewares...))
		return opts
	}
	oc := openai.NewClient(toOpts(cfg)...)
	return &OpenaiCompatiableApi{
		client: &oc,
		cfg:    cfg,
	}
}

func toOpenaiMessages(messages []MessageInput) ([]openai.ChatCompletionMessageParamUnion, error) {
	var msgs []openai.ChatCompletionMessageParamUnion
	for _, m := range messages {
		msg, err := m.ToOpenaiMessage()
		if err != nil {
			return nil, err
		}
		msgs = append(msgs, msg)
	}
	return msgs, nil
}

func (o *OpenaiCompatiableApi) Chat(ctx context.Context, model string, messages []MessageInput) (output Output, err error) {
	oms, err := toOpenaiMessages(messages)
	if err != nil {
		return output, err
	}
	cc, err := o.client.Chat.Completions.New(ctx, openai.ChatCompletionNewParams{
		Messages: oms,
		Model:    model,
	})
	if err != nil {
		return output, err
	}
	co := OpenaiCompatiableMessageOutput{OpenaiChatCompletion: cc}
	return co.MessageOutput()
}

func (o *OpenaiCompatiableApi) StreamChat(ctx context.Context, model string, messages []MessageInput) *ssestream.Stream[OutputChunk] {
	oms, err := toOpenaiMessages(messages)
	var s *OpenaiCompatiableMessageStream
	if err != nil {
		s = &OpenaiCompatiableMessageStream{
			OpenaiCompatiableStream: ssestream.NewStream[openai.ChatCompletionChunk](nil, err),
		}

	} else {
		stream := o.client.Chat.Completions.NewStreaming(ctx, openai.ChatCompletionNewParams{
			Messages: oms,
			Model:    model,
		})
		s = &OpenaiCompatiableMessageStream{
			OpenaiCompatiableStream: stream,
		}
	}
	return s.Stream()
}

func (o *OpenaiCompatiableApi) Cfg() *config.LLMConfig {
	return o.cfg
}
