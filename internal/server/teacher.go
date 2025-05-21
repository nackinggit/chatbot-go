package server

import (
	"com.imilair/chatbot/internal/model"
	"com.imilair/chatbot/internal/service/agents"
	"github.com/gin-gonic/gin"
)

func questionAnalyse(ctx *gin.Context) {
	var req model.ImageAnalyseRequest
	if err := ctx.ShouldBind(&req); err != nil {
		JSONE(ctx, err, &req)
		return
	}
	agents.TeacherService.QuestionAnalyse(ctx, &req)
}

func qaAll(ctx *gin.Context) {
	var req model.QARequest
	if err := ctx.ShouldBind(&req); err != nil {
		JSONE(ctx, err, &req)
		return
	}
	agents.TeacherService.AnswerQuestion(ctx, &req)
}

func judgeAnswer(ctx *gin.Context) {
	var req model.JudgeAnswerRequest
	if err := ctx.ShouldBind(&req); err != nil {
		JSONE(ctx, err, &req)
	}
	agents.TeacherService.JudgeAnswer(ctx, &req)
}
