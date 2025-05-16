package server

import (
	"com.imilair/chatbot/internal/model"
	"com.imilair/chatbot/internal/service/agents"
	"github.com/gin-gonic/gin"
)

func qaAll(ctx *gin.Context) {
	var req model.QuestionAnalyseRequest
	if err := ctx.Bind(&req); err != nil {
		JSON(ctx, nil, err)
	}
	SSEResponse(ctx, agents.Teacher().QuestionAnalyse(ctx, &req))
}
