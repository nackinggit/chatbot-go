package xlog

import "context"

func ctxFields(ctx context.Context) []any {
	anys := []any{}
	for _, k := range logger.ctxFields {
		anys = append(anys, k, ctx.Value(k))
	}
	return anys
}

func Debugf(format string, args ...any) {
	l := logger.Sugar()
	l.Debugf(format, args...)
}

func Infof(format string, args ...any) {
	l := logger.Sugar()
	l.Infof(format, args...)
}

func Warnf(format string, args ...any) {
	l := logger.Sugar()
	l.Warnf(format, args...)
}

func Errorf(format string, args ...any) {
	l := logger.Sugar()
	l.Errorf(format, args...)
}

func DPanicf(format string, args ...any) {
	l := logger.Sugar()
	l.DPanicf(format, args...)
}

func Panicf(format string, args ...any) {
	l := logger.Sugar()
	l.Panicf(format, args...)
}

func Fatalf(format string, args ...any) {
	l := logger.Sugar()
	l.Fatalf(format, args...)
}

func DebugC(ctx context.Context, format string, args ...any) {
	l := logger.Sugar().With(ctxFields(ctx)...)
	l.Debugf(format, args...)
}

func InfoC(ctx context.Context, format string, args ...any) {
	l := logger.Sugar().With(ctxFields(ctx)...)
	l.Infof(format, args...)
}

func WarnC(ctx context.Context, format string, args ...any) {
	l := logger.Sugar().With(ctxFields(ctx)...)
	l.Warnf(format, args...)
}

func ErrorC(ctx context.Context, format string, args ...any) {
	l := logger.Sugar().With(ctxFields(ctx)...)
	l.Errorf(format, args...)
}

func DPanicC(ctx context.Context, format string, args ...any) {
	l := logger.Sugar().With(ctxFields(ctx)...)
	l.DPanicf(format, args...)
}

func PanicC(ctx context.Context, format string, args ...any) {
	l := logger.Sugar().With(ctxFields(ctx)...)
	l.Panicf(format, args...)
}

func FatalC(ctx context.Context, format string, args ...any) {
	l := logger.Sugar().With(ctxFields(ctx)...)
	l.Fatalf(format, args...)
}
