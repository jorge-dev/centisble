package mocks

import "errors"

var (
	ErrRecordNotFound = errors.New("record not found")
	ErrInvalidInput   = errors.New("invalid input")
	ErrDuplicateKey   = errors.New("duplicate key value violates unique constraint")
)
