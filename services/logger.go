package services

import (
	"fmt"
	"io"
	"log"
	"os"
)

type LogLevel uint32

const (
	ErrorLevel LogLevel = iota
	WarnLevel
	InfoLevel
	DebugLevel
)

type Logger interface {
	Fatal(v ...any)
	Error(v ...any)
	Warn(v ...any)
	Info(v ...any)
	Debug(v ...any)
}

type logger struct {
	level LogLevel

	errors *log.Logger
	warns  *log.Logger
	infos  *log.Logger
	debugs *log.Logger
}

func NewLogger(out io.Writer, appName string, levelName string) (Logger, error) {
	flag := log.Ldate | log.Ltime | log.Llongfile

	level, err := ParseLogLevel(levelName)
	if err != nil {
		return nil, err
	}

	return &logger{
		level:  level,
		errors: log.New(out, fmt.Sprintf("%s.ERROR: ", appName), flag),
		warns:  log.New(out, fmt.Sprintf("%s.WARN: ", appName), flag),
		infos:  log.New(out, fmt.Sprintf("%s.INFO: ", appName), flag),
		debugs: log.New(out, fmt.Sprintf("%s.DEBUG: ", appName), flag),
	}, nil
}

func (l *logger) Fatal(v ...any) {
	l.errors.Output(2, fmt.Sprintln(v...))
	os.Exit(1)
}

func (l *logger) Error(v ...any) {
	l.errors.Println(v...)
}

func (l *logger) Warn(v ...any) {
	if l.level < WarnLevel {
		return
	}
	l.warns.Output(2, fmt.Sprintln(v...))
}

func (l *logger) Info(v ...any) {
	if l.level < InfoLevel {
		return
	}
	l.infos.Output(2, fmt.Sprintln(v...))
}

func (l *logger) Debug(v ...any) {
	if l.level < DebugLevel {
		return
	}
	l.debugs.Output(2, fmt.Sprintln(v...))
}

func ParseLogLevel(level string) (lvl LogLevel, err error) {
	switch level {
	case "error":
		lvl = ErrorLevel
	case "warn":
		lvl = WarnLevel
	case "info":
		lvl = InfoLevel
	case "debug":
		lvl = DebugLevel
	default:
		err = fmt.Errorf("Invalid log level")
	}
	return
}
