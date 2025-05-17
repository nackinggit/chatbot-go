package base

import (
	"bytes"
	"fmt"

	xlog "com.imilair/chatbot/bootstrap/log"
	"com.imilair/chatbot/pkg/util"
	"github.com/openai/openai-go"
	"github.com/openai/openai-go/packages/ssestream"
)

type MessageRole string

const (
	USER      MessageRole = "user"
	SYSTEM    MessageRole = "system"
	Assistant MessageRole = "assistant"
)

type MessageInput struct {
	StringContent      string         `json:"content,omitempty"`            // 文本input
	MultiModelContents []InputContent `json:"multiModelContents,omitempty"` // 多模态inputs
	Role               MessageRole    `json:"role"`
}

func UserStringMessage(content string) MessageInput {
	return MessageInput{
		StringContent: content,
		Role:          USER,
	}
}

func (input *MessageInput) ToOpenaiMessage() (res openai.ChatCompletionMessageParamUnion, err error) {
	role := input.Role
	if role == USER {
		return input.openaiUserMessage()
	} else if role == SYSTEM {
		return input.openaiSystemMessage()
	} else if role == Assistant {
		return input.openaiAssistantMessage()
	}
	return res, fmt.Errorf("unknown message role: %s", role)
}

func (input *MessageInput) openaiSystemMessage() (res openai.ChatCompletionMessageParamUnion, err error) {
	if input.StringContent != "" {
		res = openai.SystemMessage(input.StringContent)
	} else {
		err = fmt.Errorf("invalid system message input: %v", input)
	}
	return res, err
}

func (input *MessageInput) openaiAssistantMessage() (res openai.ChatCompletionMessageParamUnion, err error) {
	if input.StringContent != "" {
		res = openai.AssistantMessage(input.StringContent)
	} else {
		err = fmt.Errorf("invalid assistant message input: %v", input)
	}
	return res, err
}

func (input *MessageInput) openaiUserMessage() (res openai.ChatCompletionMessageParamUnion, err error) {
	if input.StringContent != "" {
		res = openai.UserMessage(input.StringContent)
	} else if len(input.MultiModelContents) > 0 {
		ms := []openai.ChatCompletionContentPartUnionParam{}
		for _, v := range input.MultiModelContents {
			switch v.Type {
			case Text:
				ms = append(ms, openai.TextContentPart(v.Content))
			case Image:
				ms = append(ms, openai.ImageContentPart(openai.ChatCompletionContentPartImageImageURLParam{
					URL: v.Content,
				}))
			}
		}
		res = openai.UserMessage(ms)
	} else {
		err = fmt.Errorf("invalid message input: %v", input)
	}
	return res, err
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

type OpenaiCompatiableMessageOutput struct {
	OpenaiChatCompletion *openai.ChatCompletion
}

type OpenaiChatCompletionMessage struct {
	openai.ChatCompletionMessage
	ReasoningContent string `json:"reasoning_content"`
}

func (output *OpenaiCompatiableMessageOutput) MessageOutput() (res Output, err error) {
	c := output.OpenaiChatCompletion
	resp := c.Choices[0].Message.RawJSON()
	var message OpenaiChatCompletionMessage
	err = util.Unmarshal([]byte(resp), &message)
	if err != nil {
		xlog.Warnf("Unmarshal: %v", err)
	} else {
		res = Output{
			Content:          message.Content,
			ReasoningContent: message.ReasoningContent,
			Role:             MessageRole(message.Role),
			RawJson:          c.RawJSON(),
		}
	}

	return res, err
}

type OpenaiCompatiableMessageStream struct {
	OpenaiCompatiableStream *ssestream.Stream[openai.ChatCompletionChunk]
}

func (stream *OpenaiCompatiableMessageStream) Stream() *ssestream.Stream[OutputChunk] {
	ostream := stream.OpenaiCompatiableStream
	return ssestream.NewStream[OutputChunk](&openaiDecoder{
		ostream: ostream,
	}, ostream.Err())
}

type openaiDecoder struct {
	ostream *ssestream.Stream[openai.ChatCompletionChunk]
	evt     ssestream.Event
	err     error
}

func (s *openaiDecoder) Next() bool {
	data := bytes.NewBuffer(nil)
	if s.ostream.Next() {
		cur := s.ostream.Current()
		delta := cur.Choices[0].Delta
		var output OutputChunk
		s.err = util.Unmarshal([]byte(delta.RawJSON()), &output)
		output.RawJSON = cur.RawJSON()
		value, _ := util.Marshal(output)
		_, s.err = data.Write(value)
		s.evt = ssestream.Event{
			Data: data.Bytes(),
		}
		return true
	}

	if s.ostream.Err() != nil {
		s.err = s.ostream.Err()
	}
	return false
}

func (s *openaiDecoder) Event() ssestream.Event {
	return s.evt
}

func (s *openaiDecoder) Close() error {
	return s.ostream.Close()
}

func (s *openaiDecoder) Err() error {
	return s.err
}
