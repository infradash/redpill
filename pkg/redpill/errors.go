package redpill

import (
	"errors"
)

var (
	ErrNoInput    = errors.New("missing-input")
	ErrNoESClient = errors.New("no-es-client")
)
