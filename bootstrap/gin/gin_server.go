package xgin

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"sync/atomic"
	"time"

	"com.imilair/chatbot/bootstrap/config"
	xlog "com.imilair/chatbot/bootstrap/log"
	"github.com/gin-gonic/gin"
	"golang.org/x/net/netutil"
)

type HttpMode string

const (
	ModeDebug   HttpMode = "debug"
	ModeRelease HttpMode = "release"
)

const (
	serverStatusInit = iota
	serverStatusReady
	serverStatusClosed
)

type ServerOptions struct {
	maxListenLimit    int
	readTimeout       time.Duration
	readHeaderTimeout time.Duration
	writeTimeout      time.Duration
	idleTimeout       time.Duration
	maxHeaderBytes    int
	gracefulTimeout   time.Duration
	mode              HttpMode

	tlsConfig   *tls.Config
	middlewares []gin.HandlerFunc
	router      func(*gin.Engine)
}

func (sv *ServerOptions) withConfig(cfg *config.HttpServerConfig) *ServerOptions {
	return sv.withHttpConfig(cfg.Http).withGinConfig(cfg.GinServer)
}

func (sv *ServerOptions) withHttpConfig(cfg *config.HttpConfig) *ServerOptions {
	if cfg.MaxListenLimit > 0 {
		sv.maxListenLimit = cfg.MaxHeaderBytes
	}
	if cfg.ReadTimeout > 0 {
		sv.readTimeout = cfg.ReadTimeout
	}
	if cfg.ReadHeaderTimeout > 0 {
		sv.readHeaderTimeout = cfg.ReadHeaderTimeout
	}
	if cfg.WriteTimeout > 0 {
		sv.writeTimeout = cfg.WriteTimeout
	}
	if cfg.IdleTimeout > 0 {
		sv.idleTimeout = cfg.IdleTimeout
	}
	if cfg.MaxHeaderBytes > 0 {
		sv.maxHeaderBytes = cfg.MaxHeaderBytes
	}
	if cfg.GracefulTimeout > 0 {
		sv.gracefulTimeout = cfg.GracefulTimeout
	}
	if cfg.Tls != nil {
		cert, err := tls.LoadX509KeyPair(cfg.Tls.CertFile, cfg.Tls.KeyFile)
		if err != nil {
			panic(err)
		}
		sv.tlsConfig = &tls.Config{
			Certificates: []tls.Certificate{cert},
		}
	}
	return sv
}

func (sv *ServerOptions) withGinConfig(cfg *config.GinServerConfig) *ServerOptions {
	if cfg == nil {
		return sv
	}
	if cfg.Mode != "" {
		sv.mode = HttpMode(cfg.Mode)
	}
	return sv
}

// NewServer 创建一个http server
//
// 代码示例
//
//	mws := []gin.HandlerFunc{
//		ServerRecoveryMiddleware(),
//		ServerContextMiddleware(family),
//		ServerSkywalkingMiddleware(globalTracer),
//		ServerPrometheusMiddleware(family),
//	}
//	opts := []ServerOption{
//		WithServerMiddlwares(mws...),
//	}
//	svr := NewServer(":8080", func(e *gin.Engine){
//		e.GET("/user", func(c *gin.Context) {
//			c.JSON(200, resp)
//		})
//	}, opts...)
//
//	svr.Start()
func NewServer(cfg *config.HttpServerConfig, router func(*gin.Engine), middlewares ...gin.HandlerFunc) (*HttpGinServer, error) {
	address, err := cfg.GetListen()
	if address == "" || err != nil {
		return nil, errors.New("empty listen " + err.Error())
	}

	rwTo := 2 * time.Second
	options := &ServerOptions{
		maxListenLimit:    0,
		readTimeout:       -1,
		readHeaderTimeout: rwTo,
		writeTimeout:      -1,
		idleTimeout:       -1,
		gracefulTimeout:   time.Second * 5,
		mode:              ModeRelease,
		router:            router,
		middlewares:       middlewares,
	}
	options = options.withConfig(cfg)

	switch options.mode {
	case ModeDebug:
		gin.SetMode(gin.DebugMode)
	default:
		gin.SetMode(gin.ReleaseMode)
	}

	engine := gin.New()
	engine.Use(options.middlewares...)

	if options.router != nil {
		options.router(engine)
	}

	listener, err := net.Listen("tcp", address)
	if err != nil {
		xlog.Errorf("Http listen(%s) failed : %+v", address, err)
		return nil, err
	}

	if options.maxListenLimit > 0 {
		listener = netutil.LimitListener(listener, options.maxListenLimit)
	}

	return &HttpGinServer{
		engine: engine,
		srv: &http.Server{
			Handler:           engine,
			TLSConfig:         options.tlsConfig,
			ReadTimeout:       options.readTimeout,
			ReadHeaderTimeout: options.readHeaderTimeout,
			WriteTimeout:      options.writeTimeout,
			IdleTimeout:       options.idleTimeout,
			MaxHeaderBytes:    options.maxHeaderBytes,
		},
		listener:        listener,
		gracefulTimeout: options.gracefulTimeout,
	}, nil
}

type HttpGinServer struct {
	engine          *gin.Engine
	srv             *http.Server
	listener        net.Listener
	gracefulTimeout time.Duration
	status          int32
}

func (s *HttpGinServer) Start() error {
	atomic.StoreInt32(&s.status, serverStatusReady)
	defer atomic.StoreInt32(&(s.status), serverStatusClosed)
	return s.srv.Serve(s.listener)
}

func (s *HttpGinServer) Stop() error {
	defer atomic.StoreInt32(&s.status, serverStatusClosed)
	ctx, cancel := context.WithTimeout(context.Background(), s.gracefulTimeout)
	defer cancel()
	if err := s.srv.Shutdown(ctx); err != nil {
		xlog.Errorf("stop http server error : %+v", err)
		return err
	}
	return nil
}

func (s *HttpGinServer) Engine() *gin.Engine {
	return s.engine
}

func (s *HttpGinServer) HealthCheck() error {
	switch atomic.LoadInt32(&s.status) {
	case serverStatusInit:
		return errors.New("http server is not ready")
	case serverStatusClosed:
		return errors.New("http server is closed")
	}
	return nil
}
