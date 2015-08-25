package api

import (
	"fmt"
	"github.com/qorio/maestro/pkg/registry"
)

func ToDomainName(domainClass, domainInstance string) string {
	return fmt.Sprintf("%s.%s", domainInstance, domainClass)
}

func GetEnvPath(domainClass, domainInstance, service, version string) registry.Path {
	return registry.NewPath(ToDomainName(domainClass, domainInstance), service, version, "env")
}

func GetPkgPath(domainClass, domainInstance, service, version string) registry.Path {
	return registry.NewPath(ToDomainName(domainClass, domainInstance), service, version, "pkg")
}

func GetConfPath(domainClass, service, name string) registry.Path {
	return registry.NewPath("_repill", "conf", domainClass, service, name)
}

func GetConfVersionPath(domainClass, domainInstance, service, version, name string) registry.Path {
	return registry.NewPath(ToDomainName(domainClass, domainInstance), service, version, name)
}

func GetEnvLivePath(domainClass, domainInstance, service string) registry.Path {
	return registry.NewPath(ToDomainName(domainClass, domainInstance), service, "_live", "_env")
}

func GetPkgLivePath(domainClass, domainInstance, service string) registry.Path {
	return registry.NewPath(ToDomainName(domainClass, domainInstance), service, "_live", "_pkg")
}

func GetConfLivePath(domainClass, domainInstance, service, name string) registry.Path {
	return registry.NewPath(ToDomainName(domainClass, domainInstance), service, "_live", name)
}

func GetEnvWatchPath(domainClass, domainInstance, service string) registry.Path {
	return registry.NewPath(ToDomainName(domainClass, domainInstance), service, "_watch", "_env")
}

func GetPkgWatchPath(domainClass, domainInstance, service string) registry.Path {
	return registry.NewPath(ToDomainName(domainClass, domainInstance), service, "_watch", "_pkg")
}

func GetConfWatchPath(domainClass, domainInstance, service, name string) registry.Path {
	return registry.NewPath(ToDomainName(domainClass, domainInstance), service, "_watch", name)
}
