package apperror

import (
	"errors"
	"fmt"
	"runtime"
	"sync/atomic"
	"unicode/utf8"
)

// MaxMessageLength is the hard limit for messages carried by errors.
const MaxMessageLength = 120

var (
	traceEnabled      atomic.Bool
	stacktraceEnabled atomic.Bool
)

// Configure toggles tracing behaviour. Call once during application startup.
func Configure(trace, stacktrace bool) {
	traceEnabled.Store(trace)
	stacktraceEnabled.Store(stacktrace)
}

// Error is the only error type that should cross layer boundaries.
type Error struct {
	Code    Code
	Message string
	trace   string
	cause   error
}

// API builds a client-facing error with a reuse-checked code.
func API(code Code, message string) *Error {
	return newError(code, message, nil)
}

// Runtime builds a server-side error wrapping the underlying cause.
func Runtime(code Code, cause error) *Error {
	message := InfoOf(code).PublicMessage
	if cause != nil && code != CodeInternal && code != CodePanic {
		message = InfoOf(code).PublicMessage
	}
	return newError(code, message, cause)
}

// Wrap annotates an existing error with a code.
func Wrap(cause error, code Code) *Error {
	if cause == nil {
		return newError(code, InfoOf(code).PublicMessage, nil)
	}
	return newError(code, cause.Error(), cause)
}

func newError(code Code, message string, cause error) *Error {
	return &Error{
		Code:    code,
		Message: truncate(message, MaxMessageLength),
		trace:   captureTrace(),
		cause:   cause,
	}
}

// Error implements the error interface.
func (e *Error) Error() string {
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap returns the underlying cause for errors.Is/As.
func (e *Error) Unwrap() error { return e.cause }

// Trace returns the capture point when tracing is enabled.
func (e *Error) Trace() string { return e.trace }

// StackEnabled reports whether full stack traces should be attached to logs.
func StackEnabled() bool { return stacktraceEnabled.Load() }

// Kind returns the error zone.
func (e *Error) Kind() Kind { return InfoOf(e.Code).Kind }

// HTTPStatus returns the status constant bound to the code.
func (e *Error) HTTPStatus() int { return InfoOf(e.Code).HTTPStatus }

// PublicMessage returns the message safe to expose to clients.
func (e *Error) PublicMessage() string {
	info := InfoOf(e.Code)
	if info.Kind == KindRuntime && info.PublicMessage != "" {
		return info.PublicMessage
	}
	if e.Message == "" {
		return info.PublicMessage
	}
	return e.Message
}

// Is allows errors.Is checks against both the concrete error and its code.
func (e *Error) Is(target error) bool {
	var other *Error
	if errors.As(target, &other) {
		return e.Code == other.Code
	}
	return false
}

func truncate(s string, limit int) string {
	if utf8.RuneCountInString(s) <= limit {
		return s
	}
	runes := []rune(s)
	return string(runes[:limit])
}

func captureTrace() string {
	if !traceEnabled.Load() {
		return ""
	}
	pc, file, line, ok := runtime.Caller(3)
	if !ok {
		return ""
	}
	name := "unknown"
	if fn := runtime.FuncForPC(pc); fn != nil {
		name = fn.Name()
	}
	return fmt.Sprintf("%s (%s:%d)", name, file, line)
}
