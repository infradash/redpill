package api

import (
	"net/http"
)

type Revision int32
type Unmarshaler func(*http.Request, interface{}) error

type EnvService interface {
	ListEnvs(c Context, domainClass string) ([]Env, error)
	GetEnv(c Context, domain, service, version string) (EnvList, Revision, error)
	SaveEnv(c Context, domain, service, version string, change *EnvChange, rev Revision) error
	NewEnv(c Context, domain, service, version string, vars *EnvList) (rev Revision, err error)
}

type RegistryService interface {
	GetEntry(c Context, key string) ([]byte, Revision, error)
	UpdateEntry(c Context, key string, value []byte, rev Revision) (Revision, error)
	DeleteEntry(c Context, key string, rev Revision) error
}

type DomainService interface {
	ListDomains(c Context) ([]Domain, error)
	GetDomain(c Context, domainClass string) (*DomainDetail, error)
}

type OrchestrateService interface {
	ListOrchestrations(c Context, domainClass string) ([]Orchestration, error)
	StartOrchestration(c Context, domainClass, domainInstance, orchestration string, input OrchestrationContext, note ...string) (OrchestrationInstance, error)
	GetOrchestration(c Context, domain, orchestration, instance string) (OrchestrationInstance, error)
	ListInstances(c Context, domain, orchestration string) ([]OrchestrationInstance, error)

	NewOrchestrationModel(c Context, req *http.Request, um Unmarshaler) (OrchestrationModel, error)
	SaveOrchestrationModel(c Context, domainClass string, m OrchestrationModel) error
	GetOrchestrationModel(c Context, domainClass, orchestration string) (OrchestrationModel, error)
	DeleteOrchestrationModel(c Context, domainClass, orchestration string) error
}

type Conf interface {
	IsConf(other interface{}) bool
}

type ConfService interface {
	ListConfs(c Context, domainClass, service string) ([]Conf, error)
	SaveConf(c Context, domainClass, service, name string, buff []byte) error
	GetConf(c Context, domainClass, service, name string) ([]byte, error)
	DeleteConf(c Context, domainClass, service, name string) error

	SaveConfVersion(c Context, domainClass, domainInstance, service, name, version string, buff []byte) error
	GetConfVersion(c Context, domainClass, domainInstance, service, name, version string) ([]byte, error)
	DeleteConfVersion(c Context, domainClass, domainInstance, service, name, version string) error
}
