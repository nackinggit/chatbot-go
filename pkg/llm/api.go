package llm

import (
	"fmt"

	"com.imilair/chatbot/bootstrap/config"
	xlog "com.imilair/chatbot/bootstrap/log"
	"com.imilair/chatbot/pkg/llm/api/base"
	"com.imilair/chatbot/pkg/util"
)

var models = map[string]base.LLMModel{}
var apis = map[string]base.LLMApi{}

var register = map[string]base.InitApi{}

func Init(cfgs []*config.LLMConfig) error {
	xlog.Infof("init llm %s", util.JsonString(cfgs))
	for _, cfg := range cfgs {
		var initApi base.InitApi
		if cfg.OpenaiCompatiable {
			initApi = base.InitOpenaiCompatibleApi
		} else {
			initApi = register[cfg.RegisterService]
		}
		if initApi == nil {
			return fmt.Errorf("api %s not registered", cfg.Name)
		}
		llmapi := initApi(cfg)
		apis[cfg.RegisterService] = llmapi
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

func GetModel(name string) (*base.LLMModel, error) {
	if model, ok := models[name]; ok {
		return &model, nil
	} else {
		return nil, fmt.Errorf("llm model %s not found", name)
	}
}

func GetModels(names []string) ([]*base.LLMModel, error) {
	llmmoldes := make([]*base.LLMModel, 0)
	for _, name := range names {
		if model, ok := models[name]; ok {
			llmmoldes = append(llmmoldes, &model)
		} else {
			return nil, fmt.Errorf("llm model %s not found", name)
		}
	}
	return llmmoldes, nil
}
