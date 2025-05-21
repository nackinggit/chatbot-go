package server

import (
	"com.imilair/chatbot/internal/model"
	"com.imilair/chatbot/internal/service/agents"
	"github.com/gin-gonic/gin"
)

func manghePicAnalyse(ctx *gin.Context) {
	var req model.ImageRequest
	if err := ctx.ShouldBind(&req); err != nil {
		JSONE(ctx, err, &req)
		return
	}
	agents.MangHeService.ImageAnalyse(ctx, req.ImageUrl)
}

func manghePredict(ctx *gin.Context) {
	var req model.MangHePredictRequest
	if err := ctx.ShouldBind(&req); err != nil {
		JSONE(ctx, err, &req)
		return
	}
	agents.MangHeService.Predict(ctx, &req)
}
