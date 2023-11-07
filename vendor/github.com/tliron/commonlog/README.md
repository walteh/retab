CommonLog
=========

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Reference](https://pkg.go.dev/badge/github.com/tliron/commonlog.svg)](https://pkg.go.dev/github.com/tliron/commonlog)
[![Go Report Card](https://goreportcard.com/badge/github.com/tliron/commonlog)](https://goreportcard.com/report/github.com/tliron/commonlog)

A common Go API for structured *and* unstructured logging with support for pluggable backends
and sinks.

Supported backends (you can log *to* these APIs):

* simple (built-in textual, colorized backend; see below)
* [klog](https://github.com/kubernetes/klog) (used by the [Kubernetes client library](https://github.com/kubernetes/client-go/))
* [systemd journal](https://www.freedesktop.org/software/systemd/man/systemd-journald.service.html)
* [zerolog](https://github.com/rs/zerolog)
* [slog](https://pkg.go.dev/log/slog) (introduced in Go 1.21)

Supported sinks (you can capture logs *from* these APIs):

* [Go built-in logging](https://pkg.go.dev/log)
* [Go built-in structured logging](https://pkg.go.dev/log/slog)
* [hclog](https://github.com/hashicorp/go-hclog) (used by many HashiCorp libraries)

Rationale
---------

The main design goal is to unite your logging APIs and allow you to change its backend at startup. For example,
you can choose to log to stderr by default or use journald when running as a systemd service. Sinks allow
you to use your selected backend with imported 3rd-party libraries when they use different logging APIs. This
design goal is inspired by [SLF4J](https://www.slf4j.org/), and we must lament that the Go ecosystem ended up
with the same logging challenges that have existed for years in the Java ecosystem.

A secondary design goal is to provide a home for a full-featured unstructured textual logging library, which we
call the ["simple" backend](simple/). It supports rich, customizable formatting, including ANSI coloring when
logging to a terminal (even on Windows). So, CommonLog is useful if you're just looking for a straightforward
logging solution right now that will not block you from using other backends in the future.

Note that efficiency and performance are *not* in themselves design goals for CommonLog, and indeed there is
always some overhead involved in wrapper and sink implementations. For example, using
[zerolog](https://github.com/rs/zerolog) *directly* involves no allocations, but using it via CommonLog will add
allocations. To put it simply: if you want zero allocation you *must* use zerolog directly. Sinks can be especially
inefficient because they may have to rely on capturing and parsing of log text. Programming is all about tradeoffs:
CommonLog provides compatibility and flexibility at the cost of some efficiency and performance. However,
as always, beware of [premature optimization](https://wiki.c2.com/?PrematureOptimization). How you are storing
or transmitting your log messages is likely the biggest factor in your optimization narrative.

A FAQ is: Why not standardize on the built-in [slog](https://pkg.go.dev/log/slog) API? Slog indeed is a big
step forward for Go, not only because it supports structured messages, but also because it decouples the
handler, an interface, from the logger. This enables alternative backends, a feature tragically missing from
Go's [log library](https://pkg.go.dev/log). Unfortunately, slog was introduced only in Go 1.21 and is thus
not used by much go Go's pre-1.21 ecosystem of 3rd-party libraries. CommonLog supports slog both as a backend
and as a sink, so you can easily mix the CommonLog API with slog API *and* handling.

Features
--------

* Fine-grained control over verbosity via hierarchical log names. For example, "engine.parser.background"
  inherits from "engine.parser", which in turn inherits from "engine". The empty name is the root of the
  hierarchy. Each name's default verbosity is that of its parent, which you can then override with
  `commonlog.SetMaxLevel()`.
* Support for call stack depth. This can be used by a backend (for example, by klog) to find out where in
  the code the logging happened.
* No need to create logger objects. The "true" API entrypoint is the global function `commonlog.NewMessage`,
  which you provide with a name and a level. The default logger type is just a convenient wrapper around it
  that provides the familiar unstructured functions.
* The unstructured `commonlog.Logger` type is an interface, allowing you to more easily switch implementations
  per use without having to introduce a whole backend. For example, you can assign the `commonlog.MOCK_LOGGER`
  to disable a logger without changing the rest of your implementation. Compare with Go's built-in
  [`Logger`](https://pkg.go.dev/log#Logger) type, which frustratingly is a struct rather than an interface.

Basic Usage
-----------

The easiest way to plug in a backend is to anonymously import the correct sub-package into your program's
main package:

```go
import (
    _ "github.com/tliron/commonlog/simple"
)
```

This should enable the backend with sensible defaults. Specifically it will log to stderr with verbosity at
the "notice" max level.

Example of structured logging:

```go
import (
    "github.com/tliron/commonlog"
    _ "github.com/tliron/commonlog/simple"
    "github.com/tliron/kutil/util"
)

func main() {
    if m := commonlog.NewErrorMessage(0, "engine", "parser"); m != nil {
        m.Set("message", "Hello world!").Set("myfloat", 10.2).Send()
    }
    util.Exit(0)
}
```

Note that `commonlog.NewMessage` will return nil if the message cannot be created, for example if
the message level is higher than the max level for that name, so you always need to check against nil.

`Set` can accept any key and value, but two special keys are recognized by the API:

* `message`: The main description of the message. This is the key used by unstructured logging.
* `scope`: An optional identifier that can be used to group messages, making them easier to filter
  (e.g. by grep on text). Backends may handle this specially. Unstructured backends may, for example,
  add it as a bracketed prefix for messages.

Also note that calling `util.Exit(0)` to exit your program is not absolutely necessary, however
it's good practice because it makes sure to flush buffered log messages for some backends.

Unstructured logging is just a trivial case of structured logging in which only the `message` key
is used. However, CommonLog provides a more familiar logging API:

```go
import (
    "github.com/tliron/commonlog"
    _ "github.com/tliron/commonlog/simple"
    "github.com/tliron/kutil/util"
)

var log = commonlog.GetLogger("engine.parser")

func main() {
    log.Errorf("Hello %s!", "world")
    util.Exit(0)
}
```

Use conditional logging to optimize for costly unstructured message creation, e.g.:

```go
if log.AllowLevel(commonlog.Debug) {
    log.Debugf("Status is: %s", getStatusFromDatabase())
}
```

The scope logger can be used to automatically set the "scope" key for another logger. It
automatically detects nesting, in which case it appends the new scope separated by a ".",
e.g.:

```go
var log = commonlog.GetLogger("engine.parser")
var validationLog = commonlog.NewScopeLogger(log, "validation")
var syntaxLog = commonlog.NewScopeLogger(validationLog, "syntax")

func main() {
    // Name is "engine.parser" and scope is "validation.syntax"
    syntaxLog.Errorf("Hello %s!", "world")
    ...
}
```

Configuration
-------------

All backends can be configured via a common API to support writing to files or stderr. For example, to increase
verbosity and write to a file:

```go
func main() {
    path := "myapp.log"
    commonlog.Configure(1, &path) // nil path would write to stdout
    ...
}
```

Backends may also have their own (non-common) configuration APIs related to their specific
features.

You can set the max level (verbosity) using either the global API or a logger. For
example, here is a way to make all logging verbose by default, except for one name:

```go
func init() {
    commonlog.SetMaxLevel(commonlog.Debug) // the root
    commonlog.SetMaxLevel(commonlog.Error, "engine", "parser")
}
```

Descendents of "engine.parser", e.g. "engine.parser.analysis", would inherit the
"engine.parser" log levels rather than the root's. Here's the same effect using unstructured
loggers:

```go
var rootLog = commonlog.GetLogger("")
var parserLog = commonlog.GetLogger("engine.parser")

func init() {
    rootLog.SetMaxLevel(commonlog.Debug)
    parserLog.SetMaxLevel(commonlog.Error)
}
```

It's important to note that the configuration APIs are not thread safe. This includes
`Configure()` and `SetMaxLevel()`. Thus, make sure to get all your configuration done before
you start sending log messages. A good place for this is `init()` or `main()` functions.

Color
-----

For the simple backend you must explicitly attempt to enable ANSI color if desired. Note that
if it's unsupported by the terminal then no ANSI codes will be sent (unless you force it via
`terminal.EnableColor(true)`):

```go
import (
    "github.com/tliron/commonlog"
    _ "github.com/tliron/commonlog/simple"
    "github.com/tliron/kutil/terminal"
    "github.com/tliron/kutil/util"
)

func main() {
    cleanup, err := terminal.EnableColor(false)
    util.FailOnError(err)
    if cleanup != nil {
        util.OnExitError(cleanup)
    }
    commonlog.GetLogger("engine.parser").Error("Hello world!") // errors are in red
    util.Exit(0)
}
```
