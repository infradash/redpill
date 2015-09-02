package dockerapi

import (
	. "github.com/infradash/redpill/pkg/api"
	"github.com/qorio/maestro/pkg/zk"
	"github.com/qorio/omni/common"
)

type agent struct {
	Api       string                 `json:"api"`
	DockerApi string                 `json:"docker_api"`
	Version   map[string]interface{} `json:"version"`
}

func (this agent) IsDockerProxy(other interface{}) bool {
	return common.TypeMatch(this, other)
}

type Service struct {
	conn    zk.ZK
	domains DomainService
}

func NewService(pool func() zk.ZK, domains DomainService) DockerProxyService {
	s := new(Service)
	s.conn = pool()
	s.domains = domains
	return s
}

func (this *Service) GetProxy(c Context, domainClass, domainInstance, target string) (DockerProxy, error) {
	return nil, nil
}

func (this *Service) ListProxies(c Context, domainClass, domainInstance string) ([]DockerProxy, error) {
	return nil, nil
}
