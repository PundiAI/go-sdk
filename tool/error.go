package tool

import "github.com/pkg/errors"

var _ error = (*IgnoreException)(nil)

type IgnoreException struct {
	err error
}

func NewIgnoreException(err error) IgnoreException {
	return IgnoreException{err: err}
}

func (e IgnoreException) Error() string {
	return e.err.Error()
}

func IsIgnoreException(err error) bool {
	causeErr := errors.Cause(err)
	return errors.As(causeErr, &IgnoreException{})
}
