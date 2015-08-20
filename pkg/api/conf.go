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
	SaveConf(c Context, domainClass, service, name string, buff []byte) error
	GetConf(c Context, domainClass, service, name string) ([]byte, error)
	DeleteConf(c Context, domainClass, service, name string) error

	SaveConfVersion(c Context, domainClass, domainInstance, service, name, version string, buff []byte) error
	GetConfVersion(c Context, domainClass, domainInstance, service, name, version string) ([]byte, error)
	DeleteConfVersion(c Context, domainClass, domainInstance, service, name, version string) error
}
