# What is this?
`slog` is an opinionated, structured logger that wraps the standard library's
`log` package.

# Features
* Log each message as JSON
* Every log has metadata that includes:
    * UTC time in unix nano seconds
    * File name and line number
    * Level - info, warn, or error
* Logs can contain permanent key-value fields that log with every message
* Logs can contain key-value fields that log for just one message
* Defaults to stdout (but is configurable with any `io.Writer`)

# How to use
Without fields:
```go
package main

import "github.com/safe-waters/slog"

func main() {
	slog.Info("hello world")
	slog.Warn("hello world")
	slog.Error("hello world")
}

// Output:
// {"_metadata":{"file":"main.go:6","level":"info","time":"1622178871606977200"},"message":"hello world"}
// {"_metadata":{"file":"main.go:7","level":"warn","time":"1622178871607224000"},"message":"hello world"}
// {"_metadata":{"file":"main.go:8","level":"error","time":"1622178871607270900"},"message":"hello world"}
```

With fields:
```go
package main

import "github.com/safe-waters/slog"

func main() {
	slog.Infof(slog.Fields{"ip": "localhost"}, "hello world")
	slog.Warnf(slog.Fields{"ip": "0.0.0.0"}, "hello world")
	slog.Errorf(slog.Fields{"ip": "10.0.0.0"}, "hello world")
}

// Output:
// {"_metadata":{"file":"main.go:6","level":"info","time":"1622178920099374600"},"fields":{"ip":"localhost"},"message":"hello world"}
// {"_metadata":{"file":"main.go:7","level":"warn","time":"1622178920099502700"},"fields":{"ip":"0.0.0.0"},"message":"hello world"}
// {"_metadata":{"file":"main.go:8","level":"error","time":"1622178920099581500"},"fields":{"ip":"10.0.0.0"},"message":"hello world"}
```

With permanent fields:
```go
package main

import (
	"os"

	"github.com/safe-waters/slog"
)

func main() {
	// See documentation for DefaultSkip. It determines the stack
	// frame for reporting the file name and line number. If you wrap
	// slog in your own logger, you will likely initialize the
	// logger with slog.DefaultSkip+1.
	l := slog.New(slog.DefaultSkip, os.Stdout, slog.Fields{"ip": "localhost"})

	l.Info("hello world")

	// Permanent fields take precedence.
	l.Warnf(slog.Fields{"ip": "shadowed", "local": "unaffected"}, "hello world")
}

// Output:
// {"_metadata":{"file":"main.go:16","level":"info","time":"1622178968714522200"},"fields":{"ip":"localhost"},"message":"hello world"}
// {"_metadata":{"file":"main.go:19","level":"warn","time":"1622178968714619900"},"fields":{"ip":"localhost","local":"unaffected"},"message":"hello world"}
```
