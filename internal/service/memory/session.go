package memory

import (
	"context"
	"errors"
	"time"

	xlog "com.imilair/chatbot/bootstrap/log"
	"com.imilair/chatbot/internal/service"

	"com.imilair/chatbot/internal/service/config"
	"com.imilair/chatbot/pkg/embedding"
	"com.imilair/chatbot/pkg/llm/api/base"
	"com.imilair/chatbot/pkg/util"
	"com.imilair/chatbot/pkg/util/ttlmap"
)

type sessionManager struct {
	endflag    chan bool
	sessions   *ttlmap.TTLMap
	cfg        *config.MemoryConfig
	eventModel *base.LLMModel
}

var innerManager *sessionManager

func (t *sessionManager) Name() string {
	return "service.memory"
}

func (t *sessionManager) InitAndStart() (err error) {
	innerManager = &sessionManager{
		sessions: ttlmap.New(1000, 7200),
	}
	mcfg := service.Config.Memory
	innerManager.cfg = mcfg
	innerManager.endflag = make(chan bool)
	innerManager.eventModel = &base.LLMModel{}
	if _, err := embedding.GetEmbeddingApi(mcfg.LongMemory.EmbApi); err != nil {
		return err
	}
	ctx := context.Background()
	util.AsyncGoWithDefault(ctx, func() {
		xlog.Infof("start memory handler...")
		for {
			select {
			case <-innerManager.endflag:
				return
			default:
				keys := innerManager.sessions.Keys()
				for _, key := range keys {
					if session, ok := innerManager.sessions.Get(key); ok {
						s := session.(*Session)
						items := s.shortMemory.fetchForgotMemories(ctx)
						if len(items) > 20 {
							s.shortMemory.doForget(ctx)
							s.longMemory.append(ctx, items)
							t.summary(ctx, s, items)
						}
					}
				}
			}
			time.Sleep(30 * time.Minute)
		}
	})
	return nil
}

func (t *sessionManager) summary(ctx context.Context, session *Session, items []*MemoryItems) {
	//todo 总结
}

func (t *sessionManager) Stop() {
	innerManager.endflag <- true
}

func init() {
	service.Register(&sessionManager{})
}

func getSession(id string) *Session {
	if session, ok := innerManager.sessions.Get(id); ok {
		return session.(*Session)
	}

	longM, err := NewLongMemory(innerManager.cfg.LongMemory, id)
	if err != nil {
		xlog.Warnf("初始化长期记忆出错, error: %v", err)
		return nil
	}
	shotM := NewShortMemory(id, innerManager.cfg.ShortMemory)
	session := &Session{
		ID:          id,
		longMemory:  longM,
		shortMemory: shotM,
	}
	innerManager.sessions.Put(id, session)
	return session
}

type Session struct {
	ID          string
	longMemory  *LongMemory
	shortMemory *ShortMemory
}

func AddMemory(ctx context.Context, sessionId string, memory *MemoryItems) error {
	session := getSession(sessionId)
	if session == nil {
		xlog.Infof("session not found: %s", sessionId)
		return errors.New("session not found")
	}
	return session.shortMemory.append(ctx, memory)
}

// FetchRelatedMemory 根据 sessionId 获取相关的记忆信息
func FetchRelatedMemory(ctx context.Context, sessionId string, input string, maxWords int) []*MemoryItems {
	session := getSession(sessionId)
	if session == nil {
		xlog.Infof("session not found: %s", sessionId)
		return []*MemoryItems{}
	}
	shotMemories := session.shortMemory.fetch(ctx)
	longMmemories := session.longMemory.fetchRelated(ctx, input)
	var memories = []*MemoryItems{}
	curWords := 0
	for _, memory := range shotMemories {
		if curWords >= maxWords {
			break
		} else {
			memories = append(memories, memory)
			curWords += memory.WordCount()
		}
	}
	for _, memory := range longMmemories {
		if curWords >= maxWords {
			break
		} else {
			memories = append(memories, memory)
			curWords += memory.WordCount()
		}
	}
	return memories
}
