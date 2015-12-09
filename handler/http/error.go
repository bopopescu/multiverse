package http

import (
	"errors"
	"fmt"
)

// Errors used for protocol control flow.
var (
	ErrBadRequest    = errors.New("bad request")
	ErrLimitExceeded = errors.New("limit")
	ErrUnauthorized  = errors.New("unauthorized")
)

// Error is used to carry additional error informaiton reported back to clients.
type Error struct {
	Err     error
	Message string
}

func wrapError(err error, msg string) *Error {
	return &Error{
		Err:     err,
		Message: fmt.Sprintf("%s: %s", err.Error(), msg),
	}
}

func (e *Error) Error() string {
	return e.Message
}

func unwrapError(err error) error {
	switch e := err.(type) {
	case *Error:
		return e.Err
	}

	return err
}