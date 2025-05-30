package server

import (
	"errors"

	"com.imilair/chatbot/internal/model"
	"com.imilair/chatbot/internal/service/agents"
	"github.com/gin-gonic/gin"
)

func extractName(ctx *gin.Context) {
	var req model.ExtractNameRequest
	if err := ctx.ShouldBind(&req); err != nil {
		JSONE(ctx, err, &req)
		return
	}
	resp, err := agents.AssistantService.ExtractName(ctx, &req)
	JSONR(ctx, resp, err)
}

func commentPic(ctx *gin.Context) {
	var req model.CommentPicRequest
	if err := ctx.ShouldBind(&req); err != nil {
		JSONE(ctx, err, &req)
		return
	}
	resp, err := agents.AssistantService.CommentPic(ctx, &req)
	JSONR(ctx, resp, err)
}

func commentPost(ctx *gin.Context) {
	JSONE[any](ctx, errors.New("not implemented"), nil)
}

func comicTranslate(ctx *gin.Context) {
	var req model.ImageRequest
	if err := ctx.ShouldBind(&req); err != nil {
		JSONE(ctx, err, &req)
		return
	}
	resp, err := agents.AssistantService.ComicTranslate(ctx, &req)
	JSONR(ctx, resp, err)
}

func outsideList(ctx *gin.Context) {
	var req model.OutsideListRequest
	if err := ctx.ShouldBind(&req); err != nil {
		JSONE(ctx, err, &req)
		return
	}
	resp, err := agents.AssistantService.OutsideList(ctx, &req)
	JSONR(ctx, resp, err)
}
