package logger

import (
	"github.com/go-logr/logr"
	"github.com/go-logr/zapr"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

// Logger is a custom logger struct that embeds logr.Logger.
type Logger struct {
	logr.Logger
}

// LogLevel type
type LogLevel int8

// Logging levels
const (
	DebugLevel LogLevel = iota - 1
	InfoLevel
	WarnLevel
	ErrorLevel
	DPanicLevel
	PanicLevel
	FatalLevel
)

// Options is a struct for configuring the Logger.
type Options struct {
	Development bool
	Level       LogLevel
	Debug       bool
}

// New returns a new instance of Logger with the given options.
func New(opts Options) *Logger {
	var zapLogger *zap.Logger
	var err error

	config := zap.Config{
		Level:       zap.NewAtomicLevelAt(zapcore.Level(opts.Level)),
		Development: opts.Development,
		Encoding:    "json",
		EncoderConfig: zapcore.EncoderConfig{
			TimeKey:        "timestamp",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "message",
			StacktraceKey:  "stacktrace",
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeLevel:    zapcore.LowercaseLevelEncoder,
			EncodeTime:     zapcore.ISO8601TimeEncoder,
			EncodeDuration: zapcore.StringDurationEncoder,
			EncodeCaller:   zapcore.ShortCallerEncoder,
		},
		OutputPaths:      []string{"stdout"},
		ErrorOutputPaths: []string{"stderr"},
		InitialFields:    map[string]interface{}{"platform": "server"},
	}

	if opts.Debug {
		config.Encoding = "console"
		config.EncoderConfig.EncodeLevel = zapcore.CapitalLevelEncoder
		config.EncoderConfig.EncodeTime = zapcore.TimeEncoderOfLayout("2006-01-02 15:04:05.000")
	}

	zapLogger, err = config.Build()
	if err != nil {
		panic(err)
	}

	logger := zapr.NewLogger(zapLogger)

	return &Logger{
		Logger: logger,
	}
}

func (l *Logger) Warn(msg string, keysAndValues ...interface{}) {
	l.Logger.V(0).Info(msg, keysAndValues...)
}
