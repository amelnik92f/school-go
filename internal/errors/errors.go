package errors

import (
	"errors"
	"fmt"
)

// Common application errors
var (
	ErrNotFound      = errors.New("not found")
	ErrInvalidInput  = errors.New("invalid input")
	ErrUnauthorized  = errors.New("unauthorized")
	ErrForbidden     = errors.New("forbidden")
	ErrConflict      = errors.New("conflict")
	ErrInternal      = errors.New("internal error")
	ErrDatabaseError = errors.New("database error")
)

// NotFoundError wraps a not found error with additional context
type NotFoundError struct {
	Resource string
	ID       interface{}
}

func (e *NotFoundError) Error() string {
	if e.ID != nil {
		return fmt.Sprintf("%s with id %v not found", e.Resource, e.ID)
	}
	return fmt.Sprintf("%s not found", e.Resource)
}

func (e *NotFoundError) Is(target error) bool {
	return target == ErrNotFound
}

// ValidationError wraps validation errors
type ValidationError struct {
	Field   string
	Message string
}

func (e *ValidationError) Error() string {
	return fmt.Sprintf("validation error on field '%s': %s", e.Field, e.Message)
}

func (e *ValidationError) Is(target error) bool {
	return target == ErrInvalidInput
}

// DatabaseError wraps database errors with context
type DatabaseError struct {
	Operation string
	Err       error
}

func (e *DatabaseError) Error() string {
	return fmt.Sprintf("database error during %s: %v", e.Operation, e.Err)
}

func (e *DatabaseError) Is(target error) bool {
	return target == ErrDatabaseError
}

func (e *DatabaseError) Unwrap() error {
	return e.Err
}

// NewNotFoundError creates a new NotFoundError
func NewNotFoundError(resource string, id interface{}) error {
	return &NotFoundError{Resource: resource, ID: id}
}

// NewValidationError creates a new ValidationError
func NewValidationError(field, message string) error {
	return &ValidationError{Field: field, Message: message}
}

// NewDatabaseError creates a new DatabaseError
func NewDatabaseError(operation string, err error) error {
	return &DatabaseError{Operation: operation, Err: err}
}
