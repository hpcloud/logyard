package log2

import (
	"fmt"
	"log"
	"os"
	"runtime/debug"
)

type Logger struct {
	*log.Logger
}

var std = &Logger{log.New(os.Stderr, "", log.LstdFlags)}

func Info(v ...interface{}) {
	std.Output(2, fmt.Sprintf("[INFO] %s", v...))
}

func Infof(format string, v ...interface{}) {
	std.Output(2, fmt.Sprintf("[INFO] "+format, v...))
}

func Warn(v ...interface{}) {
	std.Output(2, fmt.Sprintf("[WARN] %s", v...))
}

func Warnf(format string, v ...interface{}) {
	std.Output(2, fmt.Sprintf("[WARN] "+format, v...))
}

func Error(v ...interface{}) {
	std.Output(2, fmt.Sprintf("[ERROR] %s", v...))
}

func Errorf(format string, v ...interface{}) {
	std.Output(2, fmt.Sprintf("[ERROR] "+format, v...))
}

func Fatal(v ...interface{}) {
	std.Output(2, fmt.Sprintf("[FATAL] %s", v...)+"\n"+string(debug.Stack()))
	os.Exit(1)
}

func Fatalf(format string, v ...interface{}) {
	std.Output(2, fmt.Sprintf("[FATAL] "+format, v...)+"\n"+string(debug.Stack()))
	os.Exit(1)
}

func (l *Logger) Info(v ...interface{}) {
	l.Output(2, fmt.Sprintf("[INFO] %s", v...))
}

func (l *Logger) Infof(format string, v ...interface{}) {
	l.Output(2, fmt.Sprintf("[INFO] "+format, v...))
}

func (l *Logger) Warn(v ...interface{}) {
	l.Output(2, fmt.Sprintf("[WARN] %s", v...))
}

func (l *Logger) Warnf(format string, v ...interface{}) {
	l.Output(2, fmt.Sprintf("[WARN] "+format, v...))
}

func (l *Logger) Error(v ...interface{}) {
	l.Output(2, fmt.Sprintf("[ERROR] %s", v...))
}

func (l *Logger) Errorf(format string, v ...interface{}) {
	l.Output(2, fmt.Sprintf("[ERROR] "+format, v...))
}

// Fatal behaves mostly like log.Fatal but it also prints the current
// goroutine's stacktrace.
func (l *Logger) Fatal(v ...interface{}) {
	l.Output(2, fmt.Sprintf("[FATAL] %s", v...)+"\n"+string(debug.Stack()))
	os.Exit(1)
}

func (l *Logger) Fatalf(format string, v ...interface{}) {
	l.Output(2, fmt.Sprintf("[FATAL] "+format, v...)+"\n"+string(debug.Stack()))
	os.Exit(1)
}

func New() *Logger {
	return &Logger{log.New(os.Stderr, "", log.LstdFlags)}
}
