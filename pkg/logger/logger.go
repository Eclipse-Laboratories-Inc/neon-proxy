package logger

type LogSettings struct {
	Name  string `yaml:"name"`
	Level string `yaml:"level"`
	Path  string `yaml:"path"`
}

type Event interface {
	Msg(msg string)
	Msgf(format string, v ...interface{})
	Err(err error) Event
	Str(key string, val string) Event
	Int(key string, i int) Event
	Float64(key string, f float64) Event
	Bool(key string, b bool) Event
	Interface(key string, i interface{}) Event
}

type Logger interface {
	Debug() Event
	Info() Event
	Error() Event
	With() Context
}

type Context interface {
	Str(key string, val string) Context
	Int(key string, i int) Context
	Float64(key string, f float64) Context
	Bool(key string, b bool) Context
	Interface(key string, i interface{}) Context
	Logger() Logger
}
