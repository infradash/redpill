package api

import (
	"fmt"
	"github.com/qorio/maestro/pkg/registry"
	"github.com/qorio/maestro/pkg/zk"
)

func ToDomainName(domainClass, domainInstance string) string {
	return fmt.Sprintf("%s.%s", domainInstance, domainClass)
}

func GetDomainPath(domainClass, domainInstance string) registry.Path {
	return registry.NewPath(ToDomainName(domainClass, domainInstance))
}
func GetEnvPath(domainClass, domainInstance, service, version string) registry.Path {
	return registry.NewPath(ToDomainName(domainClass, domainInstance), service, version, "env")
}

func GetPkgPath(domainClass, domainInstance, service, version string) registry.Path {
	return registry.NewPath(ToDomainName(domainClass, domainInstance), service, version, "pkg")
}

func GetConfPath(domainClass, service, name string) registry.Path {
	return registry.NewPath("_redpill", "conf", domainClass, service, name)
}

func GetConfVersionPath(domainClass, domainInstance, service, version, name string) registry.Path {
	return registry.NewPath(ToDomainName(domainClass, domainInstance), service, version, "conf", name)
}

func GetEnvLivePath(domainClass, domainInstance, service string) registry.Path {
	return registry.NewPath(ToDomainName(domainClass, domainInstance), service, "_live", "env")
}

func GetPkgLivePath(domainClass, domainInstance, service string) registry.Path {
	return registry.NewPath(ToDomainName(domainClass, domainInstance), service, "_live", "pkg")
}

func GetConfLivePath(domainClass, domainInstance, service, name string) registry.Path {
	return registry.NewPath(ToDomainName(domainClass, domainInstance), service, "_live", name)
}

func GetEnvWatchPath(domainClass, domainInstance, service string) registry.Path {
	return registry.NewPath(ToDomainName(domainClass, domainInstance), service, "_watch", "env")
}

func GetPkgWatchPath(domainClass, domainInstance, service string) registry.Path {
	return registry.NewPath(ToDomainName(domainClass, domainInstance), service, "_watch", "pkg")
}

func GetConfWatchPath(domainClass, domainInstance, service, name string) registry.Path {
	return registry.NewPath(ToDomainName(domainClass, domainInstance), service, "_watch", name)
}

func GetDockerProxyPath(domainClass, domainInstance, target string) registry.Path {
	return registry.NewPath(ToDomainName(domainClass, domainInstance), "dash", target)
}

func VisitEnvVersions(zc zk.ZK, domainClass, domainInstance, service string,
	visit func(version string, parent *zk.Node) bool) error {

	return zk.Visit(zc, registry.NewPath(ToDomainName(domainClass, domainInstance), service),
		func(p registry.Path, v []byte) bool {
			switch p.Base() {
			case "live", "_live", "_watch":
			default:
				if envs, err := zc.Get(p.Sub("env").Path()); err == nil {
					if !visit(p.Base(), envs) {
						return false
					}
				}
			}
			return true
		})
}

// Visit all the known default confs for a service
func VisitConfs(zc zk.ZK, domainClass, service string, visit func(name string) bool) error {
	return zk.Visit(zc, registry.NewPath("_redpill", "conf", domainClass, service),
		func(p registry.Path, v []byte) bool {
			if !visit(p.Base()) {
				return false
			}
			return true
		})
}

// Visit versions of a named conf in a given environment
func VisitConfVersions(zc zk.ZK, domainClass, domainInstance, service, name string,
	visit func(version string) bool) error {

	return zk.Visit(zc, registry.NewPath(ToDomainName(domainClass, domainInstance), service),
		func(p registry.Path, v []byte) bool {
			switch p.Base() {
			case "live", "_live", "_watch":
			default:
				if !visit(p.Base()) {
					return false
				}
			}
			return true
		})
}

func VisitPkgVersions(zc zk.ZK, domainClass, domainInstance, service string,
	unmarshal func([]byte) PkgModel,
	visit func(version string, pkg PkgModel) bool) error {

	return zk.Visit(zc, registry.NewPath(ToDomainName(domainClass, domainInstance), service),
		func(p registry.Path, v []byte) bool {
			switch p.Base() {
			case "live", "_live", "_watch":
			default:
				if !visit(p.Base(), unmarshal(v)) {
					return false
				}
			}
			return true
		})
}

func VisitDockerProxies(zc zk.ZK, domainClass, domainInstance string,
	unmarshal func([]byte) DockerProxy,
	visit func(agent string, proxy DockerProxy) bool) error {

	return zk.Visit(zc, registry.NewPath(ToDomainName(domainClass, domainInstance), "dash"),
		func(p registry.Path, v []byte) bool {
			switch p.Base() {
			default:
				if !visit(p.Base(), unmarshal(v)) {
					return false
				}
			}
			return true
		})
}
