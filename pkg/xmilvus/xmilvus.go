package xmilvus

import (
	"context"

	"com.imilair/chatbot/bootstrap/config"
	xlog "com.imilair/chatbot/bootstrap/log"
	"github.com/milvus-io/milvus-sdk-go/v2/client"
)

type XMilvus struct {
	client client.Client
	cfg    *config.MilvusConfig
}

func NewXMilvus(cfg *config.MilvusConfig) (*XMilvus, error) {
	client, err := client.NewClient(context.Background(), client.Config{
		Address:  cfg.Address,
		Username: cfg.Username,
		Password: cfg.Password,
	})
	if err != nil {
		return nil, err
	}
	return &XMilvus{
		client: client,
		cfg:    cfg,
	}, nil
}

func (x *XMilvus) Close() {
	if x.client != nil {
		xlog.Infof("Close Milvus client: %s", x.cfg.Address)
		x.client.Close()
	}
}
