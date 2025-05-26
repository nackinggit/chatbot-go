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
	for _, memory := range util.ReverseSlice(shotMemories) {
		if curWords >= maxWords {
			break
		} else {
			memories = append(memories, memory)
			curWords += memory.WordCount()
		}
	}
	for _, memory := range util.ReverseSlice(longMmemories) {
		if curWords >= maxWords {
			break
		} else {
			memories = append(memories, memory)
			curWords += memory.WordCount()
		}
	}
	return util.ReverseSlice(memories)
}
