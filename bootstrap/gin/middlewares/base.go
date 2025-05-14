package middlewares

import (
	"net"
	"net/url"
	"strings"
	"time"

	"com.imilair/chatbot/bootstrap/config"
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
