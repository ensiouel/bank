package apperror

import (
	"errors"
	"fmt"
)

var lastCode Code

type Code uint64

type Error struct {
	Code    Code   `json:"code"`
	Status  string `json:"status"`
	Message string `json:"message,omitempty"`
	Scope   string `json:"-"`
	Err     error  `json:"-"`
}

func New(status string) Error {
	lastCode++

	return Error{
		Code:    lastCode,
		Status:  status,
		Message: "",
	}
}

func NewWithCode(code int, status string) Error {
	code++
	return Error{
		Code:    Code(code),
		Status:  status,
		Message: "",
	}
}

func Is(target error, apperr error) (err Error, ok bool) {
	if errors.Is(target, apperr) {
		return target.(Error), true
	}

	return
}

func (error Error) Error() string {
	if error.Err != nil {
		return fmt.Sprintf("error: %s: %s: %s", error.Status, error.Scope, error.Err.Error())
	}

	return fmt.Sprintf("error: %s: %s: %s", error.Status, error.Scope, error.Message)
}

func (error Error) Is(target error) bool {
	if target == nil {
		return false
	}

	err, ok := target.(Error)
	if !ok {
		return false
	}

	return error.Code == err.Code
}

func (error Error) WithMessage(message string) Error {
	error.Message = message

	return error
}

func (error Error) WithError(err error) Error {
	error.Err = err

	return error
}

func (error Error) WithScope(scope string) Error {
	error.Scope = scope

	return error
}
