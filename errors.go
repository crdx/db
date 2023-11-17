package db

import (
	"errors"
)

var handleError func(err error)

// SetErrorHandler sets a function to be called if any database operations fail. Set to nil to use
// the default (panic).
func SetErrorHandler(f func(err error)) {
	handleError = f
}

// must0 is like lo.Must0 but calls the defined error handler if there is one.
func must0(err any) {
	if err == nil {
		return
	}

	var handle func(error)
	if handleError != nil {
		handle = handleError
	} else {
		handle = func(err error) {
			panic(err)
		}
	}

	switch e := err.(type) {
	case bool:
		if !e {
			handle(errors.New("not ok"))
		}

	case error:
		handle(e)

	default:
		panic("invalid err")
	}
}

func must[T any](value T, err any) T {
	must0(err)
	return value
}
