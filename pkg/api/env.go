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

type EnvService interface {
	ListDomainEnvs(c Context, domainClass string) ([]Env, error)
	GetEnv(c Context, domain, service, version string) (EnvList, Revision, error)
	SaveEnv(c Context, domain, service, version string, change *EnvChange, rev Revision) error
	NewEnv(c Context, domain, service, version string, vars *EnvList) (rev Revision, err error)
}
