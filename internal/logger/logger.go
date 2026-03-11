package logger

import (
	"fmt"
	"io"
	"log"
	"os"
)

type Param struct {
	Key   string
	Value interface{}
}

func (p Param) String() string {
	return fmt.Sprintf("%s=%v", p.Key, p.Value)
}

type Logger interface {
	Debug(msg string, params ...Param)
	Info(msg string, params ...Param)
	Error(msg string, params ...Param)
}

type StandardLoggerConfig struct {
	ShowTimestamp bool
	Debug         bool
	Out           io.Writer
}

type StandardLogger struct {
	debug  bool
	stdlog *log.Logger
}

func (sl StandardLogger) Debug(msg string, params ...Param) {
	if sl.debug {
		sl.logWithPrefix("debug", msg, params)
	}
}

func (sl StandardLogger) Info(msg string, params ...Param) {
	sl.logWithPrefix("info", msg, params)
}

func (sl StandardLogger) Error(msg string, params ...Param) {
	sl.logWithPrefix("error", msg, params)
}

func (sl StandardLogger) logWithPrefix(prefix string, msg string, params []Param) {
	var s string

	s += fmt.Sprintf("[%s] ", prefix)
	s += fmt.Sprintf("%s ", msg)

	for _, param := range params {
		s += fmt.Sprintf("%s ", param.String())
	}

	sl.stdlog.Println(s)
}

func NewStandardLogger(pkg string, cfg StandardLoggerConfig) StandardLogger {
	var flags int

	if cfg.ShowTimestamp {
		flags = log.LstdFlags
	}

	if cfg.Out == nil {
		cfg.Out = os.Stdout
	}

	return StandardLogger{
		debug:  cfg.Debug,
		stdlog: log.New(cfg.Out, fmt.Sprintf("%s ", pkg), flags),
	}
}
