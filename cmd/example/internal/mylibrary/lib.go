package mylibrary

import (
	"github.com/exactlylabs/go-errors/pkg/errors"
)

// ErrLibraryBaseError is our base error for this library
var ErrLibraryBaseError = errors.NewWithType("", "LibraryBaseError")

// ErrInvalidError is a specific error from this Library, it wraps ErrLibraryBaseError
// It also adds a new Type string, to be used when reporting to Sentry using github.com/exactlylabs/go-monitor/pkg/sentry package
var ErrInvalidError = errors.WrapWithType(ErrLibraryBaseError, "something went wrong", "ErrSomething")

func DoLibraryStuff() error {
	// Wrap Sentinel error ErrInvalidError to include the correct stacktrace, otherwise it would miss this function
	return errors.SentinelWithStack(ErrInvalidError)
}
