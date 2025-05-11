package log

import (
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/natefinch/lumberjack"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Options struct {
	LogFileDir    string //文件保存地方
	AppName       string //日志文件前缀
	ErrorFileName string
	WarnFileName  string
	InfoFileName  string
	DebugFileName string
	Level         zapcore.Level //日志等级
	MaxSize       int           //日志文件小大（M）
	MaxBackups    int           // 最多存在多少个切片文件
	MaxAge        int           //保存的最大天数
	Development   bool          //是否是开发模式
	zap.Config
}

var (
	l                              *Logger
	sp                             = string(filepath.Separator)
	errWS, warnWS, infoWS, debugWS zapcore.WriteSyncer       // IO输出
	debugConsoleWS                 = zapcore.Lock(os.Stdout) // 控制台标准输出
	errorConsoleWS                 = zapcore.Lock(os.Stderr)
)

type Logger struct {
	*zap.Logger
	sync.RWMutex
	Opts        *Options `json:"opts"`
	zapConfig   zap.Config
	initialized bool
}

func NewLogger(opts *Options) *zap.Logger {
	l = &Logger{Opts: opts}
	l.Lock()
	defer l.Unlock()
	if l.initialized {
		l.Info("[NewLogger] logger initEd")
		return nil
	}
	if l.Opts == nil {
		l.Opts = &Options{
			LogFileDir:    "",
			AppName:       "app",
			ErrorFileName: "error.log",
			WarnFileName:  "warn.log",
			InfoFileName:  "info.log",
			DebugFileName: "debug.log",
			Level:         zapcore.DebugLevel,
			MaxSize:       100,
			MaxBackups:    60,
			MaxAge:        30,
			Development:   true,
		}
	}

	if l.Opts.LogFileDir == "" {
		l.Opts.LogFileDir, _ = filepath.Abs(filepath.Dir(filepath.Join(".")))
		l.Opts.LogFileDir += sp + "logs" + sp
	}
	if l.Opts.Development {
		l.zapConfig = zap.NewDevelopmentConfig()
		l.zapConfig.EncoderConfig.EncodeTime = timeEncoder
	} else {
		l.zapConfig = zap.NewProductionConfig()
		l.zapConfig.EncoderConfig.EncodeTime = timeUnixNano
	}
	if len(l.Opts.OutputPaths) == 0 {
		l.zapConfig.OutputPaths = []string{"stdout"}
	}
	if len(l.Opts.ErrorOutputPaths) == 0 {
		l.zapConfig.OutputPaths = []string{"stderr"}
	}
	l.zapConfig.Level.SetLevel(l.Opts.Level)
	l.init()
	l.initialized = true
	return l.Logger
}

func (l *Logger) init() {
	l.setSyncs()
	var err error
	l.Logger, err = l.zapConfig.Build(l.cores())
	if err != nil {
		panic(err)
	}
	defer l.Logger.Sync()
}

func (l *Logger) setSyncs() {
	f := func(fname string) zapcore.WriteSyncer {
		return zapcore.AddSync(&lumberjack.Logger{
			Filename:   l.Opts.LogFileDir + sp + l.Opts.AppName + "-" + fname,
			MaxSize:    l.Opts.MaxSize,
			MaxBackups: l.Opts.MaxBackups,
			MaxAge:     l.Opts.MaxAge,
			Compress:   true,
			LocalTime:  true,
		})
	}
	errWS = f(l.Opts.ErrorFileName)
	warnWS = f(l.Opts.WarnFileName)
	infoWS = f(l.Opts.InfoFileName)
	debugWS = f(l.Opts.DebugFileName)
}

func (l *Logger) cores() zap.Option {
	fileEncoder := zapcore.NewJSONEncoder(l.zapConfig.EncoderConfig)
	encoderConfig := consoleConfig()
	consoleEncoder := zapcore.NewConsoleEncoder(encoderConfig)

	errPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl == zapcore.ErrorLevel && zapcore.ErrorLevel-l.zapConfig.Level.Level() > -1
	})
	warnPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl == zapcore.WarnLevel && zapcore.WarnLevel-l.zapConfig.Level.Level() > -1
	})
	infoPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl == zapcore.InfoLevel && zapcore.InfoLevel-l.zapConfig.Level.Level() > -1
	})
	debugPriority := zap.LevelEnablerFunc(func(lvl zapcore.Level) bool {
		return lvl == zapcore.DebugLevel && zapcore.DebugLevel-l.zapConfig.Level.Level() > -1
	})
	cores := []zapcore.Core{
		zapcore.NewCore(fileEncoder, errWS, errPriority),
		zapcore.NewCore(fileEncoder, warnWS, warnPriority),
		zapcore.NewCore(fileEncoder, infoWS, infoPriority),
		zapcore.NewCore(fileEncoder, debugWS, debugPriority),
	}
	if l.Opts.Development {
		cores = append(cores, []zapcore.Core{
			zapcore.NewCore(consoleEncoder, errorConsoleWS, errPriority),
			zapcore.NewCore(consoleEncoder, debugConsoleWS, warnPriority),
			zapcore.NewCore(consoleEncoder, debugConsoleWS, infoPriority),
			zapcore.NewCore(consoleEncoder, debugConsoleWS, debugPriority),
		}...)
	}
	return zap.WrapCore(func(c zapcore.Core) zapcore.Core {
		return zapcore.NewTee(cores...)
	})
}
func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString(t.Format("2006-01-02 15:04:05"))
}

func timeUnixNano(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendInt64(t.UnixNano() / 1e6)
}

var logger *zap.Logger

// log instance init
func InitLog(opt *Options) {
	logger = NewLogger(opt)
}

func GetLogger() *zap.SugaredLogger {
	return logger.Sugar()
}

const (
	logTmFmtWithMS = "2006-01-02 15:04:05.000"
)

func consoleConfig() zapcore.EncoderConfig {
	// 自定义时间输出格式
	customTimeEncoder := func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString("[" + t.Format(logTmFmtWithMS) + "]")
	}

	// 自定义文件：行号输出项
	customCallerEncoder := func(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
		enc.AppendString("[" + caller.TrimmedPath() + "]")
	}

	return zapcore.EncoderConfig{
		CallerKey:      "caller_line", // 打印文件名和行数
		LevelKey:       "level_name",
		MessageKey:     "msg",
		TimeKey:        "ts",
		StacktraceKey:  "stacktrace",
		LineEnding:     zapcore.DefaultLineEnding,
		EncodeTime:     customTimeEncoder,                // 自定义时间格式
		EncodeLevel:    zapcore.CapitalColorLevelEncoder, // 小写编码器
		EncodeCaller:   customCallerEncoder,              // 全路径编码器
		EncodeDuration: zapcore.SecondsDurationEncoder,
		EncodeName:     zapcore.FullNameEncoder,
	}
}
