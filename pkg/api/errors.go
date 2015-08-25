package api

import (
	"errors"
)

var (
	ErrConflict   = errors.New("revisions-conflict")
	ErrNotFound   = errors.New("not-found")
	ErrCannotLock = errors.New("error-cannot-lock-for-udpates")
	ErrNoChanges  = errors.New("error-no-changes")
)
