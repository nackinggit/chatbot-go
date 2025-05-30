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
		return
	}
	req.CreateTime = time.Now().UnixMilli()
	resp, err := agents.ChatRoomService.InputRecommend(ctx, &req)
	JSONR(ctx, resp, err)
}
