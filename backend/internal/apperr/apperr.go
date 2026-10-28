// Package apperr defines the application error model used to produce a
// consistent JSON error envelope across every endpoint.
package apperr

import (
	"errors"
	"fmt"
	"net/http"
)

// Code is a stable, machine readable error identifier.
type Code string

const (
	CodeInternal       Code = "internal_error"
	CodeValidation     Code = "validation_error"
	CodeNotFound       Code = "not_found"
	CodeUnauthorized   Code = "unauthorized"
	CodeForbidden      Code = "forbidden"
	CodeConflict       Code = "conflict"
	CodeTooMany        Code = "too_many_requests"
	CodeBadRequest     Code = "bad_request"
	CodeUnprocessable  Code = "unprocessable_entity"
	CodeNotImplemented Code = "not_implemented"
)

// FieldError describes a single invalid input field.
type FieldError struct {
	Field   string `json:"field"`
	Message string `json:"message"`
}

// Error is the canonical application error.
type Error struct {
	Code    Code         `json:"code"`
	Message string       `json:"message"`
	Fields  []FieldError `json:"fields,omitempty"`
	status  int
	cause   error
}

// Error implements the error interface.
func (e *Error) Error() string {
	if e.cause != nil {
		return fmt.Sprintf("%s: %s: %v", e.Code, e.Message, e.cause)
	}
	return fmt.Sprintf("%s: %s", e.Code, e.Message)
}

// Unwrap exposes the wrapped cause to errors.Is/As.
func (e *Error) Unwrap() error { return e.cause }

// Status returns the HTTP status code associated with the error.
func (e *Error) Status() int {
	if e.status != 0 {
		return e.status
	}
	return statusForCode(e.Code)
}

// WithCause attaches an internal cause. The cause is never serialized to
// clients, only logged.
func (e *Error) WithCause(err error) *Error {
	e.cause = err
	return e
}

// WithFields attaches validation field errors.
func (e *Error) WithFields(fields ...FieldError) *Error {
	e.Fields = append(e.Fields, fields...)
	return e
}

func statusForCode(code Code) int {
	switch code {
	case CodeValidation, CodeBadRequest:
		return http.StatusBadRequest
	case CodeUnauthorized:
		return http.StatusUnauthorized
	case CodeForbidden:
		return http.StatusForbidden
	case CodeNotFound:
		return http.StatusNotFound
	case CodeConflict:
		return http.StatusConflict
	case CodeUnprocessable:
		return http.StatusUnprocessableEntity
	case CodeTooMany:
		return http.StatusTooManyRequests
	case CodeNotImplemented:
		return http.StatusNotImplemented
	default:
		return http.StatusInternalServerError
	}
}

// New creates an error with an explicit code and message.
func New(code Code, message string) *Error {
	return &Error{Code: code, Message: message}
}

// Newf creates an error with a formatted message.
func Newf(code Code, format string, args ...any) *Error {
	return &Error{Code: code, Message: fmt.Sprintf(format, args...)}
}

// Validation creates a validation error from one or more field errors.
func Validation(fields ...FieldError) *Error {
	return &Error{Code: CodeValidation, Message: "the request payload is invalid", Fields: fields}
}

// InvalidField is a convenience helper for a single field error.
func InvalidField(field, message string) *Error {
	return Validation(FieldError{Field: field, Message: message})
}

// NotFound creates a not-found error for a resource.
func NotFound(resource string) *Error {
	return Newf(CodeNotFound, "%s not found", resource)
}

// Forbidden creates an authorization error.
func Forbidden(message string) *Error {
	if message == "" {
		message = "you do not have permission to perform this action"
	}
	return New(CodeForbidden, message)
}

// Unauthorized creates an authentication error.
func Unauthorized(message string) *Error {
	if message == "" {
		message = "authentication required"
	}
	return New(CodeUnauthorized, message)
}

// Conflict creates a conflict error.
func Conflict(message string) *Error { return New(CodeConflict, message) }

// Internal wraps an unexpected error, hiding the internal cause from clients.
func Internal(err error) *Error {
	return (&Error{Code: CodeInternal, Message: "an unexpected error occurred"}).WithCause(err)
}

// As extracts an *Error from err, returning a generic internal error when the
// error is not part of the application error model.
func As(err error) *Error {
	if err == nil {
		return nil
	}
	var appErr *Error
	if errors.As(err, &appErr) {
		return appErr
	}
	return Internal(err)
}
