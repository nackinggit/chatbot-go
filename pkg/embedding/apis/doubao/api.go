package doubao

import (
	"context"
	"fmt"
	"math"
	"slices"

	"com.imilair/chatbot/bootstrap/config"
	xlog "com.imilair/chatbot/bootstrap/log"
	"com.imilair/chatbot/pkg/embedding/apis/base"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime"
	"github.com/volcengine/volcengine-go-sdk/service/arkruntime/model"
)

type DouBaoEmbedding struct {
	cfg    *config.EmbeddingConfig
	client *arkruntime.Client
}

func InitApi(cfg *config.EmbeddingConfig) base.EmbeddingApi {
	return &DouBaoEmbedding{
		cfg:    cfg,
		client: arkruntime.NewClientWithApiKey(cfg.ApiKey, arkruntime.WithRetryTimes(cfg.MaxRetries)),
	}
}

func (d *DouBaoEmbedding) supporModel(model string, dims int) bool {
	for _, m := range d.cfg.Models {
		if m.Model == model {
			return slices.Contains(m.Dims, dims)
		}
	}
	return false
}

func (d *DouBaoEmbedding) DoEmbedding(ctx context.Context, emmodel string, texts []string, dim int) ([]base.Vector, error) {
	if !d.supporModel(emmodel, dim) {
		xlog.Warnf("%s 不支持Embedding模型：%s", d.cfg.Company, emmodel)
		return nil, fmt.Errorf("%s 不支持Embedding模型：%s", d.cfg.Company, emmodel)
	}
	req := model.EmbeddingRequestStrings{
		Input: texts,
		Model: emmodel,
	}
	resp, err := d.client.CreateEmbeddings(ctx, req)
	if err != nil {
		return nil, err
	}
	vectors := []base.Vector{}
	for _, v := range resp.Data {
		vector := slicedNormL2(v.Embedding, dim)
		vectors = append(vectors, base.Vector{
			Vector:  vector,
			Model:   emmodel,
			Company: d.cfg.Company,
			Dim:     len(vector),
		})
	}
	return vectors, nil
}

func slicedNormL2(v []float32, dim int) []float32 {
	norm := 0.0
	for _, v := range v[:dim] {
		norm += float64(v * v)
	}
	nv := make([]float32, dim)
	norm = math.Sqrt(norm)
	for i, v := range v[:dim] {
		nv[i] = float32(float64(v) / norm)
	}
	return nv
}
