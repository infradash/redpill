package api

import (
	"net/http"
)

type DomainInfo interface {
	IsDomainInfo(other interface{}) bool
}

type DomainModel interface {
	IsDomainModel(other interface{}) bool
	DomainClass() string
	DomainInstances() []string
}

type DomainService interface {
	NewDomainModel(c Context, req *http.Request, um Unmarshaler) (DomainModel, error)
	ListDomains(c Context) ([]DomainInfo, error)
	CreateDomain(c Context, model DomainModel) error
	UpdateDomain(c Context, domainClass string, model DomainModel) error
	GetDomain(c Context, domainClass string) (DomainModel, error)
}
