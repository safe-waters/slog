package slog

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

type mockWriter struct{ byt []byte }

func (m *mockWriter) Write(p []byte) (n int, err error) {
	m.byt = p
	return 0, nil
}

func TestLog(t *testing.T) {
	t.Parallel()

	expMetaKeys := []string{"level", "file", "time"}

	tests := []struct {
		name    string
		msg     string
		lv      level
		f       Fields
		permF   Fields
		expF    Fields
		expKeys []string
	}{
		{
			name:    "trace",
			msg:     "hello",
			lv:      traceLevel,
			expKeys: []string{"_metadata", "message"},
		},
		{
			name:    "trace fields",
			msg:     "hello",
			lv:      traceLevel,
			f:       Fields{"test": "message"},
			expF:    Fields{"test": "message"},
			expKeys: []string{"_metadata", "message", "fields"},
		},
		{
			name:    "trace permanent fields",
			msg:     "hello",
			lv:      traceLevel,
			f:       Fields{"test": "shadow", "local": "message"},
			permF:   Fields{"test": "message"},
			expF:    Fields{"test": "message", "local": "message"},
			expKeys: []string{"_metadata", "message", "fields"},
		},
		{
			name:    "info",
			msg:     "hello",
			lv:      infoLevel,
			expKeys: []string{"_metadata", "message"},
		},
		{
			name:    "info fields",
			msg:     "hello",
			lv:      infoLevel,
			f:       Fields{"test": "message"},
			expF:    Fields{"test": "message"},
			expKeys: []string{"_metadata", "message", "fields"},
		},
		{
			name:    "info permanent fields",
			msg:     "hello",
			lv:      infoLevel,
			f:       Fields{"test": "shadow", "local": "message"},
			permF:   Fields{"test": "message"},
			expF:    Fields{"test": "message", "local": "message"},
			expKeys: []string{"_metadata", "message", "fields"},
		},
		{
			name:    "warn",
			msg:     "hello",
			lv:      warnLevel,
			expKeys: []string{"_metadata", "message"},
		},
		{
			name:    "warn fields",
			msg:     "hello",
			lv:      warnLevel,
			f:       Fields{"test": "message"},
			expF:    Fields{"test": "message"},
			expKeys: []string{"_metadata", "message", "fields"},
		},
		{
			name:    "warn permanent fields",
			msg:     "hello",
			lv:      warnLevel,
			f:       Fields{"test": "shadow", "local": "message"},
			permF:   Fields{"test": "message"},
			expF:    Fields{"test": "message", "local": "message"},
			expKeys: []string{"_metadata", "message", "fields"},
		},
		{
			name:    "error",
			msg:     "hello",
			lv:      errorLevel,
			expKeys: []string{"_metadata", "message"},
		},
		{
			name:    "error fields",
			msg:     "hello",
			lv:      errorLevel,
			f:       Fields{"test": "message"},
			expF:    Fields{"test": "message"},
			expKeys: []string{"_metadata", "message", "fields"},
		},
		{
			name:    "error permanent fields",
			msg:     "hello",
			lv:      errorLevel,
			f:       Fields{"test": "shadow", "local": "message"},
			permF:   Fields{"test": "message"},
			expF:    Fields{"test": "message", "local": "message"},
			expKeys: []string{"_metadata", "message", "fields"},
		},
		{
			name:    "panic",
			msg:     "hello",
			lv:      errorLevel,
			expKeys: []string{"_metadata", "message"},
		},
		{
			name:    "panic fields",
			msg:     "hello",
			lv:      panicLevel,
			f:       Fields{"test": "message"},
			expF:    Fields{"test": "message"},
			expKeys: []string{"_metadata", "message", "fields"},
		},
		{
			name:    "panic permanent fields",
			msg:     "hello",
			lv:      panicLevel,
			f:       Fields{"test": "shadow", "local": "message"},
			permF:   Fields{"test": "message"},
			expF:    Fields{"test": "message", "local": "message"},
			expKeys: []string{"_metadata", "message", "fields"},
		},
	}

	for _, test := range tests {
		var (
			test = test
			mw   = &mockWriter{}
			l    = New(mw, test.permF)
			fn   func(msg interface{})
		)

		if test.f == nil {
			fn = getLogFunc(t, l, test.lv, test.msg)
		} else {
			fn = getLogFuncf(t, l, test.lv, test.f, test.msg)
		}

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			fn(test.msg)

			var raw map[string]interface{}
			if err := json.Unmarshal(mw.byt, &raw); err != nil {
				t.Fatal(err)
			}

			if len(raw) != len(test.expKeys) {
				t.Fatalf(
					"expected '%d' keys, got '%d'",
					len(test.expKeys),
					len(raw),
				)
			}

			for _, k := range test.expKeys {
				if _, ok := raw[k]; !ok {
					t.Fatalf(
						"expected key '%s' but it did not exist",
						k,
					)
				}
			}

			var e event
			if err := json.Unmarshal(mw.byt, &e); err != nil {
				t.Fatal(err)
			}

			for _, k := range expMetaKeys {
				if e.Metadata[k] == "" {
					t.Fatalf(
						"expected key '%s' to have a value",
						k,
					)
				}
			}

			if e.Metadata["level"] != string(test.lv) {
				t.Fatalf(
					"expected level '%s', got '%s'",
					test.lv,
					e.Metadata["level"],
				)
			}

			file := fmt.Sprint(e.Metadata["file"])
			expFileName := "log_test.go"
			if !strings.HasPrefix(file, expFileName) {
				t.Fatalf(
					"expected file to contain '%s', got '%s'",
					expFileName,
					file,
				)
			}

			numColon := strings.Count(file, ":")
			expNumColon := 1
			if numColon != expNumColon {
				t.Fatalf(
					"expected '%d' colon(s), got '%d'",
					expNumColon,
					numColon,
				)
			}

			lineNum := file[len(expFileName)+1:]
			_, err := strconv.Atoi(lineNum)
			if err != nil {
				t.Fatal(err)
			}

			gotTime, err := time.Parse(time.RFC3339, fmt.Sprint(e.Metadata["time"]))
			if err != nil {
				t.Fatal(err)
			}

			gotTime = gotTime.Round(time.Minute)
			expTime := time.Now().UTC().Round(time.Minute)
			if !expTime.Equal(gotTime) {
				t.Fatalf("expected '%s' time, got '%s'", expTime, gotTime)
			}

			if test.msg != e.Message {
				t.Fatalf(
					"expected message '%s', got '%s'",
					test.msg,
					e.Message,
				)
			}

			if len(test.expF) != len(e.Fields) {
				t.Fatalf(
					"expected '%d' field(s), got '%d'",
					len(test.expF),
					len(e.Fields),
				)
			}

			for k := range test.expF {
				if test.expF[k] != e.Fields[k] {
					t.Fatalf("expected field '%s', got '%s'",
						test.expF[k],
						e.Fields[k],
					)
				}
			}
		})
	}
}

func TestDefaultStdOut(t *testing.T) {
	t.Parallel()

	l := New(nil, nil)
	w, ok := l.logger.Writer().(*os.File)
	if !ok || w != os.Stdout {
		t.Fatal(
			"expected New's Writer to default to os.Stdout, but it did not",
		)
	}
}

func TestDefaultLogger(t *testing.T) {
	t.Parallel()

	expect := func(mw *mockWriter, lv level, f Fields) {
		var e event
		if err := json.Unmarshal(mw.byt, &e); err != nil {
			t.Fatal(err)
		}

		if string(lv) != e.Metadata["level"] {
			t.Fatalf("expected level '%s', got '%s'",
				lv,
				e.Metadata["level"],
			)
		}

		if len(f) != len(e.Fields) {
			t.Fatalf(
				"expected '%d' field(s), got '%d'",
				len(f),
				len(e.Fields),
			)
		}

		file := fmt.Sprint(e.Metadata["file"])
		expFileName := "log_test.go"
		if !strings.HasPrefix(file, expFileName) {
			t.Fatalf(
				"expected file to contain '%s', got '%s'",
				expFileName,
				file,
			)
		}
	}

	var (
		msg    = "hello"
		fields = Fields{"hello": "world"}
	)

	mw := &mockWriter{}
	defaultLogger.logger.SetOutput(mw)

	Trace(msg)
	expect(mw, traceLevel, nil)

	Tracef(fields, msg)
	expect(mw, traceLevel, fields)

	Info(msg)
	expect(mw, infoLevel, nil)

	Infof(fields, msg)
	expect(mw, infoLevel, fields)

	Warn(msg)
	expect(mw, warnLevel, nil)

	Warnf(fields, msg)
	expect(mw, warnLevel, fields)

	Error(msg)
	expect(mw, errorLevel, nil)

	Errorf(fields, msg)
	expect(mw, errorLevel, fields)

	func() {
		defer func() {
			if r := recover(); r != nil {
				expect(mw, panicLevel, nil)
			}
		}()
		Panic(msg)
	}()

	func() {
		defer func() {
			if r := recover(); r != nil {
				expect(mw, panicLevel, fields)
			}
		}()
		Panicf(fields, msg)
	}()
}

func getLogFunc(
	t *testing.T,
	l *Logger,
	lv level,
	msg interface{},
) func(msg interface{}) {
	t.Helper()

	switch lv {
	case "trace":
		return l.Trace
	case "info":
		return l.Info
	case "warn":
		return l.Warn
	case "error":
		return l.Error
	case "panic":
		return func(msg interface{}) {
			defer func() {
				if r := recover(); r != nil {
					return
				}
			}()
			l.Panic(msg)
		}
	default:
		t.Fatalf("invalid level '%s'", lv)
	}

	return nil
}

func getLogFuncf(
	t *testing.T,
	l *Logger,
	lv level,
	f Fields,
	msg interface{},
) func(msg interface{}) {
	t.Helper()

	var fn func(f Fields, msg interface{})

	switch lv {
	case "trace":
		fn = l.Tracef
	case "info":
		fn = l.Infof
	case "warn":
		fn = l.Warnf
	case "error":
		fn = l.Errorf
	case "panic":
		fn = func(f Fields, msg interface{}) {
			defer func() {
				if r := recover(); r != nil {
					return
				}
			}()
			l.Panicf(f, msg)
		}
	default:
		t.Fatalf("invalid level '%s'", lv)
		return nil
	}

	return func(msg interface{}) {
		fn(f, msg)
	}
}
