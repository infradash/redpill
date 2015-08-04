package orchestrate

import (
	"github.com/golang/glog"
	. "github.com/infradash/redpill/pkg/api"
	_ "github.com/qorio/maestro/pkg/pubsub"
	"github.com/qorio/maestro/pkg/zk"
	"github.com/qorio/omni/common"
	"net/http"
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
			Name:      string(this.Name),
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

func (this *Service) StartOrchestration(c Context, domainClass, domainInstance, orchestration string, input OrchestrationContext, note ...string) (OrchestrationInstance, error) {
	glog.Infoln("Starting Orchestration=", orchestration, "DomainClass=", domainClass, "DomainInstance=", domainInstance)
	model, err := this.models.Get(domainClass, orchestration)
	if err != nil {
		return nil, err
	}
	if model == nil {
		return nil, ErrNotFound
	}

	domain := domainInstance + "." + domainClass

	instance := model.NewInstance(domain)
	instance.InstanceInfo.User = c.UserId()
	instance.InstanceInfo.Status = "Started"
	if len(note) > 0 {
		instance.InstanceInfo.Note = note[0]
	}
	instance.InstanceContext = input
	//	instance.InstanceLog = pubsub.Topic(key)
	err = this.instances.Save(instance)
	if err != nil {
		return nil, err
	}
	return instance, nil
}

func (this *Service) GetOrchestration(c Context, domain, orchestration, instance string) (OrchestrationInstance, error) {
	return this.instances.Get(instance)
}

func (this *Service) NewOrchestrationModel(c Context, req *http.Request, um Unmarshaler) (OrchestrationModel, error) {
	m := &Model{}
	err := um(req, m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (this *Service) SaveOrchestrationModel(c Context, domainClass string, m OrchestrationModel) error {
	if !m.IsOrchestrationModel(m) {
		return ErrTypeMismatch
	}
	return this.models.Save(domainClass, m.(*Model))
}
