// Package slog provides a structured logger that wraps the standard library's
// log package.
package slog

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"os"
	"runtime"
	"strings"
	"time"
)

// DefaultCallDepth is the number of stack frames to ascend in
// a goroutine and is used to determine the file name
// and line number to log.
//
// The value is sensible - importing this package and using
// it will produce the file name and line number where the
// Logger was used.
//
// If you would like to create a wrapper around this package,
// you will need to create the Logger with DefaultCallDepth+1
// when calling New.
//
// For more information, see the documentation for the standard
// library's runtime.Caller function.
const DefaultCallDepth = 3

// Logger is a wrapper around the standard library's log.Logger.
// It produces structured log messages as JSON key-value string pairs
// and has four levels, "trace", "info", "warn", and "error".
//
// It always logs the level, file name, line number, and timestamp
// in unix nano seconds (UTC) as metadata.
type Logger struct {
	callDepth       int
	logger          *log.Logger
	permanentFields Fields
}

// Fields holds key-value string pairs for logs.
type Fields map[string]string

// New returns a Logger that determines the file name and line number
// from callDepth, where to write out, and fields to permanently set that will
// appear with every log.
//
// If out is nil, it will default to os.Stdout.
//
// If permanentFields contains a key that is equal to
// a key in another method such as Infof, the permanentFields
// value will take priority.
func New(callDepth int, out io.Writer, permanentFields Fields) *Logger {
	if out == nil {
		out = os.Stdout
	}

	return &Logger{
		callDepth:       callDepth,
		logger:          log.New(out, "", 0),
		permanentFields: permanentFields,
	}
}

type level string

const (
	traceLevel level = "trace"
	infoLevel  level = "info"
	warnLevel  level = "warn"
	errorLevel level = "error"
)

var defaultLogger = New(DefaultCallDepth+1, os.Stdout, nil)

// Trace calls the default Logger's Trace method.
func Trace(msg string) {
	defaultLogger.Trace(msg)
}

// Tracef calls the default Logger's Tracef method.
func Tracef(f Fields, msg string) {
	defaultLogger.Tracef(f, msg)
}

// Info calls the default Logger's Info method.
func Info(msg string) {
	defaultLogger.Info(msg)
}

// Infof calls the default Logger's Infof method.
func Infof(f Fields, msg string) {
	defaultLogger.Infof(f, msg)
}

// Warn calls the default Logger's Warn method.
func Warn(msg string) {
	defaultLogger.Warn(msg)
}

// Warnf calls the default Logger's Warnf method.
func Warnf(f Fields, msg string) {
	defaultLogger.Warnf(f, msg)
}

// Error calls the default Logger's Error method.
func Error(msg string) {
	defaultLogger.Error(msg)
}

// Errorf calls the default Logger's Errorf method.
func Errorf(f Fields, msg string) {
	defaultLogger.Errorf(f, msg)
}

// Trace logs a message at the trace level.
func (l *Logger) Trace(msg string) {
	l.log(traceLevel, nil, msg)
}

// Tracef logs fields and a message at the trace level.
func (l *Logger) Tracef(f Fields, msg string) {
	l.log(traceLevel, f, msg)
}

// Info logs a message at the info level.
func (l *Logger) Info(msg string) {
	l.log(infoLevel, nil, msg)
}

// Infof logs fields and a message at the info level.
func (l *Logger) Infof(f Fields, msg string) {
	l.log(infoLevel, f, msg)
}

// Warn logs a message at the warn level.
func (l *Logger) Warn(msg string) {
	l.log(warnLevel, nil, msg)
}

// Warnf logs fields and a message at the warn level.
func (l *Logger) Warnf(f Fields, msg string) {
	l.log(warnLevel, f, msg)
}

// Error logs a message at the error level.
func (l *Logger) Error(msg string) {
	l.log(errorLevel, nil, msg)
}

// Errorf logs fields and a message at the error level.
func (l *Logger) Errorf(f Fields, msg string) {
	l.log(errorLevel, f, msg)
}

type event struct {
	Metadata Fields `json:"_metadata"`
	Fields   Fields `json:"fields,omitempty"`
	Message  string `json:"message"`
}

func (l *Logger) log(lv level, f Fields, msg string) {
	combinedFields := Fields{}

	for k, v := range f {
		combinedFields[k] = v
	}

	for k, v := range l.permanentFields {
		combinedFields[k] = v
	}

	e := &event{
		Metadata: Fields{
			"level": string(lv),
			"file":  l.fileInfo(),
			"time":  time.Now().UTC().Format(time.RFC3339Nano),
		},
		Fields:  combinedFields,
		Message: msg,
	}

	byt, _ := json.Marshal(e)
	l.logger.Output(l.callDepth, string(byt))
}

func (l *Logger) fileInfo() string {
	_, file, line, ok := runtime.Caller(l.callDepth)
	if !ok {
		file = "?"
		line = 0
	} else {
		slash := strings.LastIndex(file, "/")
		if slash >= 0 {
			file = file[slash+1:]
		}
	}

	return fmt.Sprintf("%s:%d", file, line)
}
