package api

type Revision int32
type EnvService interface {
	GetEnv(c Context, domain, service, version string) (EnvList, Revision, error)
	SaveEnv(c Context, domain, service, version string, change *EnvChange, rev Revision) error
}

type RegistryService interface {
	GetEntry(c Context, key string) ([]byte, Revision, error)
	UpdateEntry(c Context, key string, value []byte, rev Revision) (Revision, error)
	DeleteEntry(c Context, key string, rev Revision) error
}

type DomainService interface {
	ListDomains(c Context) ([]Domain, error)
	GetDomain(c Context, domain string) (DomainDetail, error)
}
