package api

import (
	"errors"
)

var (
	ErrConflict = errors.New("revsions-conflict")
	ErrNotFound = errors.New("not-found")
)
