package logx

import (
	"context"
	"os"
	"strings"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

const (
	UriId          = "x-3rd-uri"
	TraceId        = "x-3rd-trace"
	MethodId       = "x-3rd-method"
	logTmFmtWithMS = "2006-01-02 15:04:05.000"
)

type ILogger interface {
	Debugw(ctx context.Context, msg string, keysAndValues ...interface{})
	Infow(ctx context.Context, msg string, keysAndValues ...interface{})
	Warnw(ctx context.Context, msg string, keysAndValues ...interface{})
	Errorw(ctx context.Context, msg string, keysAndValues ...interface{})
}

type logger struct {
	z *zap.Logger
}

func NewLogger(level string) (ILogger, error) {
	var l zapcore.Level
	switch strings.ToLower(level) {
	case "debug":
		l = zapcore.DebugLevel
	case "info":
		l = zapcore.InfoLevel
	case "warn":
		l = zapcore.WarnLevel
	case "error":
		l = zapcore.ErrorLevel
	default:
		l = zapcore.InfoLevel
	}
	var core = zapcore.NewCore(
		zapcore.NewJSONEncoder(zapcore.EncoderConfig{
			CallerKey:     "caller", // 打印文件名和行数
			LevelKey:      "level",
			MessageKey:    "msg",
			TimeKey:       "ts",
			StacktraceKey: "stacktrace",
			LineEnding:    zapcore.DefaultLineEnding,
			EncodeTime: func(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
				enc.AppendString(t.Format(logTmFmtWithMS))
			},
			EncodeLevel:    zapcore.LowercaseLevelEncoder, // 小写编码器
			EncodeCaller:   zapcore.ShortCallerEncoder,    // 全路径编码器
			EncodeDuration: zapcore.SecondsDurationEncoder,
			EncodeName:     zapcore.FullNameEncoder,
		}),
		zapcore.NewMultiWriteSyncer(zapcore.AddSync(os.Stdout)),
		l,
	)
	var z = zap.New(core,
		zap.AddCaller(),
		zap.AddStacktrace(zap.ErrorLevel), // error级别日志，打印堆栈
		zap.Development(),
	)
	return &logger{z: z}, nil
}

func (l *logger) Infow(ctx context.Context, msg string, keysAndValues ...interface{}) {
	l.withTraceId(ctx).Sugar().Infow(msg, keysAndValues...)
}

func (l *logger) Debugw(ctx context.Context, msg string, keysAndValues ...interface{}) {
	l.withTraceId(ctx).Sugar().Debugw(msg, keysAndValues...)
}

func (l *logger) Warnw(ctx context.Context, msg string, keysAndValues ...interface{}) {
	l.withTraceId(ctx).Sugar().Warnw(msg, keysAndValues...)
}

func (l *logger) Errorw(ctx context.Context, msg string, keysAndValues ...interface{}) {
	l.withTraceId(ctx).Sugar().Errorw(msg, keysAndValues...)
}

func (l *logger) withTraceId(ctx context.Context) *zap.Logger {
	// 1. 定义默认字段
	fields := []zap.Field{}

	// 2. trace
	var val, ok = ctx.Value(TraceId).(string)
	if ok && len(val) > 0 {
		fields = append(fields, zap.String(TraceId, val))
	}
	// 3. uri
	val, ok = ctx.Value(UriId).(string)
	if ok && len(val) > 0 {
		fields = append(fields, zap.String(UriId, val))
	}
	// 4. method
	val, ok = ctx.Value(MethodId).(string)
	if ok && len(val) > 0 {
		fields = append(fields, zap.String(MethodId, val))
	}
	return l.z.With(fields...)
}
