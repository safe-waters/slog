[![ci](https://github.com/safe-waters/slog/actions/workflows/ci.yml/badge.svg)](https://github.com/safe-waters/slog/actions/workflows/ci.yml)
[![Go Report Card](https://goreportcard.com/badge/github.com/safe-waters/slog)](https://goreportcard.com/report/github.com/safe-waters/slog)
[![Go Reference](https://pkg.go.dev/badge/github.com/safe-waters/slog.svg)](https://pkg.go.dev/github.com/safe-waters/slog)

# What is this?

`slog` is a structured logger that wraps the standard library's
`log` package.

# Features

- Log each message as JSON
- Every log has metadata that includes:
  - UTC time in nano seconds
  - File name and line number
  - Level - trace, info, warn, error, panic, or fatal
- Logs can contain permanent key-value fields that log with every message
- Logs can contain key-value fields that log for just one message
- Defaults to stdout (but is configurable with any `io.Writer`)

# How to use

Without fields:

```go
package main

import "github.com/safe-waters/slog"

func main() {
	slog.Trace("hello world")
	slog.Info("hello world")
	slog.Warn("hello world")
	slog.Error("hello world")
	// slog.Panic will log and then panic
	// slog.Fatal will log and then os.Exit(1)
}

// Output:
// {"_metadata":{"file":"main.go:6","level":"trace","time":"2021-06-09T15:39:30.2649183Z"},"message":"hello world"}
// {"_metadata":{"file":"main.go:7","level":"info","time":"2021-06-09T15:39:30.2650656Z"},"message":"hello world"}
// {"_metadata":{"file":"main.go:8","level":"warn","time":"2021-06-09T15:39:30.265132Z"},"message":"hello world"}
// {"_metadata":{"file":"main.go:9","level":"error","time":"2021-06-09T15:39:30.2652018Z"},"message":"hello world"}
```

With fields:

```go
package main

import "github.com/safe-waters/slog"

func main() {
	slog.Tracef(slog.Fields{"ip": "10.0.0.1"}, "hello world")
	slog.Infof(slog.Fields{"ip": "localhost"}, "hello world")
	slog.Warnf(slog.Fields{"ip": "0.0.0.0"}, "hello world")
	slog.Errorf(slog.Fields{"ip": "10.0.0.0"}, "hello world")
	// slog.Panicf will log with fields and then panic
	// slog.Fatalf will log with fields and then os.Exit(1)
}

// Output:
// {"_metadata":{"file":"main.go:6","level":"trace","time":"2021-06-09T15:41:20.9950695Z"},"fields":{"ip":"10.0.0.1"},"message":"hello world"}
// {"_metadata":{"file":"main.go:7","level":"info","time":"2021-06-09T15:41:20.9951949Z"},"fields":{"ip":"localhost"},"message":"hello world"}
// {"_metadata":{"file":"main.go:8","level":"warn","time":"2021-06-09T15:41:20.9952392Z"},"fields":{"ip":"0.0.0.0"},"message":"hello world"}
// {"_metadata":{"file":"main.go:9","level":"error","time":"2021-06-09T15:41:20.9953044Z"},"fields":{"ip":"10.0.0.0"},"message":"hello world"}
```

With permanent fields:

```go
package main

import (
	"os"

	"github.com/safe-waters/slog"
)

func main() {
	l := slog.New(slog.DefaultCallDepth, os.Stdout, slog.Fields{"ip": "localhost"})

	l.Info("hello world")

	// Permanent fields take precedence.
	l.Warnf(slog.Fields{"ip": "shadowed", "local": "unaffected"}, "hello world")
}

// Output:
// {"_metadata":{"file":"main.go:12","level":"info","time":"2021-06-09T15:43:53.5588804Z"},"fields":{"ip":"localhost"},"message":"hello world"}
// {"_metadata":{"file":"main.go:15","level":"warn","time":"2021-06-09T15:43:53.5590044Z"},"fields":{"ip":"localhost","local":"unaffected"},"message":"hello world"}
```
