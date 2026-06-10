package logger

import (
	"fmt"
	"os"
	"time"
)

type Logger struct {
	subSystem string
}

type logLevel string

const (
	InfoLevel logLevel = "INFO"
	WarnLevel logLevel = "WARN"
	ErrorLevel logLevel = "ERROR"
	DebugLevel logLevel = "DEBUG"
)

func New(subSystem string) *Logger {
	return &Logger{subSystem: subSystem}
}

func (logger *Logger) Info(format string, args ...any) {
	logger.log(InfoLevel, format, args...)
}

func (logger *Logger) Warn(format string, args ...any) {
	logger.log(WarnLevel, format, args...)
}

func (logger *Logger) Error(format string, args ...any) {
	logger.log(ErrorLevel, format, args...)
}

func (logger *Logger) Debug(format string, args ...any) {
	logger.log(DebugLevel, format, args...)
}

func (logger *Logger) log(level logLevel, format string, args ...any) {
	timeStamp := time.Now().Format("2006-01-02 15:04:05")
	message := fmt.Sprintf(format, args...)
	fmt.Fprintf(os.Stderr, "[%s] [%s] [%s] %s\n", timeStamp, level, logger.subSystem, message)
}