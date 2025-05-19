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
	agents.TeacherService.QuestionAnalyse(ctx, &req)
}

func qaAll(ctx *gin.Context) {
	var req model.QARequest
	if err := ctx.ShouldBind(&req); err != nil {
		JSONE(ctx, nil, err, req)
		return
	}
	agents.TeacherService.AnswerQuestion(ctx, &req)
}
