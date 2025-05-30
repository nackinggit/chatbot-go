package xvc

import (
	"context"
	"fmt"

	"com.imilair/chatbot/bootstrap/config"
	xlog "com.imilair/chatbot/bootstrap/log"
	"com.imilair/chatbot/internal/bcode"
	"com.imilair/chatbot/pkg/util"
	"github.com/volcengine/volc-sdk-golang/service/visual"
)

var instance *visual.Visual

type EntitySegmentResponse struct {
	Foreground     string `json:"foreground"`     // 前景图
	ForegroundMask string `json:"foregroundMask"` // 前景图的mask
	Fromat         string `json:"format"`         // 格式
}

func Init(cfg *config.VolceEngineConfig) {
	instance = visual.NewInstance()
	instance.Client.SetAccessKey(cfg.Ak)
	instance.Client.SetSecretKey(cfg.Sk)
}

func EntitySegment(ctx context.Context, url string) (*EntitySegmentResponse, error) {
	if instance == nil {
		return nil, bcode.New(500, "xvc not inited")
	}
	res, c, err := instance.EntitySegment(map[string]interface{}{
		"req_key":       "entity_seg",
		"image_urls":    []string{url},
		"return_format": 3,
		"refine_mask":   1,
	})
	if err != nil {
		return nil, err
	}
	if c != 200 || res.Data == nil || len(res.Data.BinaryDataBase64) <= 0 {
		xlog.InfoC(ctx, "分割失败: %v", util.JsonString(res))
		return nil, fmt.Errorf("分割失败: %v-%v", res.Code, res.Message)
	}
	ib64s := res.Data.BinaryDataBase64
	res.Data.BinaryDataBase64 = []string{"base64 image array..."}
	xlog.Infof("结果：code: %v, err: %v, res: %v", c, err, util.JsonString(res))
	return &EntitySegmentResponse{Foreground: ib64s[0], ForegroundMask: ib64s[1], Fromat: "png"}, nil
}
