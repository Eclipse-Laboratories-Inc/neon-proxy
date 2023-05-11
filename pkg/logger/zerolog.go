package logger

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"

	"github.com/rs/zerolog"
)

type ZeroLogger struct {
	logger *zerolog.Logger
}

type ZeroLogEvent struct {
	event *zerolog.Event
}

type ZeroLogContext struct {
	context *zerolog.Context
}

type GetLoggerFunc func(name string) Logger

var defaultLogger Logger

func SetDefaultLogger(logger Logger) {
	defaultLogger = logger
}

func Get() Logger {
	if defaultLogger != nil {
		return defaultLogger
	}

	InitDefaultLogger()

	return defaultLogger
}

func Debug() Event {
	return Get().Debug()
}

func Info() Event {
	return Get().Info()
}

func Error() Event {
	return Get().Error()
}

func InitDefaultLogger() {
	zl := zerolog.New(os.Stdout)
	defaultLogger = &ZeroLogger{
		logger: &zl,
	}
}

var getLoggerFunc GetLoggerFunc

func SetLoggerFunc(getFunc GetLoggerFunc) {
	getLoggerFunc = getFunc
}

func GetByName(name string) Logger {
	if getLoggerFunc != nil {
		return getLoggerFunc(name)
	}

	return Get()
}

func NewLogger(name string, settings LogSettings) (Logger, error) {
	level, levelParsingErr := zerolog.ParseLevel(settings.Level)
	if levelParsingErr != nil {
		level = zerolog.DebugLevel
	}

	var zl zerolog.Logger

	if len(settings.Path) == 0 {
		zl = zerolog.New(&bytes.Buffer{}).Output(zerolog.ConsoleWriter{Out: os.Stdout}).Level(level).With().Timestamp().Logger()
	} else {
		_ = os.Mkdir(settings.Path, 0744)
		file, err := os.OpenFile(filepath.Join(settings.Path, fmt.Sprintf("%s.log", name)), os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
		if err != nil {
			return nil, err
		}
		zl = zerolog.New(file).Level(level).With().Timestamp().Logger()
	}

	return &ZeroLogger{logger: &zl}, nil

}

func (zl *ZeroLogger) Debug() Event {
	return &ZeroLogEvent{event: zl.logger.Debug()}
}

func (zl *ZeroLogger) Info() Event {
	return &ZeroLogEvent{event: zl.logger.Info()}
}

func (zl *ZeroLogger) Error() Event {
	return &ZeroLogEvent{event: zl.logger.Error()}
}

func (zl *ZeroLogger) Warn() Event {
	return &ZeroLogEvent{event: zl.logger.Warn()}
}

func (zl *ZeroLogger) With() Context {
	w := zl.logger.With()
	return &ZeroLogContext{context: &w}
}

func (zle *ZeroLogEvent) Msg(msg string) {
	zle.event.Msg(msg)
}

func (zle *ZeroLogEvent) Msgf(format string, v ...interface{}) {
	zle.event.Msgf(format, v...)
}

func (zle *ZeroLogEvent) Err(err error) Event {
	message := "nil"
	if err != nil {
		message = err.Error()
	}

	return &ZeroLogEvent{event: zle.event.Interface("error", struct {
		Message string
	}{Message: message})}
}

func (zle *ZeroLogEvent) Str(key string, val string) Event {
	return &ZeroLogEvent{event: zle.event.Str(key, val)}
}

func (zle *ZeroLogEvent) Int(key string, i int) Event {
	return &ZeroLogEvent{event: zle.event.Int(key, i)}
}

func (zle *ZeroLogEvent) Float64(key string, f float64) Event {
	return &ZeroLogEvent{event: zle.event.Float64(key, f)}
}

func (zle *ZeroLogEvent) Bool(key string, b bool) Event {
	return &ZeroLogEvent{event: zle.event.Bool(key, b)}
}

func (zle *ZeroLogEvent) Interface(key string, i interface{}) Event {
	return &ZeroLogEvent{event: zle.event.Interface(key, i)}
}

func (zle *ZeroLogContext) Str(key string, val string) Context {
	ctx := zle.context.Str(key, val)

	return &ZeroLogContext{context: &ctx}
}

func (zle *ZeroLogContext) Int(key string, i int) Context {
	ctx := zle.context.Int(key, i)

	return &ZeroLogContext{context: &ctx}
}

func (zle *ZeroLogContext) Float64(key string, f float64) Context {
	ctx := zle.context.Float64(key, f)

	return &ZeroLogContext{context: &ctx}
}

func (zle *ZeroLogContext) Bool(key string, b bool) Context {
	ctx := zle.context.Bool(key, b)

	return &ZeroLogContext{context: &ctx}
}

func (zle *ZeroLogContext) Interface(key string, i interface{}) Context {
	ctx := zle.context.Interface(key, i)

	return &ZeroLogContext{context: &ctx}
}

func (zle *ZeroLogContext) Logger() Logger {
	l := zle.context.Logger()

	return &ZeroLogger{logger: &l}
}

func (s LogSettings) Init() error {
	log, err := NewLogger(s.Name, s)
	if err != nil {
		return err
	}

	SetDefaultLogger(log)

	return nil
}
