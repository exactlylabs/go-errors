package errors

import (
	"fmt"
	"reflect"
	"runtime"

	"errors"
)

// Export the usual errors functions, so the caller doesn't have to import both
var (
	Is = errors.Is
	As = errors.As
)

type StackTrace []Frame

type baseError struct {
	cause      error
	msg        string
	typeStr    string
	stacktrace StackTrace
}

func (e *baseError) Error() string {
	topStack := e.stacktrace[0]
	msg := fmt.Sprintf("%s.%s %s", topStack.Package(), topStack.FuncName(), e.msg)
	if e.typeStr != "" {
		msg = fmt.Sprintf("[%s] %s", e.typeStr, msg)
	}
	if e.cause != nil {
		msg = fmt.Sprintf("%s: %s", msg, e.cause.Error())
	}
	return msg
}

func (e *baseError) Unwrap() error {
	return e.cause
}

func (e *baseError) StackTrace() StackTrace {
	return e.stacktrace
}

// Type returns either the struct name or the custom type string from typeStr attribute
func (e *baseError) Type() string {
	if e.typeStr != "" {
		return e.typeStr
	}
	return reflect.TypeOf(e).String()
}

// New creates a new error with stacktrace
func New(msg string, args ...any) error {
	msg = fmt.Sprintf(msg, args...)
	return &baseError{nil, msg, "", getStack(0)}
}

// W wraps an error in a new error with stacktrace
func W(err error) error {
	if baseErr, ok := err.(*baseError); ok {
		// propagate the typeStr up if there's no new typeStr provided
		return &baseError{err, "", baseErr.typeStr, getStack(0)}
	}
	return &baseError{err, "", "", getStack(0)}
}

// Wrap wraps the given error in a new Error with the given message, having a stacktrace
func Wrap(err error, msg string, args ...any) error {
	msg = fmt.Sprintf(msg, args...)
	if baseErr, ok := err.(*baseError); ok {
		// propagate the typeStr up if there's no new typeStr provided
		return &baseError{err, msg, baseErr.typeStr, getStack(0)}
	}
	return &baseError{err, msg, "", getStack(0)}
}

// NewWithType creates a new error with stacktrace and a custom type string returned by its Type() method
func NewWithType(msg, typeStr string, args ...any) error {
	msg = fmt.Sprintf(msg, args...)
	return &baseError{nil, msg, typeStr, getStack(0)}
}

// Wrap wraps the given error in a new error with stack trace and a custom type string returned by its Type() method
func WrapWithType(err error, msg, typeStr string, args ...any) error {
	msg = fmt.Sprintf(msg, args...)
	return &baseError{err, msg, typeStr, getStack(0)}
}

// SentinelWithStack wraps the given sentinel error and adds a stacktrace
// If Sentinel error is a baseError, it creates a copy with the new stacktrace
func SentinelWithStack(err error) error {
	if err, ok := err.(*baseError); ok {
		err.stacktrace = nil // Remove the stacktrace pointing to where the sentinel gets created, as this is useless
		return &baseError{err.cause, err.msg, err.typeStr, getStack(0)}
	}
	return &baseError{err, "", "", getStack(0)}
}

func getStack(skip int) StackTrace {
	var pcs = make([]uintptr, 32)
	n := runtime.Callers(3, pcs[:])
	pcs = pcs[0:n] // Callers doc states that we should not use the reference
	frames := runtime.CallersFrames(pcs)
	stack := make(StackTrace, 0, 32)
	for {
		frame, more := frames.Next()
		if !more {
			break
		}
		stack = append(stack, Frame{frame})
	}
	if skip <= len(stack) {
		stack = stack[skip:]
	}
	return stack
}
