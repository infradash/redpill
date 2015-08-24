package api

import ()

type Pkg interface {
	IsPkg(other interface{}) bool
}

type PkgVersions map[string]bool

type PkgService interface {
	CreatePkg(c Context, domainClass, domainInstance, service, version string, spec Pkg) error
	UpdatePkg(c Context, domainClass, domainInstance, service, version string, spec Pkg) error
	GetPkg(c Context, domainClass, domainInstance, service, version string) (Pkg, error)
	DeletePkg(c Context, domainClass, domainInstance, service, version string) error

	SetLive(c Context, domainClass, domainInstance, service, version string) error
	GetPkgLiveVersion(c Context, domainClass, domainInstance, service string) (Pkg, error)
	ListPkgVersions(c Context, domainClass, domainInstance, service string) (PkgVersions, error)
}
