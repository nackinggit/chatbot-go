package xmilvus

import (
	"context"

	"com.imilair/chatbot/bootstrap/config"
	xlog "com.imilair/chatbot/bootstrap/log"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
)

type XMilvus struct {
	*milvusclient.Client
	cfg *config.MilvusConfig
}

var xclient *XMilvus

func Init(cfg *config.MilvusConfig) error {
	xlog.Infof("Init Milvus client: %s", cfg.Address)
	client, err := milvusclient.New(context.Background(), &milvusclient.ClientConfig{
		Address:  cfg.Address,
		Username: cfg.Username,
		Password: cfg.Password,
	})
	if err != nil {
		return err
	}
	xclient = &XMilvus{
		Client: client,
		cfg:    cfg,
	}
	return nil
}

func (x *XMilvus) Close() {
	if x.Client != nil {
		xlog.Infof("Close Milvus client: %s", x.cfg.Address)
		x.Close()
	}
}

func GetClient() *XMilvus {
	return xclient
}
