package server

import (
	"io"
	"reflect"

	"com.imilair/chatbot/bootstrap/gin/middlewares"
	xlog "com.imilair/chatbot/bootstrap/log"
	"com.imilair/chatbot/internal/bcode"
	"com.imilair/chatbot/pkg/util"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"github.com/openai/openai-go/packages/ssestream"
)

func Route(e *gin.Engine) {
	apiV1 := e.Group("/api")
	{
		apiV1.POST("/user_action/callback", userActionCallback)
	}

	botv1 := apiV1.Group("/bot")
	{
		botv1.POST("/question_pic_analyse", questionAnalyse)
		botv1.POST("/qa_all", qaAll)
		botv1.POST("/qa_judge", judgeAnswer)
		botv1.POST("/manghe_pic_analyse", manghePicAnalyse)
		botv1.POST("/manghe_predict", manghePredict)
		botv1.POST("/extract_name", extractName)
		botv1.POST("/comment_pic", commentPic)
		botv1.POST("/comment_post", commentPost)
		botv1.POST("/fanyi", middlewares.RateLimitHandler(), comicTranslate)
	}
	chatroomv1 := apiV1.Group("/chat_room")
	{
		chatroomv1.POST("/recommend", inputRecommend)
	}
}

func JSONR(ctx *gin.Context, data any, err error) {
	if err != nil {
		JSONE[any](ctx, err, nil)
	}
	ctx.JSON(200, gin.H{"code": 0, "message": "", "data": data})
}

func JSONE[T any](ctx *gin.Context, err error, req *T) {
	if berr, ok := err.(bcode.BError); ok {
		ctx.JSON(200, gin.H{"code": berr.Code(), "message": berr.Message()})
	} else if ve, ok := err.(validator.ValidationErrors); ok && req != nil {
		ctx.JSON(400, gin.H{"code": 400, "message": validatorErr(ve, req)})
	} else {
		ctx.JSON(200, gin.H{"code": 500, "message": err.Error()})
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
