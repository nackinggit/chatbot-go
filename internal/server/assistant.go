package server

import (
	"com.imilair/chatbot/internal/model"
	"com.imilair/chatbot/internal/service/agents"
	"github.com/gin-gonic/gin"
)

func extractName(ctx *gin.Context) {
	var req model.ExtractNameRequest
	if err := ctx.BindJSON(&req); err != nil {
		JSONE(ctx, err, &req)
		return
	}
	resp, err := agents.AssistantService.ExtractName(ctx, &req)
	JSONR(ctx, resp, err)
}

func commentPic(ctx *gin.Context) {
	var req model.CommentPicRequest
	if err := ctx.BindJSON(&req); err != nil {
		JSONE(ctx, err, &req)
		return
	}
	resp, err := agents.AssistantService.CommentPic(ctx, &req)
	JSONR(ctx, resp, err)
}

func commentPost(ctx *gin.Context) {
}
