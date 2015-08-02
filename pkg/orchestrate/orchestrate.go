package orchestrate

import (
	"github.com/golang/glog"
	. "github.com/infradash/redpill/pkg/api"
	"github.com/qorio/maestro/pkg/zk"
	"github.com/qorio/omni/common"
	"time"
)

const (
	EnvZkHosts = "REDPILL_ZK_HOSTS"
)

type Service struct {
	conn      zk.ZK
	models    ModelStorage
	instances InstanceStorage
}

func NewService(pool func() zk.ZK,
	models func() ModelStorage,
	instances func() InstanceStorage) OrchestrateService {
	s := new(Service)
	s.conn = pool()
	s.models = models()
	s.instances = instances()
	return s
}

func (this *Service) ListOrchestrations(c Context, domain string) ([]Orchestration, error) {
	glog.Infoln("ListOrchestrations")
	models, err := this.models.GetModels(domain)
	if err != nil {
		return nil, err
	}
	orcs := []Orchestration{}
	for _, m := range models {
		orcs = append(orcs, m)
	}
	return orcs, nil
}

func (this *Model) NewInstance(domain string) *Instance {
	instance := &Instance{
		InstanceModel: *this,
		InstanceInfo: OrchestrationInfo{
			Domain:    domain,
			Id:        common.NewUUID().String(),
			StartTime: time.Now(),
		},
	}
	return instance
}

func (this *Service) ListInstances(c Context, domain, orchestration string) ([]OrchestrationInstance, error) {
	// TODO - check authorization
	instances := []OrchestrationInstance{}
	list, err := this.instances.List(domain, orchestration)
	if err != nil {
		return nil, err
	}
	for _, i := range list {
		instances = append(instances, i)
	}
	return instances, nil
}

func (this *Service) StartOrchestration(c Context, domain, orchestration string,
	input OrchestrationContext) (OrchestrationInstance, error) {

	model, err := this.models.Get(domain, orchestration)
	if err != nil {
		return nil, err
	}
	instance := model.NewInstance(domain)
	err = this.instances.Save(instance)
	if err != nil {
		return nil, err
	}
	return instance, nil
}

func (this *Service) GetOrchestration(c Context, domain, orchestration, instance string) (OrchestrationInstance, error) {
	return nil, ErrNotFound
}
