package server

import (
	"io"
	"reflect"

	xlog "com.imilair/chatbot/bootstrap/log"
	"com.imilair/chatbot/internal/bcode"
	"com.imilair/chatbot/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/openai/openai-go/packages/ssestream"
)

func Route(e *gin.Engine) {
	apiV1 := e.Group("/api")
	botv1 := apiV1.Group("/bot")
	{
		botv1.POST("/question_pic_analyse", questionAnalyse)
		botv1.POST("/qa_all", qaAll)
	}

}

func JSONE[T any](ctx *gin.Context, obj any, err error, req T) {
	if berr, ok := err.(bcode.BError); ok {
		ctx.JSON(200, gin.H{"data": obj, "code": berr.Code(), "message": berr.Message()})
	} else if ve, ok := err.(validator.ValidationErrors); ok {
		ctx.JSON(400, gin.H{"data": obj, "code": 400, "message": validatorErr(ve, req)})
	} else {
		ctx.JSON(200, gin.H{"data": obj, "code": 500, "message": err.Error()})
	}
}

func validatorErr[T any](ve validator.ValidationErrors, req T) string {
	s := reflect.TypeOf(req)
	for _, fe := range ve {
		field, _ := s.FieldByName(fe.Field())
		etag := fe.Tag() + "_err"
		eTagtext := field.Tag.Get(etag)
		errText := field.Tag.Get("err")
		if eTagtext != "" {
			return eTagtext
		} else if errText != "" {
			return errText
		}
	}
	return ve.Error()
}

func SSEResponse[T any](ctx *gin.Context, stream *ssestream.Stream[T]) {
	ctx.Writer.Header().Set("Content-Type", "text/event-stream")
	ctx.Writer.Header().Set("Cache-Control", "no-cache")
	ctx.Writer.Header().Set("Connection", "keep-alive")
	ctx.Writer.Header().Set("Transfer-Encoding", "chunked")

	ctx.Stream(func(w io.Writer) bool {
		for stream.Next() {
			xlog.Infof("data: %v", stream.Current())
			ctx.SSEvent("data", util.JsonString(stream.Current()))
			return true
		}
		if stream.Err() != nil {
			ctx.SSEvent("data", stream.Err().Error())
			return false
		}
		return false
	})
}
