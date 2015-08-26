package api

type ConfInfo interface {
	IsConfInfo(other interface{}) bool
}

type Conf interface {
	IsConf(other interface{}) bool
}

type ConfVersions map[string]bool

type LiveVersion string
type ConfLiveVersions map[string]LiveVersion

type ConfService interface {
	ListDomainConfs(c Context, domainClass string) (map[string]Conf, error)

	// Lists the 'class' level conf objects
	ListConfs(c Context, domainClass, service string) ([]ConfInfo, error)

	// Lists the instances of the conf objects -- per domain instance
	ListConfLiveVersions(c Context, domainClass, domainInstance, service string) (ConfLiveVersions, error)

	CreateConf(c Context, domainClass, service, name string, buff []byte) (Revision, error)
	UpdateConf(c Context, domainClass, service, name string, buff []byte, rev Revision) (Revision, error)
	GetConf(c Context, domainClass, service, name string) ([]byte, Revision, error)
	DeleteConf(c Context, domainClass, service, name string, rev Revision) error

	CreateConfVersion(c Context, domainClass, domainInstance, service, version, name string, buff []byte) (Revision, error)
	UpdateConfVersion(c Context, domainClass, domainInstance, service, version, name string, buff []byte, rev Revision) (Revision, error)
	GetConfVersion(c Context, domainClass, domainInstance, service, version, name string) ([]byte, Revision, error)
	DeleteConfVersion(c Context, domainClass, domainInstance, service, version, name string, rev Revision) error

	SetLive(c Context, domainClass, domainInstance, service, version, name string) error
	ListConfVersions(c Context, domainClass, domainInstance, service, name string) (ConfVersions, error)
	GetConfLiveVersion(c Context, domainClass, domainInstance, service, name string) ([]byte, error)
}
