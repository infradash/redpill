package redpill

import (
	"encoding/json"
)

type Options struct {
	WorkingDir string
}

func (this *Event) Marshal() ([]byte, error) {
	return json.Marshal(this)
}
