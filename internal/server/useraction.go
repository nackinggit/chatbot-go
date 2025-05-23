package server

import (
	"time"

	"com.imilair/chatbot/internal/model"
	"com.imilair/chatbot/internal/service/agents"
	"github.com/gin-gonic/gin"
)

func userActionCallback(ctx *gin.Context) {
	var req model.UserAction
	if err := ctx.BindJSON(&req); err != nil {
		JSONE(ctx, err, &req)
		return
	}
	if req.ActionType == model.ROOM {
		chatRoom, err := model.GetUserActionContent[model.Room](&req)
		if err != nil {
			JSONE(ctx, err, &req)
		} else {
			chatRoom.CreateTime = time.Now().UnixMilli()
			resp, err := agents.ChatRoomService.RoomActionCallback(ctx, chatRoom)
			JSONR(ctx, resp, err)
		}
		return
	} else {
		resp, err := agents.AssistantService.UserActionCallback(ctx, &req)
		JSONR(ctx, resp, err)
	}
}
