package api

import (
	"errors"
	"github.com/qorio/maestro/pkg/zk"
)

var (
	ErrConflict   = zk.ErrConflict
	ErrNotFound   = errors.New("error-not-found")
	ErrCannotLock = errors.New("error-cannot-lock-for-udpates")
	ErrNoChanges  = errors.New("error-no-changes")
)
