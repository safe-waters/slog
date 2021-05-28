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
	expectedMetadataKeys := []string{"level", "file", "time"}

	tests := []struct {
		name            string
		level           level
		fields          Fields
		permanentFields Fields
		msg             string
		expectedFields  Fields
		expectedKeys    []string
	}{
		{
			name:         "info",
			level:        infoLevel,
			expectedKeys: []string{"_metadata", "message"},
			msg:          "hello",
		},
		{
			name:           "info fields",
			level:          infoLevel,
			fields:         Fields{"test": "message"},
			expectedFields: Fields{"test": "message"},
			expectedKeys:   []string{"_metadata", "message", "fields"},
			msg:            "hello",
		},
		{
			name:            "info permanent fields",
			level:           infoLevel,
			fields:          Fields{"test": "shadow", "local": "message"},
			permanentFields: Fields{"test": "message"},
			expectedFields:  Fields{"test": "message", "local": "message"},
			expectedKeys:    []string{"_metadata", "message", "fields"},
			msg:             "hello",
		},
		{
			name:         "warn",
			level:        warnLevel,
			expectedKeys: []string{"_metadata", "message"},
			msg:          "hello",
		},
		{
			name:           "warn fields",
			level:          warnLevel,
			fields:         Fields{"test": "message"},
			expectedFields: Fields{"test": "message"},
			expectedKeys:   []string{"_metadata", "message", "fields"},
			msg:            "hello",
		},
		{
			name:            "warn permanent fields",
			level:           warnLevel,
			fields:          Fields{"test": "shadow", "local": "message"},
			permanentFields: Fields{"test": "message"},
			expectedFields:  Fields{"test": "message", "local": "message"},
			expectedKeys:    []string{"_metadata", "message", "fields"},
			msg:             "hello",
		},
		{
			name:         "error",
			level:        errorLevel,
			expectedKeys: []string{"_metadata", "message"},
			msg:          "hello",
		},
		{
			name:           "error fields",
			level:          errorLevel,
			fields:         Fields{"test": "message"},
			expectedFields: Fields{"test": "message"},
			expectedKeys:   []string{"_metadata", "message", "fields"},
			msg:            "hello",
		},
		{
			name:            "error permanent fields",
			level:           errorLevel,
			fields:          Fields{"test": "shadow", "local": "message"},
			permanentFields: Fields{"test": "message"},
			expectedFields:  Fields{"test": "message", "local": "message"},
			expectedKeys:    []string{"_metadata", "message", "fields"},
			msg:             "hello",
		},
	}

	for _, test := range tests {
		var (
			test = test
			mw   = &mockWriter{}
			l    = New(DefaultSkip, mw, test.permanentFields)
			fn   func(msg string)
		)

		if test.fields == nil {
			fn = getLogFunc(t, test.level, l, test.msg)
		} else {
			fn = getLogFuncf(t, test.level, test.fields, l, test.msg)
		}

		t.Run(test.name, func(t *testing.T) {
			t.Parallel()

			fn(test.msg)

			var raw map[string]interface{}
			if err := json.Unmarshal(mw.byt, &raw); err != nil {
				t.Fatal(err)
			}

			if len(raw) != len(test.expectedKeys) {
				t.Fatalf(
					"expected '%d' keys, got '%d'",
					len(test.expectedKeys),
					len(raw),
				)
			}

			for _, k := range test.expectedKeys {
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

			for _, k := range expectedMetadataKeys {
				if e.Metadata[k] == "" {
					t.Fatalf(
						"expected key '%s' to have a value",
						k,
					)
				}
			}

			if e.Metadata["level"] != string(test.level) {
				t.Fatalf(
					"expected level '%s', got '%s'",
					test.level,
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

			if len(test.expectedFields) != len(e.Fields) {
				t.Fatalf(
					"expected '%d' field(s), got '%d'",
					len(test.expectedFields),
					len(e.Fields),
				)
			}

			for k := range test.expectedFields {
				if test.expectedFields[k] != e.Fields[k] {
					t.Fatalf("expected field '%s', got '%s'",
						test.expectedFields[k],
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
	concrete, ok := l.logger.Writer().(*os.File)
	if !ok || concrete != os.Stdout {
		t.Fatal("expected New's Writer to default to os.Stdout, but it did not")
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

	expect := func(mw *mockWriter, msg string, l level, f Fields) {
		var e event
		if err := json.Unmarshal(mw.byt, &e); err != nil {
			t.Fatal(err)
		}

		if e.Message != msg {
			t.Fatalf("expected message '%s', got '%s'",
				e.Message,
				msg,
			)
		}

		if len(f) != len(e.Fields) {
			t.Fatalf(
				"expected '%d' field(s), got '%d'",
				len(f),
				len(e.Fields),
			)
		}

		for k := range f {
			if f[k] != e.Fields[k] {
				t.Fatalf("expected field '%s', got '%s'",
					f[k],
					e.Fields[k],
				)
			}
		}
	}

	var (
		msg    = "hello"
		fields = Fields{"hello": "world"}
	)

	mw := &mockWriter{}
	defaultLogger.logger.SetOutput(mw)

	Info(msg)
	expect(mw, msg, infoLevel, nil)

	Infof(fields, msg)
	expect(mw, msg, infoLevel, fields)

	Warn(msg)
	expect(mw, msg, warnLevel, nil)

	Warnf(fields, msg)
	expect(mw, msg, warnLevel, fields)

	Error(msg)
	expect(mw, msg, errorLevel, nil)

	Errorf(fields, msg)
	expect(mw, msg, errorLevel, fields)
}

func getLogFunc(t *testing.T, lv level, l *Logger, msg string) func(msg string) {
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

func getLogFuncf(t *testing.T, lv level, f Fields, l *Logger, msg string) func(msg string) {
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
