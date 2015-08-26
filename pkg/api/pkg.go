package api

import (
	"net/http"
)

type Pkg interface {
	IsPkg(other interface{}) bool
}

type PkgModel interface {
	IsPkgModel(other interface{}) bool
}

type PkgVersions map[string]bool

type PkgService interface {
	NewPkgModel(c Context, req *http.Request, um Unmarshaler) (PkgModel, error)

	ListDomainPkgs(c Context, domainClass string) (map[string]Pkg, error)

	CreatePkg(c Context, domainClass, domainInstance, service, version string, spec PkgModel) error
	UpdatePkg(c Context, domainClass, domainInstance, service, version string, spec PkgModel) error
	GetPkg(c Context, domainClass, domainInstance, service, version string) (PkgModel, error)
	DeletePkg(c Context, domainClass, domainInstance, service, version string) error

	SetLive(c Context, domainClass, domainInstance, service, version string) error
	GetPkgLiveVersion(c Context, domainClass, domainInstance, service string) (PkgModel, error)
	ListPkgVersions(c Context, domainClass, domainInstance, service string) (PkgVersions, error)
}
