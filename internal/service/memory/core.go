package memory

import (
	"context"
	"errors"
	"fmt"
	"time"

	"com.imilair/chatbot/internal/model/dbmodel"
	"com.imilair/chatbot/internal/service/config"
	"com.imilair/chatbot/pkg/embedding"
	"com.imilair/chatbot/pkg/embedding/apis/base"
	"com.imilair/chatbot/pkg/util"
	"com.imilair/chatbot/pkg/xmilvus"
	"com.imilair/chatbot/pkg/xredis"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"github.com/redis/go-redis/v9"
)

type MemoryItems struct {
	CreateTime int64
	Memories   []*dbmodel.LlmChatHistory
	Vector     *base.Vector
	Sid        string
}

func (mi *MemoryItems) ToString() string {
	return fmt.Sprintf("[%v]", mi.Memories)
}

type ShortMemory struct {
	SessionId  string // 会话ID
	ReserveTTL int    // 保留时长，单位：秒
	MaxSize    int    // 最大长度
}

func NewShortMemory(sessionId string, cfg *config.ShortMemoryConfig) *ShortMemory {
	return &ShortMemory{
		SessionId:  sessionId,
		ReserveTTL: cfg.TTL,
		MaxSize:    cfg.MaxSize,
	}
}

// 添加记忆
func (sm *ShortMemory) append(ctx context.Context, memoryItems *MemoryItems) error {
	_, err := xredis.ExecRedisCmd(func(mr *xredis.XRedisClient) (any, error) {
		return mr.ZAdd(ctx, sm.SessionId, redis.Z{
			Member: util.JsonString(memoryItems),
			Score:  float64(time.Now().UnixMilli()),
		}).Result()
	})
	return err
}

// 获取遗忘的记忆列表
func (sm *ShortMemory) fetchForgotMemories(ctx context.Context) []*MemoryItems {
	memories, _ := xredis.ExecRedisCmd(func(mr *xredis.XRedisClient) ([]*MemoryItems, error) {
		vals, err := mr.ZRangeByScore(ctx, sm.SessionId, &redis.ZRangeBy{
			Min: "-inf",
			Max: fmt.Sprintf("%d", time.Now().Add(-time.Duration(sm.ReserveTTL)*time.Second).Unix()),
		}).Result()
		if err != nil {
			return nil, err
		}
		memories := make([]*MemoryItems, len(vals))
		for i, v := range vals {
			mr.ZRem(ctx, sm.SessionId, v)
			memories[i] = &MemoryItems{}
			if err := util.Unmarshal([]byte(v), memories[i]); err != nil {
				return nil, err
			}
		}
		return memories, nil
	})
	return memories
}

type LongMemory struct {
	embeddingApi base.EmbeddingApi
	emmodel      string
	dim          int
	sessionId    string
}

func NewLongMemory(cfg *config.LongMemoryConfig, sessionId string) (*LongMemory, error) {
	api, err := embedding.GetEmbeddingApi(cfg.EmbApi)
	if err != nil {
		return nil, err
	}
	return &LongMemory{
		embeddingApi: api,
		emmodel:      cfg.EmbModel,
		dim:          cfg.VectorDim,
		sessionId:    sessionId,
	}, nil
}

func (lm *LongMemory) append(ctx context.Context, memoryItems []*MemoryItems) error {
	vecStrings := []string{}
	for _, item := range memoryItems {
		vecStrings = append(vecStrings, item.ToString())
	}
	vectors, err := lm.embeddingApi.DoEmbedding(ctx, lm.emmodel, vecStrings, lm.dim)
	if err != nil {
		return err
	}
	if len(memoryItems) != len(vectors) {
		return errors.New("vector length not match")
	}

	mids := []string{}
	pkeys := []string{}
	vects := [][]float32{}
	for i, item := range memoryItems {
		item.Vector = &vectors[i]
		mids = append(mids, item.Sid)
		pkeys = append(pkeys, lm.sessionId)
		vects = append(vects, vectors[i].Vector)
	}
	collname := vectors[0].CollectionName()
	insertOpts := milvusclient.NewColumnBasedInsertOption(collname).
		WithVarcharColumn("mid", mids).
		WithVarcharColumn("pkey", pkeys).
		WithFloatVectorColumn("embedding", lm.dim, vects)
	_, err = xmilvus.GetClient().Insert(ctx, insertOpts)
	return err
}
