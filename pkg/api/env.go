package api

type Env struct {
	Domain    string            `json:"domain"`
	Service   string            `json:"service"`
	Instances []string          `json:"instances"`
	Versions  []string          `json:"versions"`
	Live      map[string]string `json:"live"`
}

type EnvList map[string]interface{}
type EnvChange struct {
	Update EnvList  `json:"update,omitempty"`
	Delete []string `json:"delete,omitempty"`
}

type EnvVersions map[string]bool

type EnvService interface {
	ListDomainEnvs(c Context, domainClass string) (map[string]Env, error)
	GetEnv(c Context, domain, service, version string) (EnvList, Revision, error)
	SaveEnv(c Context, domain, service, version string, change *EnvChange, rev Revision) (Revision, error)
	NewEnv(c Context, domain, service, version string, vars *EnvList) (Revision, error)
	SetLive(c Context, domain, service, version string) error
	ListEnvVersions(c Context, domain, service string) (EnvVersions, error)
	GetEnvLiveVersion(c Context, domain, service string) (EnvList, error)
}
