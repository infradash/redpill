package redpill

import (
	"encoding/json"
)

type Options struct {
	WorkingDir string
}

type DashboardEvent struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	User        string `json:"user,omitempty"`
	Type        string `json:"type,omitempty"`
	Url         string `json:"url,omitempty"`
	Timestamp   int64  `json:"timestamp,omitempty"`
}

func (this *DashboardEvent) Marshal() ([]byte, error) {
	return json.Marshal(this)
}
