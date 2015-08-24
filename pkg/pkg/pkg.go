package pkg

import (
	"github.com/golang/glog"
	. "github.com/infradash/redpill/pkg/api"
	"github.com/qorio/maestro/pkg/zk"
)

type Service struct {
	conn    zk.ZK
	domains DomainService
}

func NewService(pool func() zk.ZK, domains DomainService) PkgService {
	s := new(Service)
	s.conn = pool()
	s.domains = domains
	return s
}

func (this *Service) CreatePkg(c Context, domainClass, domainInstance, service, version string, spec Pkg) error {
	glog.Infoln("CreatePkg:", c.UserId(), "Domain=", ToDomainName(domainClass, domainInstance),
		"Service=", service, "Version=", version)
	return nil
}

func (this *Service) UpdatePkg(c Context, domainClass, domainInstance, service, version string, spec Pkg) error {
	glog.Infoln("UpdatePkg:", c.UserId(), "Domain=", ToDomainName(domainClass, domainInstance),
		"Service=", service, "Version=", version)
	return nil
}

func (this *Service) GetPkg(c Context, domainClass, domainInstance, service, version string) (Pkg, error) {
	glog.Infoln("GetPkg", ToDomainName(domainClass, domainInstance), service)
	return nil, nil
}

func (this *Service) DeletePkg(c Context, domainClass, domainInstance, service, version string) error {
	glog.Infoln("DeletePkg", ToDomainName(domainClass, domainInstance), service)
	return nil
}

func (this *Service) SetLive(c Context, domainClass, domainInstance, service, version string) error {
	glog.Infoln("ListPkgVersions", ToDomainName(domainClass, domainInstance), service)
	return nil
}

func (this *Service) ListPkgVersions(c Context, domainClass, domainInstance, service string) (PkgVersions, error) {
	glog.Infoln("ListPkgVersions", ToDomainName(domainClass, domainInstance), service)
	return nil, nil
}

func (this *Service) GetPkgLiveVersion(c Context, domainClass, domainInstance, service string) (Pkg, error) {
	glog.Infoln("GetPkgLiveVersion", ToDomainName(domainClass, domainInstance), service)
	return nil, nil
}
