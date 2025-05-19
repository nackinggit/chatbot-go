package server

import (
	"com.imilair/chatbot/internal/model"
	"com.imilair/chatbot/internal/service/agents"
	"github.com/gin-gonic/gin"
)

func questionAnalyse(ctx *gin.Context) {
	var req model.QuestionAnalyseRequest
	if err := ctx.ShouldBind(&req); err != nil {
		JSONE(ctx, nil, err, req)
		return
	}
	agents.Teacher().QuestionAnalyse(ctx, &req)
}
