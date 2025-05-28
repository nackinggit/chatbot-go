package service

import (
	"errors"
	"sync"

	xlog "com.imilair/chatbot/bootstrap/log"
	"com.imilair/chatbot/internal/service/config"
)

var Config = config.ServiceConfig{}

type ServiceI interface {
	Name() string
	InitAndStart() error
	Stop()
}

type service struct {
	once   sync.Once
	inited bool
	is     ServiceI
}

func NewService(si ServiceI) *service {
	return &service{
		is: si,
	}
}

var reigstered = sync.Map{}

func Init() error {
	var err error
	reigstered.Range(func(key, value any) bool {
		svc, ok := value.(*service)
		if ok {
			svc.once.Do(func() {
				xlog.Infof("init service `%s`...", svc.is.Name())
				err = svc.is.InitAndStart()
				if err != nil {
					xlog.Warnf("init service `%s` failed: %v", svc.is.Name(), err)
					return
				}
				svc.inited = true
				xlog.Infof("init service %s done.", svc.is.Name())
			})
		} else {
			xlog.Warnf("Invalid service type %T", value)
			err = errors.New("invalid service type")
		}
		return err == nil
	})
	return err
}

func Stop() {
	reigstered.Range(func(key, value any) bool {
		svc, ok := value.(*service)
		if ok {
			xlog.Infof("stop service `%s`...", svc.is.Name())
			svc.is.Stop()
		} else {
			xlog.Warnf("Invalid service type %T, %v", value, key)
		}
		return true
	})
}

func Register(s ServiceI) {
	reigstered.LoadOrStore(s.Name(), NewService(s))
}

func loadService(key string) *service {
	if v, ok := reigstered.Load(key); ok {
		return v.(*service)
	}
	return nil
}

func Service[T any](key string) *T {
	svc := loadService(key)
	return any(svc.is).(*T)
}
