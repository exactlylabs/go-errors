package errors

import (
	"fmt"
	"reflect"
	"runtime"

	"errors"
)

var (
	// Is reports whether any error in err's tree matches target.
	//
	// The tree consists of err itself, followed by the errors obtained by repeatedly
	// calling Unwrap. When err wraps multiple errors, Is examines err followed by a
	// depth-first traversal of its children.
	//
	// An error is considered to match a target if it is equal to that target or if
	// it implements a method Is(error) bool such that Is(target) returns true.
	//
	// An error type might provide an Is method so it can be treated as equivalent
	// to an existing error. For example, if MyError defines
	//
	//	func (m MyError) Is(target error) bool { return target == fs.ErrExist }
	//
	// then Is(MyError{}, fs.ErrExist) returns true. See syscall.Errno.Is for
	// an example in the standard library. An Is method should only shallowly
	// compare err and the target and not call Unwrap on either.
	Is = errors.Is

	// As finds the first error in err's tree that matches target, and if one is found, sets
	// target to that error value and returns true. Otherwise, it returns false.
	//
	// The tree consists of err itself, followed by the errors obtained by repeatedly
	// calling Unwrap. When err wraps multiple errors, As examines err followed by a
	// depth-first traversal of its children.
	//
	// An error matches target if the error's concrete value is assignable to the value
	// pointed to by target, or if the error has a method As(interface{}) bool such that
	// As(target) returns true. In the latter case, the As method is responsible for
	// setting target.
	//
	// An error type might provide an As method so it can be treated as if it were a
	// different error type.
	//
	// As panics if target is not a non-nil pointer to either a type that implements
	// error, or to any interface type.
	As = errors.As
)

type StackTrace []Frame
type Metadata map[string]interface{}

type WrappedError interface {
	error
	Unwrap() error
}

type baseError struct {
	cause      error
	msg        string
	typeStr    string
	stacktrace StackTrace
	metadata   Metadata
}

func (e *baseError) Error() string {
	contextMsg := ""
	if e.typeStr != "" {
		contextMsg = e.typeStr
	}
	if len(e.stacktrace) > 0 {
		topStack := e.stacktrace[0]
		if contextMsg != "" {
			contextMsg = fmt.Sprintf("%s@%s.%s", e.typeStr, topStack.Package(), topStack.FuncName())
		} else {
			contextMsg = fmt.Sprintf("%s.%s", topStack.Package(), topStack.FuncName())
		}
	}

	msg := ""
	if contextMsg != "" {
		msg = contextMsg
	}
	if e.msg != "" {
		msg = fmt.Sprintf("[%s] %s", msg, e.msg)
	}
	if e.cause != nil {
		msg = fmt.Sprintf("%s => %s", msg, e.cause.Error())
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

// WithMetadata adds metadata to the error. This is a shallow update of the current metadata field,
// so nested fields won't be updated.
func (e *baseError) WithMetadata(meta Metadata) *baseError {
	if e.metadata == nil {
		e.metadata = meta
		return e
	}
	for k, v := range meta {
		e.metadata[k] = v
	}
	return e
}

// New creates a new error with stacktrace
func New(msg string, args ...any) *baseError {
	msg = fmt.Sprintf(msg, args...)
	return &baseError{
		nil, msg, "", getStack(0), Metadata{},
	}
}

// W wraps an error in a new error with stacktrace and propagating the metadata.
func W(err error) *baseError {
	var baseErr *baseError
	if As(err, &baseErr) {
		// propagate the typeStr up if there's no new typeStr provided
		return &baseError{err, "", baseErr.typeStr, getStack(0), baseErr.metadata}
	}
	return &baseError{err, "", "", getStack(0), Metadata{}}
}

// Wrap wraps the given error in a new Error with the given message, having a stacktrace and propagating metadata.
func Wrap(err error, msg string, args ...any) *baseError {
	msg = fmt.Sprintf(msg, args...)

	var baseErr *baseError
	if As(err, &baseErr) {
		// propagate the typeStr up if there's no new typeStr provided
		return &baseError{err, msg, baseErr.typeStr, getStack(0), baseErr.metadata}
	}
	return &baseError{err, msg, "", getStack(0), Metadata{}}
}

// NewWithType creates a new error with stacktrace and a custom type string returned by its Type() method
func NewWithType(msg, typeStr string, args ...any) *baseError {
	msg = fmt.Sprintf(msg, args...)
	return &baseError{nil, msg, typeStr, getStack(0), Metadata{}}
}

// Wrap wraps the given error in a new error with stack trace and a custom type string returned by its Type() method
func WrapWithType(err error, msg, typeStr string, args ...any) *baseError {
	msg = fmt.Sprintf(msg, args...)
	metaPtr := GetMetadata(err)
	meta := Metadata{}
	if metaPtr != nil {
		meta = *metaPtr
	}
	return &baseError{err, msg, typeStr, getStack(0), meta}
}

func NewSentinel(typeStr, msg string) *baseError {
	return &baseError{nil, msg, typeStr, nil, Metadata{}}
}

// SentinelWithStack wraps the given sentinel error and adds a stacktrace
// If Sentinel error is a baseError, it creates a copy with the new stacktrace
func SentinelWithStack(err error) *baseError {
	var baseErr *baseError
	if As(err, &baseErr) {
		baseErr.stacktrace = nil // Remove the stacktrace pointing to where the sentinel gets created, as this is useless
		return &baseError{baseErr, baseErr.msg, baseErr.typeStr, getStack(0), baseErr.metadata}
	}
	return &baseError{err, "", "", getStack(0), Metadata{}}
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

func GetMetadata(err error) *Metadata {
	var baseErr *baseError
	if As(err, &baseErr) {
		return &baseErr.metadata
	}
	return nil
}
