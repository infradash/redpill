package api

import (
	"encoding/json"
)

func (this *Event) Marshal() ([]byte, error) {
	return json.Marshal(this)
}
