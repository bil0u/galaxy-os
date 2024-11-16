package utils

import "fmt"

// `ManyErrors` is a slice of errors. It is used to accumulate multiple errors and return them as a single error.
type ManyErrors []error

// `Add` adds an error to the slice if it is not nil, and returns true if the error was added.
func (e ManyErrors) Add(err error) bool {
	if err != nil {
		e = append(e, err)
		return true
	}
	return false
}

// `ToError` returns a single error from the slice of errors.
func (e ManyErrors) ToError() error {
	if len(e) == 0 {
		return nil
	}
	var errStr string
	for _, err := range e {
		errStr += err.Error() + "\n"
	}
	return fmt.Errorf(errStr)
}
