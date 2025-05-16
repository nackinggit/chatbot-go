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

var xclient *XMilvus

func Init(cfg *config.MilvusConfig) error {
	client, err := client.NewClient(context.Background(), client.Config{
		Address:  cfg.Address,
		Username: cfg.Username,
		Password: cfg.Password,
	})
	if err != nil {
		return err
	}
	xclient = &XMilvus{
		client: client,
		cfg:    cfg,
	}
	return nil
}

func (x *XMilvus) Close() {
	if x.client != nil {
		xlog.Infof("Close Milvus client: %s", x.cfg.Address)
		x.client.Close()
	}
}

func GetClient() *XMilvus {
	return xclient
}
