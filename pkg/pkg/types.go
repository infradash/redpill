package pkg

import (
	"github.com/qorio/omni/common"
)

type pkg struct {
}

func (this pkg) IsPkg(other interface{}) bool {
	return common.TypeMatch(this, other)
}
