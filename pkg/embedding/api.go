package embedding

import (
	"fmt"

	"com.imilair/chatbot/bootstrap/config"
	"com.imilair/chatbot/pkg/embedding/apis/base"
	"com.imilair/chatbot/pkg/embedding/apis/doubao"
)

var apis = map[string]base.EmbeddingApi{}
var register = map[string]base.InitApi{}

func init() {
	register["doubao"] = doubao.InitApi
}

func Init(cfgs map[string]*config.EmbeddingConfig) error {
	for _, cfg := range cfgs {
		if _, ok := register[cfg.RegisterService]; !ok {
			return fmt.Errorf("embedding api not found: %s", cfg.RegisterService)
		}
		apis[cfg.RegisterService] = register[cfg.RegisterService](cfg)
	}
	return nil
}
