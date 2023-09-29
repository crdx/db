package db

var handleError func(err error)

// SetErrorHandler sets a function to be called if any database operations fail. Set to nil to use
// the default (panic).
func SetErrorHandler(f func(err error)) {
	handleError = f
}

// must0 is like lo.Must0 but calls the defined error handler if there is one.
func must0(err error) {
	if err != nil {
		if handleError == nil {
			panic(err)
		}

		handleError(err)
	}
}
