package api

import (
	"errors"
)

var (
	ErrNoApiProxy = errors.New("no-api-proxy")
)

type DockerProxies map[string]DockerProxy
type DockerProxy interface {
	IsDockerProxy(other interface{}) bool
	GetApiProxy() (string, error)
}

type DockerProxyService interface {
	ListProxies(c Context, domainClass, domainInstance string) (DockerProxies, error)
	GetProxy(c Context, domainClass, domainInstance, target string) (DockerProxy, error)
}
