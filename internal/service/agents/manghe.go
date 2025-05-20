package agents

import (
	xlog "com.imilair/chatbot/bootstrap/log"
	"com.imilair/chatbot/internal/service"
	"github.com/gin-gonic/gin"
)

var MangHeService *manghe

type manghe struct {
	imageAnalyseModel *AgentModel
	predictModel      *AgentModel
}

func (t *manghe) Name() string {
	return "manghe"
}

func (t *manghe) Init() (err error) {
	xlog.Infof("init service `%s`", t.Name())
	mangheCfg := service.Config.MangHe
	err = mangheCfg.Validate()
	if err != nil {
		return err
	}
	t.imageAnalyseModel, err = initModel(mangheCfg.ImageAnalyse)
	if err != nil {
		return err
	}
	t.predictModel, err = initModel(mangheCfg.Predict)
	if err != nil {
		return err
	}

	MangHeService = t
	xlog.Infof("`%s` inited", t.Name())
	return nil
}

func init() {
	service.Register(&teacher{})
}

func (m *manghe) ImageAnalyse(ctx *gin.Context, imgUrl string) {

}
