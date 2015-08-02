package api

import (
	"github.com/qorio/maestro/pkg/pubsub"
	"time"
)

type OrchestrationContext map[string]interface{}

type Orchestration interface {
	GetName() string
	GetFriendlyName() string
	GetDescription() string
	GetDefaultContext() OrchestrationContext
}

type OrchestrationInfo struct {
	Domain    string
	Id        string
	Name      string
	StartTime time.Time
	Status    string
	User      string
}

type OrchestrationInstance interface {
	Info() OrchestrationInfo
	Log() *pubsub.Topic
	Context() OrchestrationContext
}
