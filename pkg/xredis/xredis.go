package xredis

import (
	"fmt"

	"com.imilair/chatbot/bootstrap/config"
	xlog "com.imilair/chatbot/bootstrap/log"
	"com.imilair/chatbot/pkg/util"
	"github.com/redis/go-redis/v9"
)

type XRedisClient struct {
	cfg *config.RedisConfig
	*redis.Client
}

var redisclient *XRedisClient

func Init(cfg *config.RedisConfig) {
	xlog.Infof("init redis: %s", util.JsonString(cfg))
	opts := redis.Options{
		Addr:     fmt.Sprintf("%s:%d", cfg.Host, cfg.Port),
		Password: cfg.Password,
		Username: cfg.Username,
		DB:       cfg.DB,
	}
	redisclient = &XRedisClient{
		cfg:    cfg,
		Client: redis.NewClient(&opts),
	}
}

func ExecRedisCmd[T any](fn func(mr *XRedisClient) (T, error)) (T, error) {
	resp, err := fn(redisclient)
	if err != nil {
		xlog.Warnf("ExecRedisCmd error: %v", err)
		return resp, err
	}
	return resp, nil
}
