package api

import (
	"errors"
)

var (
	ErrConflict = errors.New("revisions-conflict")
	ErrNotFound = errors.New("not-found")
)
