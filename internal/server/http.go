package server

import (
	"io"

	"com.imilair/chatbot/bootstrap/gin/middlewares"
	"com.imilair/chatbot/internal/bcode"
	"com.imilair/chatbot/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/openai/openai-go/packages/ssestream"
)

func Route(e *gin.Engine) {
	apiV1 := e.Group("/api")
	apiV1.POST("/bot/question_pic_analyse", middlewares.StreamHeadersMiddleware(), qaAll)
}

func JSON(ctx *gin.Context, obj any, err error) {
	if berr, ok := err.(bcode.BError); ok {
		ctx.JSON(200, gin.H{"data": obj, "code": berr.Code(), "message": berr.Message()})
	} else {
		ctx.JSON(200, gin.H{"data": obj, "code": 500, "message": err.Error()})
	}
}

func SSEResponse[T any](ctx *gin.Context, stream *ssestream.Stream[T]) {
	for stream.Next() {
		ctx.Stream(func(w io.Writer) bool {
			ctx.SSEvent("data", util.JsonString(stream.Current()))
			return true
		})
	}
	if stream.Err() != nil {
		ctx.Stream(func(w io.Writer) bool {
			ctx.SSEvent("data", util.JsonString(stream.Current()))
			return false
		})
	}
}
