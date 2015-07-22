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

type OrchestrationInstance interface {
	Id() string
	Log() *pubsub.Topic
	StartTime() time.Time
	Context() OrchestrationContext
}
