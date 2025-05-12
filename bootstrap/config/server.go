package config

import "github.com/gin-gonic/gin"

type Server interface {
	Start() error
	Stop() error
	HealthCheck() error
	Engine() *gin.Engine
}
