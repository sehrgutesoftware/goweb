package goweb

import (
	"errors"
	"fmt"
	"net/http"
)

var (
	// ErrGeneric indicates that an unspecified error occurs.
	ErrGeneric = NewError("generic", "generic error", http.StatusInternalServerError)
)

// ErrorCoder defines the interface of an error that has a unique code.
//
// It is used in [RespondError] to determine the error code and message.
type ErrorCoder interface {
	error
	ErrorCode() string // ErrorCode returns the error code
}

// APIError is an error that can be returned to the client.
//
// It is used in [RespondError] to determine the HTTP status code and optional
// payload of the response.
type APIError interface {
	ErrorCoder
	StatusCode() int  // StatusCode returns the HTTP status code
	ErrorDetail() any // ErrorDetail returns optional related data
}

// codeError is an error with a unique code, an HTTP status and optional data.
type codeError struct {
	// err is the underlying error that caused this error, if any.
	// It also determines the error message.
	err error
	// code is a unique error code that can be used to identify the error.
	code string
	// statusCode is the HTTP status code associated with the error.
	statusCode int
	// data is optional data associated with the error.
	data any
}

// Error returns the error message.
func (e *codeError) Error() string {
	return e.err.Error()
}

// ErrorCode returns the error code identifying the error.
func (e *codeError) ErrorCode() string {
	return e.code
}

// StatusCode returns the HTTP status code associated with the error.
func (e *codeError) StatusCode() int {
	return e.statusCode
}

// ErrorDetail returns any optional data associated with the error.
func (e *codeError) ErrorDetail() any {
	return e.data
}

// NewError creates a new [codeError] with given code, message and HTTP status.
func NewError(code, message string, status int) *codeError {
	return &codeError{
		err:        errors.New(message),
		code:       code,
		statusCode: status,
	}
}

// Wrap wraps the given error in a new [codeError].
func (e *codeError) Wrap(err error) *codeError {
	return &codeError{
		err:        fmt.Errorf("%w: %w", e.err, err),
		code:       e.code,
		statusCode: e.statusCode,
		data:       e.data,
	}
}

// Apply returns a copy of the error with the given data.
func (e *codeError) Apply(data any) *codeError {
	return &codeError{
		err:        e.err,
		code:       e.code,
		statusCode: e.statusCode,
		data:       data,
	}
}

// Is reports whether the error is the same as the target error.
func (e *codeError) Is(target error) bool {
	if e == target {
		return true
	}

	if e == nil || target == nil {
		return false
	}

	t, ok := target.(*codeError)
	if !ok {
		return false
	}

	return e.code == t.code
}

// Unwrap returns the underlying error.
func (e *codeError) Unwrap() error {
	return e.err
}
