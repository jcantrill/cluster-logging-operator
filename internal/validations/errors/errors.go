package errors

import (
	"fmt"
	"reflect"
)

type ValidationError struct {
	msg string
}

func (e *ValidationError) Error() string {
	return e.msg
}

func NewValidationError(msg string, args ...string) *ValidationError {
	return &ValidationError{msg: fmt.Sprintf(msg, args)}
}

func IsValidationError(err error) bool {
	neededType := reflect.TypeOf(&ValidationError{})
	if reflect.TypeOf(err).AssignableTo(neededType) {
		return true
	}
	return false
}
