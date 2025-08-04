package services

import (
	"errors"
	"fmt"
)

// ErrResourceNotFound is returned when a requested resource is not found
type ErrResourceNotFound struct {
	error
}

func NewErrResourceNotFound(err error) *ErrResourceNotFound {
	return &ErrResourceNotFound{err}
}

func NewErrAlbumNotFound(id string) *ErrResourceNotFound {
	return &ErrResourceNotFound{fmt.Errorf("album %s not found", id)}
}

func NewErrMediaNotFound(id string) *ErrResourceNotFound {
	return &ErrResourceNotFound{fmt.Errorf("media %s not found", id)}
}

// ErrResourceExistsAlready is returned when trying to create a resource that already exists
type ErrResourceExistsAlready struct {
	error
}

func NewErrAlbumExistsAlready(id string) *ErrResourceExistsAlready {
	return &ErrResourceExistsAlready{fmt.Errorf("album %s already exists", id)}
}

func NewErrMediaExistsAlready(id string) *ErrResourceExistsAlready {
	return &ErrResourceExistsAlready{fmt.Errorf("media %s already exists", id)}
}

// IsErrResourceNotFound checks if an error is a resource not found error
func IsErrResourceNotFound(err error) bool {
	_, ok := err.(*ErrResourceNotFound)
	return ok
}

// IsErrResourceExistsAlready checks if an error is a resource exists already error
func IsErrResourceExistsAlready(err error) bool {
	_, ok := err.(*ErrResourceExistsAlready)
	return ok
}

type ErrUpdateAlbum struct {
	error
}

func NewErrUpdateAlbum(reason string) *ErrUpdateAlbum {
	return &ErrUpdateAlbum{errors.New(reason)}
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
