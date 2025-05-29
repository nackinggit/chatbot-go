package middlewares

import (
	"bytes"
	"crypto/md5"
	"encoding/hex"
	"io"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"com.imilair/chatbot/bootstrap/config"
	xlog "com.imilair/chatbot/bootstrap/log"
	"com.imilair/chatbot/pkg/util"
	"golang.org/x/time/rate"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

const (
	bodyKey = "cbody"
)

func cacheBody(ctx *gin.Context) []byte {
	vb, ok := ctx.Get(bodyKey)
	if !ok {
		if ctx.Request.Method == "POST" {
			buf := &bytes.Buffer{}
			tea := io.TeeReader(ctx.Request.Body, buf)
			body, err := io.ReadAll(tea)
			if err != nil {
				xlog.Panicf("read body err: %+v", err)
			}
			ctx.Request.Body = io.NopCloser(buf)

			ctx.Set(bodyKey, body)
			return body
		} else if ctx.Request.Method == "GET" {
			resp := make(map[string]any)
			query := ctx.Request.URL.Query()
			for k, v := range query {
				if len(v) == 1 {
					resp[k] = v[0]
				} else {
					resp[k] = v
				}
			}
			bs, _ := util.Marshal(resp)
			ctx.Set(bodyKey, bs)
			return bs
		}
	}
	return vb.([]byte)
}

func setmid(ctx *gin.Context) {
	bs := cacheBody(ctx)
	hash := md5.Sum(bs)
	ctx.Set("x-mid", hex.EncodeToString(hash[:]))
	rid := ctx.GetHeader("X-Request-Id")
	if rid == "" {
		rid = util.NewSnowflakeID().String()
	}
	ctx.Set("x-request-id", rid)
}

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
		setmid(ctx)
		xlog.InfoC(ctx, "request: %v", ctx.Request.URL.Path)
		blw := &ResponseWriterWrapper{Body: bytes.NewBufferString(""), ResponseWriter: ctx.Writer}
		ctx.Writer = blw
		ctx.Next()
		xlog.InfoC(ctx, "response: %v: %v, cost: %v", ctx.Request.URL.Path, ctx.Writer.Status(), time.Since(now))
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

func GetMid(ctx *gin.Context) string {
	if s, ok := ctx.Value("x-mid").(string); ok {
		return s
	}
	return ""
}
