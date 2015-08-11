package api

import (
	"net/http"
)

type Revision int32
type Unmarshaler func(*http.Request, interface{}) error

type RegistryService interface {
	GetEntry(c Context, key string) ([]byte, Revision, error)
	UpdateEntry(c Context, key string, value []byte, rev Revision) (Revision, error)
	DeleteEntry(c Context, key string, rev Revision) error
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
