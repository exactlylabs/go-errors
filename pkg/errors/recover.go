package errors

import (
	"fmt"
)

func RecoverPanic(r interface{}, errPtr *error) {
	var err error
	if r != nil {
		if panicErr, ok := r.(error); ok {
			if baseErr, ok := panicErr.(*baseError); ok {
				err = baseErr
			} else {
				baseErr := Wrap(panicErr, "caught panic")
				baseErr.stacktrace = baseErr.stacktrace[2:]
				err = baseErr
			}
		} else {
			baseErr := New(fmt.Sprintf("caught panic: %v", r))
			baseErr.stacktrace = baseErr.stacktrace[2:]
			err = baseErr
		}
	}

	if err != nil {
		// Pop twice: once for the errors package, then again for the defer function we must
		// run this under. We want the stacktrace to originate at the source of the panic, not
		// in the infrastructure that catches it.
		*errPtr = err
	}
}
