package middlewares

import (
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

var (
	domains = []string{"imgo.tv", "mgtv.com"}
)

type CORSConfig struct {
	AllowAllOrigins bool          `yaml:"allowAllOrigins"`
	AllowOrigins    []string      `yaml:"allowOrigins"`
	AllHeaders      []string      `yaml:"allHeaders"`
	MaxAge          time.Duration `yaml:"maxAge"`
}

func CORSWithConfig(cfg *CORSConfig) gin.HandlerFunc {
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
			for _, domain := range domains {
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
