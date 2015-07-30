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
	conn   zk.ZK
	models ModelStorage
}

func NewService(pool func() zk.ZK, models func() ModelStorage) OrchestrateService {
	s := new(Service)
	s.conn = pool()
	s.models = models()
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

func (this *Service) ListRunningOrchestrations(c Context, domain string) ([]OrchestrationInstance, error) {
	return []OrchestrationInstance{}, nil
}

func (this *Service) StartOrchestration(c Context, domain, orchestration string, input OrchestrationContext) (OrchestrationInstance, error) {
	return &Instance{
		id:        common.NewUUID().String(),
		startTime: time.Now(),
	}, nil
}

func (this *Service) GetOrchestration(c Context, domain, orchestration, instance string) (OrchestrationInstance, error) {
	return nil, ErrNotFound
}
