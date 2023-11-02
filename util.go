package db

// chain runs the provided functions until it reaches one that returns a non-nil error, then returns
// it. Returns nil if none of the functions errored.
func chain(fs ...func() error) error {
	for _, f := range fs {
		if err := f(); err != nil {
			return err
		}
	}

	return nil
}
