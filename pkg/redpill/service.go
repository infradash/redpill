package redpill

import (
	"sync"
)

var init_schema sync.Once

type Revision int32
type EnvService interface {
	GetEnv(domain, service, version string) (EnvList, Revision, error)
	UpdateEnv(domain, service, version string, EnvChange, rev Revision) error
}

type Registry interface {
	GetRegistry(key string) (string, error)
	UpdateRegistry(key, value string) error
	DeleteRegistry(key string) error
}

type DomainService interface {
	ListDomains() ([]Domain, error)
	GetDomain(domain string) (DomainDetail, error)
}

type Service interface {
	DomainService(*Context) DomainService

	Stop()
	Close() error
}

type serviceImpl struct {
	options Options
	lock    sync.Mutex
	stop    chan bool
}

func NewService(options Options) (Service, error) {
	impl := &serviceImpl{
		stop: make(chan bool),
	}

	return impl, nil
}

func (this *serviceImpl) Stop() {
	this.stop <- true
}

func (this *serviceImpl) Close() error {
	this.stop <- true
	return nil
}

func (this *serviceImpl) DomainService(c *Context) DomainService {
	d := &domainService{
		context: c,
	}
	return d
}
