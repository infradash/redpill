package conf

import ()

type ConfStorage interface {
	Save(domainClass, service, name string, content []byte) error
	Get(domainClass, service, name string) ([]byte, error)
	Delete(domainClass, service, name string) error
	ListAll(domainClass, service string) ([]string, []int, error)

	SaveVersion(domainClass, domainInstance, service, name, version string, content []byte) error
	GetVersion(domainClass, domainInstance, service, name, version string) ([]byte, error)
	DeleteVersion(domainClass, domainInstance, service, name, version string) error
}
