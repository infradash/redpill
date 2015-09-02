package api

type DockerProxy interface {
	IsDockerProxy(other interface{}) bool
}

type DockerProxyService interface {
	ListProxies(c Context, domainClass, domainInstance string) ([]DockerProxy, error)
	GetProxy(c Context, domainClass, domainInstance, target string) (DockerProxy, error)
}
