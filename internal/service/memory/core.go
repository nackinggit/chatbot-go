package memory

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	xlog "com.imilair/chatbot/bootstrap/log"
	"com.imilair/chatbot/internal/model/dbmodel"
	"com.imilair/chatbot/internal/service/config"
	"com.imilair/chatbot/internal/service/dao"
	"com.imilair/chatbot/pkg/embedding"
	"com.imilair/chatbot/pkg/embedding/apis/base"
	"com.imilair/chatbot/pkg/util"
	"com.imilair/chatbot/pkg/xmilvus"
	"com.imilair/chatbot/pkg/xredis"
	"github.com/milvus-io/milvus/client/v2/entity"
	"github.com/milvus-io/milvus/client/v2/milvusclient"
	"github.com/redis/go-redis/v9"
)

type Event struct {
	Where     string `json:"where"`
	Event     string `json:"event"`
	Todo      string `json:"todo"`
	Emotional string `json:"emotional"`
}

type MemoryItems struct {
	CreateTime int64
	Memories   []*dbmodel.LlmChatHistory
	Vector     *base.Vector
	Sid        string
}

func (mi *MemoryItems) ToString() string {
	strs := []string{}
	for _, h := range mi.Memories {
		strs = append(strs, h.Message)
	}
	return strings.Join(strs, "\n")
}

func (mi *MemoryItems) WordCount() int {
	c := 0
	for _, h := range mi.Memories {
		c += len([]rune(h.Message))
	}
	return c
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

func (sm *ShortMemory) fetch(ctx context.Context) []*MemoryItems {
	memories, _ := xredis.ExecRedisCmd(func(mr *xredis.XRedisClient) ([]*MemoryItems, error) {
		vals, err := mr.ZRangeByScore(ctx, sm.SessionId, &redis.ZRangeBy{
			Max: fmt.Sprintf("%d", time.Now().UnixMilli()),
			Min: fmt.Sprintf("%d", time.Now().Add(-time.Duration(sm.ReserveTTL)*time.Second).UnixMilli()),
		}).Result()
		if err != nil {
			return []*MemoryItems{}, err
		}
		items := make([]*MemoryItems, len(vals))
		for i, val := range vals {
			err = util.Unmarshal([]byte(val), &items[i])
			if err != nil {
				return []*MemoryItems{}, err
			}
		}
		return items, nil
	})
	return memories
}

// 获取遗忘的记忆列表
func (sm *ShortMemory) fetchForgotMemories(ctx context.Context) []*MemoryItems {
	memories, _ := xredis.ExecRedisCmd(func(mr *xredis.XRedisClient) ([]*MemoryItems, error) {
		vals, err := mr.ZRangeByScore(ctx, sm.SessionId, &redis.ZRangeBy{
			Min: "-inf",
			Max: fmt.Sprintf("%d", time.Now().Add(-time.Duration(sm.ReserveTTL)*time.Second).UnixMilli()),
		}).Result()
		if err != nil {
			return nil, err
		}
		memories := make([]*MemoryItems, len(vals))
		for i, v := range vals {
			memories[i] = &MemoryItems{}
			if err := util.Unmarshal([]byte(v), memories[i]); err != nil {
				return nil, err
			}
		}
		return memories, nil
	})
	return memories
}

func (sm *ShortMemory) doForget(ctx context.Context) {
	xredis.ExecRedisCmd(func(mr *xredis.XRedisClient) (any, error) {
		return mr.ZRemRangeByScore(ctx, sm.SessionId,
			"-inf",
			fmt.Sprintf("%d", time.Now().Add(-time.Duration(sm.ReserveTTL)*time.Second).UnixMilli()),
		).Result()
	})
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
	if err != nil {
		return err
	}
	hiss := []*dbmodel.LlmChatHistory{}
	for _, item := range memoryItems {
		hiss = append(hiss, item.Memories...)
	}
	return dao.BatchInsert(ctx, hiss[0], hiss)
}

func (lm *LongMemory) fetchRelated(ctx context.Context, input string) []*MemoryItems {
	topK := 40
	vetors, err := lm.embeddingApi.DoEmbedding(ctx, lm.emmodel, []string{input}, lm.dim)
	if err != nil {
		xlog.Warnf("embeddingApi.DoEmbedding: %v", err)
		return []*MemoryItems{}
	}
	vetor := vetors[0]
	searchOpt := milvusclient.NewSearchOption(vetor.CollectionName(), topK, []entity.Vector{entity.FloatVector(vetor.Vector)}).
		WithANNSField("embedding").WithOutputFields("mid")
	ret, err := xmilvus.GetClient().Search(ctx, searchOpt)
	if err != nil {
		xlog.Warnf("milvusclient.Search err: %v", err)
		return []*MemoryItems{}
	}
	sids := []string{}
	for _, result := range ret {
		ms := result.GetColumn("mid").FieldData().GetScalars()
		if ms != nil {
			sids = append(sids, ms.GetStringData().Data...)
		}
	}
	if len(sids) == 0 {
		return []*MemoryItems{}
	}
	hiss, err := dao.LlmModelDao.QueryChathistoryBySids(ctx, sids)
	if err != nil {
		xlog.Warnf("QueryChathistoryBySids err: %v", err)
		return []*MemoryItems{}
	}
	items := []*MemoryItems{}
	tmap := map[string]*MemoryItems{}
	for _, his := range hiss {
		item := tmap[his.Sid]
		if item == nil {
			item = &MemoryItems{
				Sid:        his.Sid,
				Memories:   []*dbmodel.LlmChatHistory{his},
				CreateTime: his.CreatedAt.UnixMilli(),
			}
			tmap[his.Sid] = item
			items = append(items, item)
		} else {
			item.Memories = append(item.Memories, his)
		}
	}
	return items
}
