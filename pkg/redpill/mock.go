package redpill

import (
	"github.com/qorio/omni/auth"
)

func MockAuthContext(authed bool, context auth.Context) (bool, auth.Context) {
	return true, nil
}
