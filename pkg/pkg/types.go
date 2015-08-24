package pkg

import (
	"github.com/qorio/omni/common"
)

type pkg struct {
	Docker
}

func (this pkg) IsPkg(other interface{}) bool {
	return common.TypeMatch(this, other)
}
