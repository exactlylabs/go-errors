package main

import (
	"flag"
	"fmt"

	"github.com/exactlylabs/go-errors/cmd/example/internal/mylibrary"
	"github.com/exactlylabs/go-errors/pkg/errors"
	"github.com/exactlylabs/go-monitor/pkg/sentry"
)

// Objectives:

// Create an error with a stack trace that you can configure the name of the error when sentry reports it.

// Create a structure where you can have a "BaseError" for your library/application, so you can distinct between errors that are "internal" and errors that are "external".

func main() {
	dsn := flag.String("sentry-dsn", "", "Sentry DSN")
	flag.Parse()
	sentry.Setup(*dsn, "test", "test", "test")
	defer sentry.NotifyIfPanic()

	err := TopLevelFunc()
	fmt.Println(err)
	fmt.Printf("Is SomeLib error: %v\n", errors.Is(err, mylibrary.ErrLibraryBaseError))
	fmt.Printf("Is ErrSomething error: %v\n", errors.Is(err, mylibrary.ErrInvalidError))
	panic(err)
}

func TopLevelFunc() (err error) {
	err = mylibrary.DoLibraryStuff()
	// Here, we have an error and want now to wrap it into our own sentinel AppError
	// One possible approach is to:
	return errors.W(err)
}
