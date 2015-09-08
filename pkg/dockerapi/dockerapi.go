package dockerapi

import (
	"encoding/json"
	. "github.com/infradash/redpill/pkg/api"
	"github.com/qorio/maestro/pkg/zk"
	"github.com/qorio/omni/common"
)

type agent struct {
	Api       string                 `json:"api"`
	DockerApi string                 `json:"dockerapi"`
	Version   map[string]interface{} `json:"version"`
}

func (this agent) IsDockerProxy(other interface{}) bool {
	return common.TypeMatch(this, other)
}

func (this agent) GetApiProxy() (string, error) {
	if this.DockerApi == "" {
		return "", ErrNoApiProxy
	}
	return this.DockerApi, nil
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
	a := new(agent)
	err := zk.GetObject(this.conn, GetDockerProxyPath(domainClass, domainInstance, target), a)
	return a, err
}

func (this *Service) ListProxies(c Context, domainClass, domainInstance string) (DockerProxies, error) {
	result := make(DockerProxies)

	err := VisitDockerProxies(this.conn, domainClass, domainInstance,
		func(buff []byte) DockerProxy {
			a := new(agent)
			if err := json.Unmarshal(buff, a); err == nil {
				return a
			} else {
				return nil
			}
		},
		func(agent string, proxy DockerProxy) bool {
			if proxy != nil {
				result[agent] = proxy
				return true
			}
			return false
		})

	return result, err
}
