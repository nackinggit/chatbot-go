package middlewares

import (
	"bytes"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"com.imilair/chatbot/bootstrap/config"
	xlog "com.imilair/chatbot/bootstrap/log"
	"golang.org/x/time/rate"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func CORSWithConfig(cfg *config.CORSConfig) gin.HandlerFunc {
	corsCfg := cors.Config{
		AllowAllOrigins:  cfg.AllowAllOrigins,
		AllowCredentials: true,
		MaxAge:           time.Duration(cfg.MaxAge),
		AllowOriginFunc: func(origin string) bool {
			u, _ := url.Parse(origin)
			host, _, err := net.SplitHostPort(u.Host)
			if err != nil {
				host = u.Host
			}
			for _, domain := range cfg.Domains {
				if strings.HasSuffix(host, domain) {
					return true
				}
			}
			return false
		},
	}
	if !cfg.AllowAllOrigins {
		corsCfg.AllowOrigins = cfg.AllowOrigins
	}
	return cors.New(corsCfg)
}

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
		xlog.Infof("request: %v", ctx.Request.URL.Path)
		blw := &ResponseWriterWrapper{Body: bytes.NewBufferString(""), ResponseWriter: ctx.Writer}
		ctx.Writer = blw
		ctx.Next()
		xlog.Infof("response: %v: %v, cost: %v", ctx.Request.URL.Path, ctx.Writer.Status(), time.Since(now))
	}
}

var limiter = rate.NewLimiter(2, 5)

func RateLimitHandler() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		if limiter.AllowN(time.Now().Add(2*time.Second), 1) {
			ctx.Next()
		} else {
			ctx.AbortWithStatus(http.StatusTooManyRequests)
		}
	}
}
