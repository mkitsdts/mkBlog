package log

import (
	"log"
)

// 日志级别
type Level int8

const (
	DebugLevel = iota
	InfoLevel
	WarnLevel
	ErrorLevel
	FatalLevel
	PanicLevel
)

// 日志结构体

type Logger struct {
	*log.Logger
}

func (l *Logger) Debug(v ...interface{}) {

}

func (l *Logger) Debugf(format string, v ...interface{}) {

}

func (l *Logger) Info(v ...interface{}) {

}

func (l *Logger) Infof(format string, v ...interface{}) {

}

func (l *Logger) Warn(v ...interface{}) {

}

func (l *Logger) Warnf(format string, v ...interface{}) {

}

func (l *Logger) Error(v ...interface{}) {

}

func (l *Logger) Errorf(format string, v ...interface{}) {

}

func (l *Logger) Panic(v ...interface{}) {

}

func (l *Logger) Panicf(format string, v ...interface{}) {

}
