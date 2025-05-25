package coze

import (
	"context"
	"errors"
	"fmt"
	"time"

	"com.imilair/chatbot/bootstrap/config"
	xlog "com.imilair/chatbot/bootstrap/log"
	"com.imilair/chatbot/pkg/llm/api/base"
	"com.imilair/chatbot/pkg/xhttp"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/packages/ssestream"
)

const (
	pubkey      = "E739JkYUH0kyhq6UlP_D95Bx1JneRViMxuxw-2C13Ks"
	private_key = `-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQCmDKif+M1qkRLS
O5fUiQjis8SZN88WLcoRKZiwueAdmxegp6YcoiBKqPMI77UCTrFqgSedYQyVO4r4
ztRsPMPAuzOOyvQQ9D5BfBnsQEIBUkcKKXyEFJuybwCF7EviwSj3Dfv+nHcn9vLN
/3dUYwo+yT6hakWw2Ld1Ih5iU3TtgMfhouywb/eQOvZUD+Kn1BCArv0MfCZJeEXF
6KbTaJSX0hPpUPJFWEIKdqQvz7xBFR5ts5nfEiV5p33gRq0ulfDk5YTiMh7hCa/X
U/VPKo75z1LI5OABbfJlujAoUYi0DSpwMDKncqTl/YdjsNT0SjSMrw1pSpUWzOpA
U46S3IgdAgMBAAECggEABUHrpkvUKZtIAroqahGs42XiPaCxZAgiOJqXnOoRuslx
/TjGe9JO6yMMz08s4CxG/H9afMzssoS7Xa180KPrspYRNJfGYQs+tZRjZLvT6zAX
k26LV8pB2l7R3/h10aTtWJJlVKM2UaCSNDjYjPM2CPvJvOyJ7tn5x96cvCvDM5B3
nYoZAzfvCt3Xkm1X42lWfxvvXtJWnwI7MefJyIqNzko3rtMIhdY/vO8IPnavDflr
xP4zv1uNaEu7DAzYcXQKsHpKATE9tjzzXRtYtFd9f1MeUsu+Es1QCTeY4ewAGmVW
v7qs+jAmH3h5g4P8bBTgR8G9OunqXGA1Lz+HAMGZ2QKBgQDkTAQboZB2k6WIGev9
b3OzXyffTifCezQLkzbI+ZEKGeIUnybBxk1h3U1qBHssiodD40eNDbp03cC16OuN
sIUhH/8dhZ+7KfCVJN8tZZSGP5oYeGXSnOVD8a+xqPXlHDeJmCl1yhl1G8yYz0ip
Mh4zCwmuxSH5VRD1eM5f+h3uWQKBgQC6MvDbl4mGjPXA5sh1Gbpa/aj74jRoLfiN
KQGQunhsNIqe8DaejwwQKdu/Qwt8nUW4PfK8HrCQa/31B5k9q0U+tr7zuPiAtboj
vT5hGM6KBJFLFCdgbYw/k5bCofcrTqurrgJrQm9ZTTRrSj4lcidr1Eobo2tACbyz
hymcME+XZQKBgFtQAFCg9axH7+yZGaf7vRZgmA0cMJD8QFvk3QPTtmyI38GJyrG0
xFzBbGZcNnwhSGsh7AuCEzMNQzg/WoAIu6b9Kkg/mxz8cGrnHZEF0TtFEzh4Z5mv
AZCEidaQkxG5kIkrYGHpnPcXUGVKe3CZSDT4VD4gQS9+E9NrJ3iCDRi5AoGAPCtd
/fgYLuy6NZ3eRUkNGX5C7zKH8Op6GVOY9+XqKD1KVlYVsGNVaJu+MS4/NgO0lfce
y3+3WtQq+tV7xZvlAoEXb7bkRuNyxT3QPJxBkgQr13Ep0FVWLu1ImJiyQMJpY08V
5QdQ6DC0sb8KGhurdYLid8/1RnpfCjyxS5GpBqkCgYEAyL02uWLq2NXeDgDKl3lZ
PikxXM1yE8hKVkYVH0nDBZtpynCe4wbD7Folni13RAUZ0V92EgJ6of5RZ01Iy7zs
28dsSl1BEgbV1EPYlHNVjMYWsfdPcgB8SVLqKHLev0ShNkdchiAKOCjlsuj245Ph
aRqQzsEgy4AfjbEByOP3GQc=
-----END PRIVATE KEY-----
`
)

type CozeApi struct {
	client xhttp.Client
	cfg    *config.LLMConfig
}

func toCozeMessages(messages []*base.MessageInput) ([]*CozeMessage, error) {
	oms := []*CozeMessage{}
	for _, msg := range messages {
		if msg.StringContent != "" {
			oms = append(oms, &CozeMessage{
				Role:        string(msg.Role),
				Content:     msg.StringContent,
				ContentType: "text",
			})
		} else if len(msg.MultiModelContents) > 0 {
			cozeObjs := []CozeObjectString{}
			for _, m := range msg.MultiModelContents {
				if m.Type == base.Image {
					cozeObjs = append(cozeObjs, CozeObjectStringImgObject(m.Content))
				} else if m.Type == base.Text {
					cozeObjs = append(cozeObjs, CozeObjectStringTextObject(m.Content))
				} else if m.Type == base.File {
					cozeObjs = append(cozeObjs, CozeObjectStringFileObject(m.Content))
				}
			}
			oms = append(oms, CozeMessageMultimodalMessage(cozeObjs))
		}
	}
	return oms, nil
}

func (o *CozeApi) ChatWithVarials(ctx context.Context, model string, messages []*base.MessageInput, variables map[string]any) (output base.Output, err error) {
	oms, err := toCozeMessages(messages)
	if err != nil {
		return output, err
	}
	chatResp, err := o.createChat(model, oms, variables)
	if err != nil {
		return output, err
	}
	if chatResp.Id == "" {
		xlog.Warnf("chat response is empty")
		return output, errors.New("chat response is empty")
	}
	for {
		status, err := o.queryChatStatus(chatResp)
		if err != nil {
			return output, err
		}
		if status == "failed" || status == "requires_action" || status == "canceled" {
			return output, fmt.Errorf("chat status err: %s", status)
		} else if status == "completed" {
			break
		} else {
			time.Sleep(1 * time.Second)
		}
	}
	text, err := o.extractText(chatResp)
	if err != nil {
		return output, err
	}
	output.Role = base.Assistant
	output.Content = text
	return output, errors.New("not implemented")
}

func (o *CozeApi) Chat(ctx context.Context, model string, messages []*base.MessageInput) (base.Output, error) {
	return o.ChatWithVarials(ctx, model, messages, nil)
}

func (o *CozeApi) StreamChat(ctx context.Context, model string, messages []*base.MessageInput) *ssestream.Stream[base.OutputChunk] {
	return (&base.OpenaiCompatiableMessageStream{
		OpenaiCompatiableStream: ssestream.NewStream[openai.ChatCompletionChunk](nil, errors.New("not implemented")),
	}).Stream()
}

func (o *CozeApi) Cfg() *config.LLMConfig {
	return o.cfg
}
