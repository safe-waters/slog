package slog

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

type mockWriter struct{ byt []byte }

func (m *mockWriter) Write(p []byte) (n int, err error) {
	m.byt = p
	return len(m.byt), nil
}

func TestLog(t *testing.T) {
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
	}

	for _, test := range tests {
		var (
			test = test
			mw   = &mockWriter{}
			l    = New(DefaultSkip, mw, test.permF)
			fn   func(msg string)
		)

		if test.f == nil {
			fn = getLogFunc(t, test.lv, l, test.msg)
		} else {
			fn = getLogFuncf(t, test.lv, test.f, l, test.msg)
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

			file := e.Metadata["file"]
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

			unixTimeInt, err := strconv.ParseInt(
				e.Metadata["time"],
				10,
				64,
			)
			if err != nil {
				t.Fatal(err)
			}

			gotTime := time.Unix(0, unixTimeInt).Round(time.Minute)
			expTime := time.Now().Round(time.Minute)
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

	l := New(DefaultSkip, nil, nil)
	w, ok := l.logger.Writer().(*os.File)
	if !ok || w != os.Stdout {
		t.Fatal(
			"expected New's Writer to default to os.Stdout, but it did not",
		)
	}
}

func TestUnknownStack(t *testing.T) {
	t.Parallel()

	mw := &mockWriter{}
	l := New(10000000, mw, nil)
	l.Info("hello world")

	var e event
	if err := json.Unmarshal(mw.byt, &e); err != nil {
		t.Fatal(err)
	}

	file := e.Metadata["file"]
	expFile := "?:0"
	if expFile != file {
		t.Fatalf(
			"expected file '%s', got '%s'",
			expFile,
			file,
		)
	}
}

func TestDefaultLogger(t *testing.T) {
	t.Parallel()

	expect := func(mw *mockWriter, l level, f Fields) {
		var e event
		if err := json.Unmarshal(mw.byt, &e); err != nil {
			t.Fatal(err)
		}

		if string(l) != e.Metadata["level"] {
			t.Fatalf("expected level '%s', got '%s'",
				l,
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
	}

	var (
		msg    = "hello"
		fields = Fields{"hello": "world"}
	)

	mw := &mockWriter{}
	defaultLogger.logger.SetOutput(mw)

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
}

func getLogFunc(
	t *testing.T,
	lv level,
	l *Logger,
	msg string,
) func(msg string) {
	t.Helper()

	switch lv {
	case "info":
		return l.Info
	case "warn":
		return l.Warn
	case "error":
		return l.Error
	default:
		t.Fatalf("invalid level '%s'", lv)
	}

	return nil
}

func getLogFuncf(
	t *testing.T,
	lv level,
	f Fields,
	l *Logger,
	msg string,
) func(msg string) {
	t.Helper()

	var fn func(f Fields, msg string)

	switch lv {
	case "info":
		fn = l.Infof
	case "warn":
		fn = l.Warnf
	case "error":
		fn = l.Errorf
	default:
		t.Fatalf("invalid level '%s'", lv)
		return nil
	}

	return func(msg string) {
		fn(f, msg)
	}
}
