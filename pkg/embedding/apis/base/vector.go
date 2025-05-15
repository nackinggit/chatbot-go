package base

import (
	"context"

	"com.imilair/chatbot/bootstrap/config"
)

type Vector struct {
	Model   string    `json:"model"`
	Company string    `json:"company"`
	Vector  []float32 `json:"vector"`
	Dim     int       `json:"dim"`
}

type InitApi func(cfg *config.EmbeddingConfig) EmbeddingApi

type EmbeddingApi interface {
	DoEmbedding(ctx context.Context, model string, texts []string, dims int) ([]Vector, error)
}
