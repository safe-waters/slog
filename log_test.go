package slog

import (
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"testing"
	"time"
)

type logLevelFunc func(msg string)

type mockWriter struct{ byt []byte }

func (m *mockWriter) Write(p []byte) (n int, err error) {
	m.byt = p
	return len(m.byt), nil
}

func TestLog(t *testing.T) {
	mw := &mockWriter{}
	defaultLogger = New(defaultLogger.skip, mw, nil)

	var (
		l   = New(DefaultSkip, mw, nil)
		fns = map[string][]logLevelFunc{
			"info":  {l.Info, Info},
			"warn":  {l.Warn, Warn},
			"error": {l.Error, Error},
		}
		expectedKeys         = []string{"_metadata", "message"}
		expectedMetadataKeys = []string{"level", "file", "time"}
		msg                  = "hello world"
	)

	for level, fns := range fns {
		for _, fn := range fns {
			level := level
			fn := fn

			t.Run(level, func(t *testing.T) {
				fn(msg)

				var raw map[string]interface{}
				if err := json.Unmarshal(mw.byt, &raw); err != nil {
					t.Fatal(err)
				}

				if len(raw) != len(expectedKeys) {
					t.Fatalf(
						"expected '%d' keys, got '%d'",
						len(expectedKeys),
						len(raw),
					)
				}

				for _, k := range expectedKeys {
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

				if e.Metadata["level"] != level {
					t.Fatalf(
						"expected level '%s', got '%s'",
						level,
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

				if msg != e.Message {
					t.Fatalf(
						"expected message '%s', got '%s'",
						msg,
						e.Message,
					)
				}
			})
		}
	}
}

type logLevelFuncf func(f Fields, msg string)

func TestLogf(t *testing.T) {
	mw := &mockWriter{}
	defaultLogger = New(defaultLogger.skip, mw, nil)

	var (
		l   = New(DefaultSkip, mw, nil)
		fns = map[string][]logLevelFuncf{
			"info":  {l.Infof, Infof},
			"warn":  {l.Warnf, Warnf},
			"error": {l.Errorf, Errorf},
		}
		fields = Fields{"key": "val"}
		msg    = "hello world"
	)

	for level, fns := range fns {
		for _, fn := range fns {
			level := level
			fn := fn

			t.Run(level, func(t *testing.T) {
				fn(fields, msg)

				var e event
				if err := json.Unmarshal(mw.byt, &e); err != nil {
					t.Fatal(err)
				}

				if e.Metadata["level"] != level {
					t.Fatalf(
						"expected level '%s', got '%s'",
						level,
						e.Metadata["level"],
					)
				}

				numFields := len(e.Fields)
				expNumFields := len(fields)
				if expNumFields != numFields {
					t.Fatalf(
						"expected '%d' fields, got '%d'",
						expNumFields,
						numFields,
					)
				}

				for k := range fields {
					if fields[k] != e.Fields[k] {
						t.Fatalf(
							"expected '%s' field for key '%s', got '%s'",
							fields[k],
							k,
							e.Fields[k],
						)
					}
				}
			})
		}
	}
}

func TestFields(t *testing.T) {
	var (
		mw              = &mockWriter{}
		permanentKey    = "permanent"
		permanentVal    = "override"
		permanentFields = Fields{permanentKey: permanentVal}
		localKey        = "local"
		localVal        = "unaffected"
		localFields     = Fields{permanentKey: "shadowed", localKey: localVal}
		l               = New(DefaultSkip, mw, permanentFields)
	)

	l.Infof(localFields, "hello world")

	var e event
	if err := json.Unmarshal(mw.byt, &e); err != nil {
		t.Fatal(err)
	}

	if len(localFields) != len(e.Fields) {
		t.Fatalf(
			"expected '%d' fields, got '%d'",
			len(localFields),
			len(e.Fields),
		)
	}

	if e.Fields[permanentKey] != permanentVal {
		t.Fatalf(
			"expected '%s' val for key '%s', got '%s'",
			permanentVal,
			permanentKey,
			e.Fields[permanentKey],
		)
	}

	if e.Fields[localKey] != localVal {
		t.Fatalf(
			"expected '%s' val for key '%s', got '%s'",
			localVal,
			localKey,
			e.Fields[localKey],
		)
	}
}

func TestDefaultStdOut(t *testing.T) {
	l := New(DefaultSkip, nil, nil)
	concrete, ok := l.logger.Writer().(*os.File)
	if !ok || concrete != os.Stdout {
		t.Fatal("expected New's Writer to default to os.Stdout, but it did not")
	}
}

func TestUnknownStack(t *testing.T) {
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
