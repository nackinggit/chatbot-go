package server

import (
	"time"

	"com.imilair/chatbot/internal/model"
	"com.imilair/chatbot/internal/service/agents"
	"github.com/gin-gonic/gin"
)

func inputRecommend(ctx *gin.Context) {
	var req model.InputRecommendRequest
	if err := ctx.ShouldBind(&req); err != nil {
		JSONE(ctx, err, &req)
	}
	req.CreateTime = time.Now().UnixMilli()
	agents.ChatRoomService.InputRecommend(ctx, &req)
}
