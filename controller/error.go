package controller

import (
	"errors"
	"fmt"
)

const errFmt = "%s: %s"

// Common errors
var (
	ErrInvalidEntity = errors.New("invalid entity")
	ErrNotFound      = errors.New("resource not found")
)

type Error struct {
	err error
	msg string
}

func (e *Error) Error() string {
	return e.msg
}

// IsInvalidEntity indciates if err is ErrInvalidEntity.
func IsInvalidEntity(err error) bool {
	return unwrapError(err) == ErrInvalidEntity
}

func unwrapError(err error) error {
	switch e := err.(type) {
	case *Error:
		return e.err
	}

	return err
}

func wrapError(err error, format string, args ...interface{}) error {
	return &Error{
		err: err,
		msg: fmt.Sprintf(
			errFmt,
			err.Error(),
			fmt.Sprintf(format, args...),
		),
	}
}
