package llm

import (
	"errors"
	"fmt"

	"com.imilair/chatbot/bootstrap/config"
	"com.imilair/chatbot/pkg/llm/api/base"
)

var models = map[string]base.LLMModel{}
var apis = map[string]base.LLMApi{}

var register = map[string]base.InitApi{}

func Init(cfgs map[string]*config.LLMConfig) error {
	for name, cfg := range cfgs {
		var initApi base.InitApi
		if cfg.OpenaiCompatiable {
			initApi = base.InitOpenaiCompatibleApi
		} else {
			initApi = register[name]
		}
		if initApi == nil {
			return fmt.Errorf("api %s not registered", name)
		}
		llmapi := initApi(cfg)
		apis[name] = llmapi
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

func GetModel(name string) (base.LLMModel, error) {
	if model, ok := models[name]; ok {
		return model, nil
	} else {
		return base.LLMModel{}, errors.New("model not found")
	}
}
