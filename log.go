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
// and has the levels, "trace", "info", "warn", "error", "panic",
// and "fatal".
//
// It always logs the level, file name, line number, and timestamp
// in unix nano seconds (UTC) as metadata.
type Logger struct {
	callDepth       int
	logger          *log.Logger
	permanentFields Fields
}

// Fields holds key-value pairs for logs.
type Fields map[string]interface{}

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
	panicLevel level = "panic"
	fatalLevel level = "fatal"
)

var defaultLogger = New(DefaultCallDepth+1, os.Stdout, nil)

// Trace calls the default Logger's Trace method.
func Trace(msg interface{}) {
	defaultLogger.Trace(msg)
}

// Tracef calls the default Logger's Tracef method.
func Tracef(f Fields, msg interface{}) {
	defaultLogger.Tracef(f, msg)
}

// Info calls the default Logger's Info method.
func Info(msg interface{}) {
	defaultLogger.Info(msg)
}

// Infof calls the default Logger's Infof method.
func Infof(f Fields, msg interface{}) {
	defaultLogger.Infof(f, msg)
}

// Warn calls the default Logger's Warn method.
func Warn(msg interface{}) {
	defaultLogger.Warn(msg)
}

// Warnf calls the default Logger's Warnf method.
func Warnf(f Fields, msg interface{}) {
	defaultLogger.Warnf(f, msg)
}

// Error calls the default Logger's Error method.
func Error(msg interface{}) {
	defaultLogger.Error(msg)
}

// Errorf calls the default Logger's Errorf method.
func Errorf(f Fields, msg interface{}) {
	defaultLogger.Errorf(f, msg)
}

// Panic calls the default Logger's Panic method.
func Panic(msg interface{}) {
	defaultLogger.Panic(msg)
}

// Panicf calls the default Logger's Panicf method.
func Panicf(f Fields, msg interface{}) {
	defaultLogger.Panicf(f, msg)
}

// Fatal calls the default Logger's Fatal method.
func Fatal(msg interface{}) {
	defaultLogger.Fatal(msg)
}

// Fatalf calls the default Logger's Fatalf method.
func Fatalf(f Fields, msg interface{}) {
	defaultLogger.Fatalf(f, msg)
}

// Trace logs a message at the trace level.
func (l *Logger) Trace(msg interface{}) {
	l.log(traceLevel, nil, msg)
}

// Tracef logs fields and a message at the trace level.
func (l *Logger) Tracef(f Fields, msg interface{}) {
	l.log(traceLevel, f, msg)
}

// Info logs a message at the info level.
func (l *Logger) Info(msg interface{}) {
	l.log(infoLevel, nil, msg)
}

// Infof logs fields and a message at the info level.
func (l *Logger) Infof(f Fields, msg interface{}) {
	l.log(infoLevel, f, msg)
}

// Warn logs a message at the warn level.
func (l *Logger) Warn(msg interface{}) {
	l.log(warnLevel, nil, msg)
}

// Warnf logs fields and a message at the warn level.
func (l *Logger) Warnf(f Fields, msg interface{}) {
	l.log(warnLevel, f, msg)
}

// Error logs a message at the error level.
func (l *Logger) Error(msg interface{}) {
	l.log(errorLevel, nil, msg)
}

// Errorf logs fields and a message at the error level.
func (l *Logger) Errorf(f Fields, msg interface{}) {
	l.log(errorLevel, f, msg)
}

// Panic logs a message at the panic level and then panics with the message.
func (l *Logger) Panic(msg interface{}) {
	l.log(panicLevel, nil, msg)
}

// Panicf logs fields and a message at the panic level and then panics with the fields and message.
func (l *Logger) Panicf(f Fields, msg interface{}) {
	l.log(panicLevel, f, msg)
}

// Fatal logs a message at the fatal level followed by os.Exit(1).
func (l *Logger) Fatal(msg interface{}) {
	l.log(fatalLevel, nil, msg)
	os.Exit(1)
}

// Fatalf logs fields and a message at the fatal level followed by os.Exit(1).
func (l *Logger) Fatalf(f Fields, msg interface{}) {
	l.log(fatalLevel, f, msg)
	os.Exit(1)
}

type event struct {
	Metadata Fields      `json:"_metadata"`
	Fields   Fields      `json:"fields,omitempty"`
	Message  interface{} `json:"message"`
}

func (l *Logger) log(lv level, f Fields, msg interface{}) {
	combinedFields := Fields{}

	for k, v := range f {
		if v == nil {
			v = "nil"
		}
		combinedFields[k] = fmt.Sprint(v)
	}

	for k, v := range l.permanentFields {
		if v == nil {
			v = "nil"
		}
		combinedFields[k] = fmt.Sprint(v)
	}

	if msg == nil {
		msg = "nil"
	}

	e := &event{
		Metadata: Fields{
			"level": string(lv),
			"file":  l.fileInfo(),
			"time":  time.Now().UTC().Format(time.RFC3339Nano),
		},
		Fields:  combinedFields,
		Message: fmt.Sprint(msg),
	}

	byt, _ := json.Marshal(e)
	es := string(byt)
	l.logger.Output(l.callDepth, es)

	if lv == panicLevel {
		panic(es)
	}
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
