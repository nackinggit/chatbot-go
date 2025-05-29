package memory

import (
	"context"
	"errors"
	"strings"
	"time"

	xlog "com.imilair/chatbot/bootstrap/log"
	"com.imilair/chatbot/internal/model/dbmodel"
	"com.imilair/chatbot/internal/service"

	"com.imilair/chatbot/internal/service/config"
	"com.imilair/chatbot/internal/service/dao"
	"com.imilair/chatbot/pkg/embedding"
	"com.imilair/chatbot/pkg/llm/api/base"
	"com.imilair/chatbot/pkg/util"
	"com.imilair/chatbot/pkg/util/ttlmap"
)

type sessionManager struct {
	ticker     *time.Ticker
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
	if err := mcfg.Validate(); err != nil {
		xlog.Warnf("memory config error: %v", err)
		return err
	}
	innerManager.cfg = mcfg
	innerManager.eventModel = &base.LLMModel{}
	innerManager.ticker = time.NewTicker(30 * time.Minute)
	if _, err := embedding.GetEmbeddingApi(mcfg.LongMemory.EmbApi); err != nil {
		return err
	}
	ctx := context.Background()
	util.AsyncGoWithDefault(ctx, func() {
		xlog.Infof("start memory handler...")
		for range innerManager.ticker.C {
			keys := innerManager.sessions.Keys()
			for _, key := range keys {
				if session, ok := innerManager.sessions.Get(key); ok {
					s := session.(*Session)
					items := s.shortMemory.fetchForgotMemories(ctx)
					if len(items) > 20 {
						s.shortMemory.doForget(ctx)
						if s.longMemory != nil {
							s.longMemory.append(ctx, items)
						}
						t.summary(ctx, s, items)
					}
				}
			}
		}
	})
	return nil
}

func (t *sessionManager) summary(ctx context.Context, session *Session, items []*MemoryItems) {
	parse := func(str string) ([]*Event, error) {
		var events []*Event
		err := util.TryParseJsonArray(str, &events)
		return events, err
	}

	extractEvent := func(mis []*MemoryItems) (events []*Event, err error) {
		text := []string{}
		for _, mi := range mis {
			text = append(text, mi.ToString())
		}
		ms := []*base.MessageInput{
			base.SystemStringMessage(`你擅长从原始的对话日志中总结并提取重点事件，
提出出来的内容，以json的结构返回，参考下面的例子：
[{"where": <事件地点,非必填>, "event":<事件名称>, "todo":<代办项,非必填>, "emotional":<对话的情绪状态,非必填>}]
你每次提取的事件不超过5个。 
你只做事件提取不做别的事情。`),
			base.UserStringMessage(strings.Join(text, "\n")),
		}
		for i := 0; i < 3; i++ {
			output, ie := t.eventModel.Chat(ctx, ms)
			if ie != nil {
				return nil, ie
			}
			ret, ie := parse(output.Content)
			if ie == nil {
				return ret, nil
			}
			err = ie
		}
		return events, err
	}
	//todo 总结
	events, _ := extractEvent(items)
	timeStr := time.Now().Format("2006-01-02")
	if len(events) == 0 {
		userId := items[0].Memories[0].ImUserID
		botId := items[0].Memories[0].ImBotID
		ches := []*dbmodel.ChatHistoryEvent{}
		for _, event := range events {
			ches = append(ches, &dbmodel.ChatHistoryEvent{
				Userid:    int32(util.StringToInt64(userId)),
				Botid:     int32(util.StringToInt64(botId)),
				DateStr:   timeStr,
				Event:     event.Event,
				Addr:      event.Where,
				Todo:      event.Todo,
				Emotional: event.Emotional,
			})
		}
		dao.BatchInsert(ctx, &dbmodel.ChatHistoryEvent{}, ches)
	}
}

func (t *sessionManager) Stop() {
	innerManager.ticker.Stop()
}

func init() {
	service.Register(&sessionManager{})
}

func doGetSession(id string, persistant bool) *Session {
	if session, ok := innerManager.sessions.Get(id); ok {
		return session.(*Session)
	}
	shotM := NewShortMemory(id, innerManager.cfg.ShortMemory)
	session := &Session{
		ID:          id,
		shortMemory: shotM,
	}
	if persistant {
		longM, err := NewLongMemory(innerManager.cfg.LongMemory, id)
		if err != nil {
			xlog.Warnf("初始化长期记忆出错, error: %v", err)
			return nil
		}
		session.longMemory = longM
	}
	innerManager.sessions.Put(id, session)
	return session
}

func GetSession(ctx context.Context, id string) *Session {
	return doGetSession(id, true)
}

func GetTempSession(ctx context.Context, id string) *Session {
	return doGetSession(id, false)
}

type Session struct {
	ID          string
	longMemory  *LongMemory
	shortMemory *ShortMemory
	lastAccess  int64
}

func (session *Session) IsActiveBefore(duration time.Duration) bool {
	return session.lastAccess+int64(duration.Seconds()) > time.Now().Unix()
}

func (session *Session) SetSessionActive() {
	session.lastAccess = time.Now().Unix()
}

func (session *Session) AddMemory(ctx context.Context, memory *MemoryItems) error {
	if session == nil {
		return errors.New("session not found")
	}
	err := session.shortMemory.append(ctx, memory)
	if err != nil {
		return err
	}
	return nil
}

// FetchRelatedMemory 根据 sessionId 获取相关的记忆信息
func (session *Session) FetchRelatedMemory(ctx context.Context, input string, maxWords int) []*MemoryItems {
	if session == nil {
		xlog.Infof("session not found")
		return []*MemoryItems{}
	}
	var memories = []*MemoryItems{}
	shotMemories := session.shortMemory.fetch(ctx)
	curWords := 0
	for _, memory := range util.ReverseSlice(shotMemories) {
		if curWords >= maxWords {
			break
		} else {
			memories = append(memories, memory)
			curWords += memory.WordCount()
		}
	}
	if session.longMemory != nil {
		longMmemories := session.longMemory.fetchRelated(ctx, input)
		for _, memory := range util.ReverseSlice(longMmemories) {
			if curWords >= maxWords {
				break
			} else {
				memories = append(memories, memory)
				curWords += memory.WordCount()
			}
		}
	}
	return util.ReverseSlice(memories)
}
