package orchestrate

import (
	dash "github.com/infradash/dash/pkg/executor"
	. "github.com/infradash/redpill/pkg/api"
	"github.com/qorio/maestro/pkg/docker"
	"github.com/qorio/maestro/pkg/pubsub"
)

type ModelStorage interface {
	GetModels(domain string) ([]Model, error)
	Get(domain, name string) (Model, error)
	Save(domain string, model Model) error
}

type InstanceStorage interface {
	Save(instance *Instance) error
	Get(id string) (*Instance, error)
	List(domain, orchestration string) ([]Instance, error)
}

type Model struct {
	dash.ExecutorConfig

	FriendlyName   string                 `json:"friendly_name"`
	Description    string                 `json:"dsecription"`
	DefaultContext map[string]interface{} `json:"default_context"`

	// Different way of running this - docker, exec (with dash), or some scheduler api call.
	Docker *docker.ContainerControl `json:"docker,omitempty"`
}

func (this Model) GetName() string {
	return string(this.Name)
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
	InstanceModel   Model                  `json:"model"`
	InstanceInfo    OrchestrationInfo      `json:"info"`
	InstanceLog     pubsub.Topic           `json:"log"`
	InstanceContext map[string]interface{} `json:"context"`
}

func (this Instance) Model() Model {
	return this.InstanceModel
}

func (this Instance) Info() OrchestrationInfo {
	return this.InstanceInfo
}

func (this Instance) Log() *pubsub.Topic {
	return &this.InstanceLog
}

func (this Instance) Context() OrchestrationContext {
	return this.InstanceContext
}
