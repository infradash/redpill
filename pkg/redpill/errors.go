package redpill

import (
	"errors"
)

var (
	ErrNoInput  = errors.New("missing-input")
	ErrConflict = errors.New("revsions-conflict")
)
