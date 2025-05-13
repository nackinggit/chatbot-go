package base

import (
	"errors"
	"fmt"

	"github.com/openai/openai-go"
)

type MessageOutput struct {
	ReasoningContent string `json:"reasoning"`
	Content          string `json:"content"`
	Role             string `json:"role"`
}

type MessageInput struct {
	StringContent      string         `json:"content,omitempty"`            // 文本input
	MultiModelContents []InputContent `json:"multiModelContents,omitempty"` // 多模态inputs
	Role               string         `json:"role"`
}

type InputContent struct {
	Type    InputType `json:"type"`
	Content string    `json:"content"`
}

type InputType string

const (
	Text  InputType = "text"
	Image InputType = "image"
)

func (input MessageInput) ToOpenaiMessage() (m openai.ChatCompletionMessageParamUnion, err error) {
	r, err := convertInputContent(input)
	if err != nil {
		return m, err
	}

	switch input.Role {
	case "user":
		if v, ok := r.(string); ok {
			m = openai.UserMessage(v)
		} else if v, ok := r.([]openai.ChatCompletionContentPartUnionParam); ok {
			m = openai.UserMessage(v)
		}
	case "assistant":
		if v, ok := r.(string); ok {
			m = openai.AssistantMessage(v)
		} else if _, ok := r.([]openai.ChatCompletionContentPartUnionParam); ok {
			err = errors.New("assistant message cannot contain multiple content parts")
		}
	case "system":
		if v, ok := r.(string); ok {
			m = openai.SystemMessage(v)
		} else if v, ok := r.([]openai.ChatCompletionContentPartUnionParam); ok {
			tv := make([]openai.ChatCompletionContentPartTextParam, len(v))
			for i, content := range v {
				if content.OfText != nil {
					tv[i] = *content.OfText
				} else {
					err = errors.New("system message cannot contain content parts other than text")
					break
				}
			}
			m = openai.SystemMessage(tv)
		}
	default:
		err = fmt.Errorf("unknown message role: %s", input.Role)
	}
	return m, err
}

func convertInputContent(input MessageInput) (output any, err error) {
	if input.StringContent != "" {
		output = input.StringContent
	} else if len(input.MultiModelContents) > 0 {
		var openaiContents []openai.ChatCompletionContentPartUnionParam
		for _, content := range input.MultiModelContents {
			switch content.Type {
			case Text:
				openaiContents = append(openaiContents, openai.TextContentPart(content.Content))
			case Image:
				openaiContents = append(openaiContents, openai.ImageContentPart(openai.ChatCompletionContentPartImageImageURLParam{
					URL: content.Content,
				}))
			default:
				err = fmt.Errorf("invalid input content type: %s", content.Type)
			}
			output = openaiContents
		}
	} else {
		err = fmt.Errorf("invalid input content type: %v", input)
	}
	return output, err
}
