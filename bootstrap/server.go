package bootstrap

import (
	"context"
	"errors"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"sync"
	"sync/atomic"
	"syscall"
	"time"

	"com.imilair/chatbot/bootstrap/config"
	xgin "com.imilair/chatbot/bootstrap/gin"
	xlog "com.imilair/chatbot/bootstrap/log"
	"com.imilair/chatbot/pkg/embedding"
	"com.imilair/chatbot/pkg/llm"
	"com.imilair/chatbot/pkg/util"
	"com.imilair/chatbot/pkg/xmilvus"
	"com.imilair/chatbot/pkg/xmysql"
	"com.imilair/chatbot/pkg/xredis"

	ginmiddlewars "com.imilair/chatbot/bootstrap/gin/middlewares"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/spf13/viper"
)

var (
	gapp   app
	Config config.Config
)

type app struct {
	middlewares []gin.HandlerFunc
	httpRouter  func(*gin.Engine)
	httpServer  *xgin.HttpGinServer
}

var (
	ConfigPath string
)

func init() {
	flag.StringVar(&ConfigPath, "configpath", "./configs", "application config path")
}

func (a *app) Init(router func(e *gin.Engine), middlewares ...gin.HandlerFunc) error {
	if !flag.Parsed() {
		flag.Parse()
	}
	version := Version{}
	version.Init()

	err := initConfig()
	if err != nil {
		return err
	}
	if Config.Logger != nil {
		xlog.InitLog(Config.Logger)
	}
	if len(Config.Embedding) > 0 {
		err = embedding.Init(Config.Embedding)
		if err != nil {
			return err
		}
	}
	if len(Config.MySql) > 0 {
		xmysql.Init(Config.MySql)
	}
	if len(Config.LLMS) > 0 {
		err = llm.Init(Config.LLMS)
		if err != nil {
			return err
		}
	}
	if Config.Redis != nil {
		xredis.Init(Config.Redis)
	}
	if Config.Milvus != nil {
		err = xmilvus.Init(Config.Milvus)
		if err != nil {
			return err
		}
	}
	a.httpRouter = router
	a.middlewares = append(a.middlewares, ginmiddlewars.LogHandler())
	a.middlewares = append(a.middlewares, middlewares...)
	return nil
}

func initConfig() error {
	if _, err := os.Stat(ConfigPath); err == nil {
		viper.AddConfigPath(ConfigPath)
	} else {
		xlog.Warnf("application config path not setting, use default : %s", "./configs")
		viper.AddConfigPath("./configs")
	}
	viper.SetConfigName("application")
	err := viper.ReadInConfig()
	if err != nil {
		return err
	}

	env := viper.GetString("env")
	if env == "" {
		env = "default"
		xlog.Infof("application env not setting, use default : %s", env)
		viper.Set("env", env)
	} else {
		viper.SetConfigName(fmt.Sprintf("application-%s", env))
		err := viper.MergeInConfig()
		if err != nil {
			return err
		}
	}

	if err := viper.Unmarshal(&Config); err != nil {
		panic(fmt.Errorf("application init err : %v", err))
	}
	if Config.Logger != nil && Config.Logger.LogFilename == "" {
		Config.Logger.LogFilename = Config.App.Name
	}
	xlog.Infof("application config : %v", util.JsonString(Config))
	return nil
}

func DefaultHealthCheckHandler(ctx *gin.Context) {
	ctx.AbortWithStatus(http.StatusOK)
}

func DefaultPrometheusHandler(ctx *gin.Context) {
	h := promhttp.Handler()
	h.ServeHTTP(ctx.Writer, ctx.Request)
}

func (a *app) start() error {
	errCh := make(chan error, 1)

	start := false
	attachedMonitor := false

	attachMonitor := func(e *gin.Engine) {
		attachedMonitor = true
		monitor := e.Group("/monitor")
		monitor.GET("/health", DefaultHealthCheckHandler)
		monitor.GET("/prometheus", DefaultPrometheusHandler)
	}

	if Config.HttpServer != nil {
		start = true
		httpServerCfg := Config.HttpServer
		svr, err := xgin.NewServer(httpServerCfg, a.httpRouter, a.middlewares...)
		a.httpServer = svr
		if err != nil {
			errCh <- err
		} else {
			if !attachedMonitor {
				attachMonitor(svr.Engine())
			}

			go func() {
				defer recovery(func(i any) {
					errCh <- fmt.Errorf("panic : %v", i)
				})
				xlog.Infof("gin server start at port: %v", httpServerCfg.Http.Port)
				errCh <- svr.Start()
			}()
		}
	}

	if !start {
		errCh <- nil
	}

	return <-errCh
}

type Handler func(any)

func recovery(hr ...Handler) {
	if r := recover(); r != nil {
		buf := make([]byte, 1<<18)
		n := runtime.Stack(buf, false)
		xlog.Errorf("%v, Stack: %s", r, buf[0:n])
		for _, h := range hr {
			h(r)
		}
	}
}

// close input traffic
func (a *app) preStop() error {
	errs := []error{}
	if a.httpServer != nil {
		err := a.httpServer.Stop()
		if err != nil {
			errs = append(errs, err)
		}
	}
	return errors.Join(errs...)
}

// garbage collection
func (a *app) postStop() error {
	return nil
}

type Server interface {
	Start() error
	Stop() error
	Config() any
}

func Run(server Server, router func(e *gin.Engine), middlewares ...gin.HandlerFunc) {
	if err := gapp.Init(router, middlewares...); err != nil {
		xlog.Fatalf("application init error : %v", err)
		panic(err)
	}

	serviceCfg := server.Config()
	if serviceCfg != nil {
		if err := viper.Unmarshal(serviceCfg); err != nil {
			panic(fmt.Errorf("service init err : %v", err))
		}
	}

	r := &appRunner{
		signals: make(chan os.Signal, 1),
		server:  server,
		errCh:   make(chan error, 2),
	}
	r.run()
	r.Wait()
}

type appRunner struct {
	stop    int32
	signals chan os.Signal
	server  Server
	wg      sync.WaitGroup
	errCh   chan error
}

func (r *appRunner) run() {
	go r.handleSignal()
	go r.handleErr()
	r.handleStart()
}

func (r *appRunner) handleStart() {
	r.wg.Add(1)

	err := r.server.Start()
	if err != nil {
		r.errCh <- err
	}

	util.AsyncGoWithDefault(context.Background(), func() {
		r.errCh <- gapp.start()
	})
}

func (r *appRunner) handlStop() {
	if !atomic.CompareAndSwapInt32(&r.stop, 0, 1) {
		return
	}

	defer r.wg.Done()

	go func() {
		to := 5 * time.Second
		timeout := Config.GetGracefulTimeout()
		if timeout > time.Second {
			to = timeout
		}

		time.Sleep(to)
		xlog.Error("stop : timeout")
		os.Exit(1)
	}()

	// stop input traffic
	err := gapp.preStop()
	if err != nil {
		xlog.Errorf("pre stop error : %v", err)
	} else {
		xlog.Infof("pre stop")
	}

	err = r.server.Stop()
	if err != nil {
		xlog.Errorf("stop app error : %v", err)
	} else {
		xlog.Info("stop app")
	}
	r.Wait()
	// clean up
	err = gapp.postStop()
	if err != nil {
		xlog.Errorf("post stop error : %v", err)
	} else {
		xlog.Info("post stop")
	}
}

func (r *appRunner) Wait() {
	r.wg.Wait()
	close(r.errCh)
	_ = xlog.Sync()
}

func (r *appRunner) handleSignal() {
	signal.Notify(r.signals, syscall.SIGPIPE, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT, syscall.SIGABRT)
	for {
		sig := <-r.signals
		xlog.Infof("received signal: %s", sig)
		switch sig {
		case syscall.SIGPIPE:
		default:
			r.handlStop()
			return
		}
	}
}

func (r *appRunner) handleErr() {
	for {
		err, ok := <-r.errCh
		if !ok {
			return
		}
		if err != nil {
			xlog.Errorf("received error : %v", err)
			r.handlStop()
			continue
		}
	}
}
