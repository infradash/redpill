package api

type Context interface {
	UserId() string
}

type Revision int32
type EnvService interface {
	GetEnv(c Context, domain, service, version string) (EnvList, Revision, error)
	SaveEnv(c Context, domain, service, version string, change *EnvChange, rev Revision) error
}

type Registry interface {
	GetRegistry(c Context, key string) (string, error)
	UpdateRegistry(c Context, key, value string) error
	DeleteRegistry(c Context, key string) error
}

type DomainService interface {
	ListDomains(c Context) ([]Domain, error)
	GetDomain(c Context, domain string) (DomainDetail, error)
}
