package server

import (
	"time"

	"com.imilair/chatbot/bootstrap/gin/middlewares"
	xlog "com.imilair/chatbot/bootstrap/log"
	"com.imilair/chatbot/internal/model"
	"com.imilair/chatbot/internal/service/agents"
	"com.imilair/chatbot/pkg/util"
	"com.imilair/chatbot/pkg/util/ttlmap"
	xvc "com.imilair/chatbot/pkg/volceengine"
	"github.com/gin-gonic/gin"
)

var ur = ttlmap.New(10000, 30)

func userActionCallback(ctx *gin.Context) {
	var req model.UserAction
	if err := ctx.ShouldBind(&req); err != nil {
		JSONE(ctx, err, &req)
		return
	}
	mid := middlewares.GetMid(ctx)
	if mid != "" && ur.Contains(mid) {
		xlog.Infof("message %s handled...", util.JsonString(req))
		JSONR(ctx, nil, nil)
		return
	} else {
		ur.Put(mid, true)
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

func entitySegment(ctx *gin.Context) {
	var req model.ImageRequest
	if err := ctx.ShouldBind(&req); err != nil {
		JSONE(ctx, err, &req)
		return
	}
	resp, err := xvc.EntitySegment(req.ImageUrl)
	JSONR(ctx, resp, err)
}
