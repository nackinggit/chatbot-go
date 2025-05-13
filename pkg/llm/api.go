package llm

import (
	"fmt"

	"com.imilair/chatbot/bootstrap/config"
	"com.imilair/chatbot/pkg/llm/api/base"
	"com.imilair/chatbot/pkg/llm/api/doubao"
)

var apis = map[string]base.LLMApi{}
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
		apis[name] = initApi(cfg)
	}
	return nil
}
