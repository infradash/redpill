package orchestrate

import (
	. "github.com/infradash/redpill/pkg/api"
	"github.com/qorio/maestro/pkg/pubsub"
)

type Instance struct {
	InstanceModel   Model                  `json:"model"`
	InstanceInfo    OrchestrationInfo      `json:"info"`
	InstanceLog     pubsub.Topic           `json:"log"`
	InstanceContext map[string]interface{} `json:"context"`

	log_done chan bool
}

func (this Instance) Model() Orchestration {
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

// If the instance is currently running, the container will be publishing to
// the topic and there's nothing to do.
// If the instance is stopped and we have historical output of the instance,
// we simply read it out and send it to the same topic.  This will allow the
// client to use the same UI to replay the history of the past instance.
func (this *Instance) LogPlayback() <-chan bool {
	this.log_done = make(chan bool)

	return this.log_done
}
