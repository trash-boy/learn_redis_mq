package log

import (
	"context"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
	lumberjack "gopkg.in/natefinch/lumberjack.v2"
)
type Logger interface {
	Error(v ...interface{})
	Warn(v ...interface{})
	Info(v ...interface{})
	Debug(v ...interface{})
	Errorf(format string, v ...interface{})
	Warnf(format string, v ...interface{})
	Infof(format string, v ...interface{})
	Debugf(format string, v ...interface{})
}

var (
	defaultLogger Logger
)
func init(){
	defaultLogger = newSugarLogger(NewOptions())
}

type Options struct {
	LogName string
	LogLevel string
	FileName string
	MaxAge int
	MaxSize int
	MaxBackups int
	Compress bool
}

type Option func(*Options)

func NewOptions(opts ...Option)Options{
	options := Options{
		LogName: "app",
		LogLevel: "info",
		FileName: "app.log",
		MaxAge: 10,
		MaxSize: 100,
		MaxBackups: 3,
		Compress: true,
	}

	for _,opt := range opts{
		opt(&options)
	}

	return options
}

func WithLogLevel(level string)Option{
	return func(options *Options) {
		options.LogLevel = level
	}
}

func WithFileName(filename string)Option{
	return func(options *Options) {
		options.FileName = filename
	}
}

var Levels = map[string]zapcore.Level{
	"":zapcore.DebugLevel,
	"debug":zapcore.DebugLevel,
	"info":zapcore.InfoLevel,
	"warn":zapcore.WarnLevel,
	"error":zapcore.ErrorLevel,
	"fatal":zapcore.FatalLevel,
}

type zapLoggerWrapper struct {
	*zap.SugaredLogger
	options Options
}

func (w *zapLoggerWrapper) getEncoder() zapcore.Encoder {
	encodeConfig := zap.NewProductionEncoderConfig()
	encodeConfig.EncodeTime = zapcore.ISO8601TimeEncoder

	encodeConfig.EncodeLevel = zapcore.CapitalLevelEncoder
	return zapcore.NewConsoleEncoder(encodeConfig)
}

func (w *zapLoggerWrapper) getLogWriter() zapcore.WriteSyncer{
	return zapcore.AddSync(&lumberjack.Logger{
		Filename: w.options.FileName,
		MaxAge: w.options.MaxAge,
		MaxSize: w.options.MaxSize,
		MaxBackups: w.options.MaxBackups,
		Compress: w.options.Compress,
	})
}

func newSugarLogger(options Options)*zapLoggerWrapper{
	w := &zapLoggerWrapper{
		options: options,
	}
	encoder := w.getEncoder()
	writeSyncer := w.getLogWriter()
	core := zapcore.NewCore(encoder, writeSyncer, Levels[options.LogLevel])
	w.SugaredLogger = zap.New(core, zap.AddCaller(),zap.AddCallerSkip(1)).Sugar()
	return w
}


func GetDefaultLogger()Logger{
	return defaultLogger
}

func Debugf(format string,args ...interface{}){
	GetDefaultLogger().Debugf(format,args)
}

func Infof(format string,args ...interface{}){
	GetDefaultLogger().Infof(format,args)
}

func Warnf(format string,args ...interface{}){
	GetDefaultLogger().Warnf(format,args)
}

func Errorf(format string,args ...interface{}){
	GetDefaultLogger().Errorf(format,args)
}

// DebugContext 打印 Debug 日志
func DebugContext(ctx context.Context, args ...interface{}) {
	GetDefaultLogger().Debug(args...)
}

// DebugContextf 打印 Debug 日志
func DebugContextf(ctx context.Context, format string, args ...interface{}) {
	GetDefaultLogger().Debugf(format, args...)
}

// InfoContext 打印 Info 日志
func InfoContext(ctx context.Context, args ...interface{}) {
	GetDefaultLogger().Info(args...)
}

// InfoContextf 打印 Info 日志
func InfoContextf(ctx context.Context, format string, args ...interface{}) {
	GetDefaultLogger().Infof(format, args...)
}

// WarnContext 打印 Warn 日志
func WarnContext(ctx context.Context, args ...interface{}) {
	GetDefaultLogger().Warn(args...)
}

// WarnContextf 打印 Warn 日志
func WarnContextf(ctx context.Context, format string, args ...interface{}) {
	GetDefaultLogger().Warnf(format, args...)
}

// ErrorContext 打印 Error 日志
func ErrorContext(ctx context.Context, args ...interface{}) {
	GetDefaultLogger().Error(args...)
}

func ErrorContextf(ctx context.Context, format string, args ...interface{}) {
	GetDefaultLogger().Errorf(format, args...)
}

func Fatalf(format string, args ...interface{}) {
	Errorf(format, args...)
}





