package orchestrate

import (
	. "github.com/infradash/redpill/pkg/api"
	"github.com/qorio/maestro/pkg/pubsub"
	"time"
)

type ModelStorage interface {
	GetModels(domain string) ([]Model, error)
	Save(model Model) error
}

type Model struct {
	Name           string                 `json:"name"`
	FriendlyName   string                 `json:"friendly_name"`
	Description    string                 `json:"dsecription"`
	DefaultContext map[string]interface{} `json:"default_context"`
}

func (this Model) GetName() string {
	return this.Name
}

func (this Model) GetFriendlyName() string {
	return this.FriendlyName
}

func (this Model) GetDescription() string {
	return this.Description
}

func (this Model) GetDefaultContext() OrchestrationContext {
	return this.DefaultContext
}

type Instance struct {
	id        string
	log       pubsub.Topic
	startTime time.Time
	context   map[string]interface{}
}

func (this *Instance) Id() string {
	return this.id
}

func (this *Instance) StartTime() time.Time {
	return this.startTime
}

func (this *Instance) Log() *pubsub.Topic {
	return &this.log
}

func (this *Instance) Context() OrchestrationContext {
	return this.context
}
