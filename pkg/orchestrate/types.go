package orchestrate

import (
	dash "github.com/infradash/dash/pkg/executor"
	. "github.com/infradash/redpill/pkg/api"
	"github.com/qorio/maestro/pkg/docker"
)

type ModelStorage interface {
	GetModels(domainClass string) ([]Model, error)
	Get(domainClass, name string) (*Model, error)
	Save(domainClass string, model *Model) error
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

func (this Model) IsOrchestrationModel(ptr interface{}) bool {
	_, isa := ptr.(*Model)
	return isa
}
