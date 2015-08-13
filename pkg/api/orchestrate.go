package api

import (
	"github.com/qorio/maestro/pkg/pubsub"
	"time"
)

type OrchestrationModel interface {
	IsOrchestrationModel(interface{}) bool
}

type OrchestrationList []OrchestrationDescription
type OrchestrationDescription struct {
	Name         string      `json:"name,omitempty"`
	Label        string      `json:"label,omitempty"`
	Description  string      `json:"description,omitempty"`
	ActivateUrl  string      `json:"activate_url"`
	DefaultInput interface{} `json:"default_input,omitempty"`
}

type OrchestrationContext map[string]interface{}

type Orchestration interface {
	GetName() string
	GetFriendlyName() string
	GetDescription() string
	GetDefaultContext() OrchestrationContext
}

type OrchestrationInfo struct {
	Domain         string    `json:"domain"`
	Id             string    `json:"id"`
	Name           string    `json:"name"`
	StartTime      time.Time `json:"start_time"`
	CompletionTime time.Time `json:"completion_time"`
	Status         string    `json:"status"`
	User           string    `json:"user"`
	Note           string    `json:"note"`
}

type OrchestrationInstance interface {
	Model() Orchestration
	Info() OrchestrationInfo
	Log() *pubsub.Topic
	Context() OrchestrationContext
}
