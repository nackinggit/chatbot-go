package llm

import (
	"fmt"

	"com.imilair/chatbot/bootstrap/config"
	xlog "com.imilair/chatbot/bootstrap/log"
	"com.imilair/chatbot/pkg/llm/api/base"
	"com.imilair/chatbot/pkg/llm/api/coze"
	"com.imilair/chatbot/pkg/util"
)

var apis = map[string]base.LLMApi{}

var register = map[string]base.InitApi{
	"coze": coze.InitApi,
}

func Init(cfgs []*config.LLMConfig) error {
	xlog.Infof("init llm %s", util.BeautifulJson(cfgs))
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
	}
	return nil
}

func GetApi(api string) (base.LLMApi, error) {
	llmapi, ok := apis[api]
	if !ok {
		return nil, fmt.Errorf("llm api %s not found", api)
	}
	return llmapi, nil
}
