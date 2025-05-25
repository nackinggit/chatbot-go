package main

import (
	"com.imilair/chatbot/bootstrap"
	xlog "com.imilair/chatbot/bootstrap/log"
	"com.imilair/chatbot/internal/server"
	"com.imilair/chatbot/internal/service"
	"com.imilair/chatbot/internal/service/config"
)

type app struct {
}

func newServer() bootstrap.Server[config.ServiceConfig] {
	return &app{}
}

// 创建或启动资源
func (a *app) Start() error {
	err := service.Init()
	if err != nil {
		xlog.Warnf("service init failed: %v", err)
		return err
	}
	xlog.Info("service init success")
	return nil
}

// 回收资源
func (a *app) Stop() error {
	service.Stop()
	xlog.Info("service stop success")
	return nil
}

// 返回配置结构,如果返回nil,则需要自己初始化app配置
func (a *app) Config() *config.ServiceConfig {
	return &service.Config
}

func main() {
	// 启动
	bootstrap.Run(newServer(), server.Route)
}
