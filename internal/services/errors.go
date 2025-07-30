package services

import "fmt"

// ErrResourceNotFound is returned when a requested resource is not found
type ErrResourceNotFound struct {
	Resource string
	ID       string
}

func (e *ErrResourceNotFound) Error() string {
	return fmt.Sprintf("%s with ID '%s' not found", e.Resource, e.ID)
}

// ErrInvalidInput is returned when input validation fails
type ErrInvalidInput struct {
	Field   string
	Value   interface{}
	Message string
}

func (e *ErrInvalidInput) Error() string {
	if e.Message != "" {
		return fmt.Sprintf("invalid input for field '%s': %s", e.Field, e.Message)
	}
	return fmt.Sprintf("invalid input for field '%s': %v", e.Field, e.Value)
}

// ErrConflict is returned when a resource conflict occurs
type ErrConflict struct {
	Resource string
	Message  string
}

func (e *ErrConflict) Error() string {
	return fmt.Sprintf("conflict with %s: %s", e.Resource, e.Message)
}
