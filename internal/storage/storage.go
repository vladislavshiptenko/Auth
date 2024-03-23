package storage

import (
	"errors"
)

var (
	ErrUserNotFound = errors.New("user not found")
	ErrUserExist    = errors.New("user exists")
	LinkNotFound    = errors.New("link not found")
)
