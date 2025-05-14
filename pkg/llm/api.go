package llm

import (
	"fmt"

	"com.imilair/chatbot/bootstrap/config"
	"com.imilair/chatbot/pkg/llm/api/base"
	"com.imilair/chatbot/pkg/llm/api/doubao"
)

var models = map[string]base.LLMModel{}

var register = map[string]base.InitApi{
	"doubao": doubao.InitApi,
}

func Init(cfgs map[string]*config.LLMConfig) error {
	for name, cfg := range cfgs {
		initApi := register[name]
		if initApi == nil {
			return fmt.Errorf("api %s not registered", name)
		}
		llmapi := initApi(cfg)
		for _, model := range cfg.Models {
			models[model.Name] = base.LLMModel{
				Api:   llmapi,
				Name:  model.Name,
				Model: model.Model,
			}
		}
	}
	return nil
}
