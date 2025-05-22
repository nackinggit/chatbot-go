package embedding

import (
	"fmt"

	"com.imilair/chatbot/bootstrap/config"
	xlog "com.imilair/chatbot/bootstrap/log"
	"com.imilair/chatbot/pkg/embedding/apis/base"
	"com.imilair/chatbot/pkg/embedding/apis/doubao"
)

var apis = map[string]base.EmbeddingApi{}
var register = map[string]base.InitApi{}

func init() {
	register["doubao"] = doubao.InitApi
}

func Init(cfgs []*config.EmbeddingConfig) error {
	xlog.Infof("init embedding apis ...")
	for _, cfg := range cfgs {
		if _, ok := register[cfg.RegisterService]; !ok {
			return fmt.Errorf("embedding api not found: %s", cfg.RegisterService)
		}
		apis[cfg.RegisterService] = register[cfg.RegisterService](cfg)
	}
	return nil
}

func GetEmbeddingApi(api string) (base.EmbeddingApi, error) {
	if eapi, ok := apis[api]; ok {
		return eapi, nil
	}
	return nil, fmt.Errorf("embedding api `%s` not found", api)
}
