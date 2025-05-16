package xgin

import (
	"bytes"
	"time"

	xlog "com.imilair/chatbot/bootstrap/log"
	"github.com/gin-gonic/gin"
)

type ResponseWriterWrapper struct {
	gin.ResponseWriter
	Body *bytes.Buffer // 缓存
}

func (w ResponseWriterWrapper) Write(b []byte) (int, error) {
	w.Body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w ResponseWriterWrapper) WriteString(s string) (int, error) {
	w.Body.WriteString(s)
	return w.ResponseWriter.WriteString(s)
}

func LogHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		now := time.Now()
		xlog.Info("request: %v", ctx.Request.URL.Path)
		blw := &ResponseWriterWrapper{Body: bytes.NewBufferString(""), ResponseWriter: ctx.Writer}
		ctx.Writer = blw
		ctx.Next()
		xlog.Info("response: %v: %v, cost: %v", ctx.Request.URL.Path, ctx.Writer.Status(), time.Since(now))
	}
}
