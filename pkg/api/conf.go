package api

import ()

type ConfInfo interface {
	IsConfInfo(other interface{}) bool
}

type Conf interface {
	IsConf(other interface{}) bool
}

type ConfService interface {
	ListDomainConfs(c Context, domainClass string) (map[string]Conf, error)
	ListConfs(c Context, domainClass, service string) ([]ConfInfo, error)
	SaveConf(c Context, domainClass, service, name string, buff []byte, rev Revision) error
	GetConf(c Context, domainClass, service, name string) ([]byte, Revision, error)
	DeleteConf(c Context, domainClass, service, name string, rev Revision) error

	SaveConfVersion(c Context, domainClass, domainInstance, service, name, version string, buff []byte, rev Revision) error
	GetConfVersion(c Context, domainClass, domainInstance, service, name, version string) ([]byte, Revision, error)
	DeleteConfVersion(c Context, domainClass, domainInstance, service, name, version string, rev Revision) error
}
