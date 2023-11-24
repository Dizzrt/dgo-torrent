package dlog

import (
	"fmt"
	"sync"
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	"gopkg.in/natefinch/lumberjack.v2"
)

const (
	logTimeFormat = "2006-01-02 15:04:05 Z07"

	logdir  = "./log"
	logFile = "log"
)

var (
	once   sync.Once
	logger *zap.SugaredLogger
)

func Init() {
	once.Do(func() {
		hook := &lumberjack.Logger{
			Filename:   fmt.Sprintf("%s/%s", logdir, logFile),
			MaxSize:    10,
			MaxBackups: 4,
			MaxAge:     30,
			LocalTime:  true,
			Compress:   false,
		}

		writeSyncer := zapcore.BufferedWriteSyncer{
			WS:   zapcore.AddSync(hook),
			Size: 4096,
		}

		cores := []zapcore.Core{
			zapcore.NewCore(encoder(), zapcore.AddSync(&writeSyncer), zapcore.InfoLevel),
		}

		tee := zapcore.NewTee(cores...)
		lg := zap.New(tee).WithOptions(zap.AddCallerSkip(1), zap.AddStacktrace(zapcore.ErrorLevel))

		logger = lg.Sugar()
	})
}

func encoder() zapcore.Encoder {
	return zapcore.NewConsoleEncoder(
		zapcore.EncoderConfig{
			TimeKey:        "time",
			LevelKey:       "level",
			NameKey:        "logger",
			CallerKey:      "caller",
			MessageKey:     "msg",
			StacktraceKey:  "stacktrace",
			FunctionKey:    zapcore.OmitKey,
			LineEnding:     zapcore.DefaultLineEnding,
			EncodeDuration: zapcore.MillisDurationEncoder,
			EncodeLevel:    levelEncoder,
			EncodeCaller:   callerEncoder,
			EncodeTime:     timeEncoder,
		})
}

func levelEncoder(level zapcore.Level, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString("[" + level.CapitalString() + "]")
}

func timeEncoder(t time.Time, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString("[" + t.Format(logTimeFormat) + "]")
}

func callerEncoder(caller zapcore.EntryCaller, enc zapcore.PrimitiveArrayEncoder) {
	enc.AppendString("[" + caller.TrimmedPath() + "]")
}

func L() *zap.SugaredLogger {
	return logger
}

func Info(args ...interface{}) {
	logger.Info(args...)
}

func Infof(template string, args ...interface{}) {
	logger.Infof(template, args...)
}

func Warn(args ...interface{}) {
	logger.Warn(args...)
}

func Warnf(template string, args ...interface{}) {
	logger.Warnf(template, args...)
}

func Error(args ...interface{}) {
	logger.Error(args...)
}

func Errorf(template string, args ...interface{}) {
	logger.Errorf(template, args...)
}

func Panic(args ...interface{}) {
	logger.Panic(args...)
}

func Panicf(template string, args ...interface{}) {
	logger.Panicf(template, args...)
}

func DPanic(args ...interface{}) {
	logger.DPanic(args...)
}

func DPanicf(template string, args ...interface{}) {
	logger.DPanicf(template, args...)
}

func Fatal(args ...interface{}) {
	logger.Fatal(args...)
}

func Fatalf(template string, args ...interface{}) {
	logger.Fatalf(template, args...)
}
