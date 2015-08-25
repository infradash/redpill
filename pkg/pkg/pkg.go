package pkg

import (
	"github.com/golang/glog"
	. "github.com/infradash/redpill/pkg/api"
	"github.com/qorio/maestro/pkg/registry"
	"github.com/qorio/maestro/pkg/zk"
	"net/http"
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

func (this *Service) NewPkgModel(c Context, req *http.Request, um Unmarshaler) (PkgModel, error) {
	m := new(pkg)
	err := um(req, m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (this *Service) CreatePkg(c Context, domainClass, domainInstance, service, version string, spec PkgModel) error {
	pkgPath := GetPkgPath(domainClass, domainInstance, service, version)
	glog.Infoln("CreatePkg:", c.UserId(), "Path=", pkgPath)

	if zk.PathExists(this.conn, pkgPath) {
		return ErrConflict
	}
	return zk.CreateOrSet(this.conn, pkgPath, spec)
}

func (this *Service) UpdatePkg(c Context, domainClass, domainInstance, service, version string, spec PkgModel) error {
	pkgPath := GetPkgPath(domainClass, domainInstance, service, version)
	glog.Infoln("UpdatePkg:", c.UserId(), "Path=", pkgPath)

	if !zk.PathExists(this.conn, pkgPath) {
		return ErrNotFound
	}
	return zk.CreateOrSet(this.conn, pkgPath, spec)
}

func (this *Service) GetPkg(c Context, domainClass, domainInstance, service, version string) (PkgModel, error) {
	pkgPath := GetPkgPath(domainClass, domainInstance, service, version)
	glog.Infoln("GetPkg:", c.UserId(), "Path=", pkgPath)

	if !zk.PathExists(this.conn, pkgPath) {
		return nil, ErrNotFound
	}
	v := new(pkg)
	err := zk.GetObject(this.conn, pkgPath, v)
	return v, err
}

func (this *Service) DeletePkg(c Context, domainClass, domainInstance, service, version string) error {
	pkgPath := GetPkgPath(domainClass, domainInstance, service, version)
	glog.Infoln("DeletePkg:", c.UserId(), "Path=", pkgPath)

	if !zk.PathExists(this.conn, pkgPath) {
		return ErrNotFound
	}
	return zk.DeleteObject(this.conn, pkgPath)
}

func (this *Service) SetLive(c Context, domainClass, domainInstance, service, version string) error {
	pkgPath := GetPkgPath(domainClass, domainInstance, service, version)
	glog.Infoln("SetLive:", c.UserId(), "Path=", pkgPath)

	if !zk.PathExists(this.conn, pkgPath) {
		return ErrNotFound
	}

	err := zk.CreateOrSetString(this.conn, GetPkgLivePath(domainClass, domainInstance, service), pkgPath.Path())
	if err != nil {
		return err
	}
	return zk.Increment(this.conn, GetPkgWatchPath(domainClass, domainInstance, service), 1)
}

func (this *Service) GetPkgLiveVersion(c Context, domainClass, domainInstance, service string) (PkgModel, error) {
	p := zk.GetString(this.conn, GetPkgLivePath(domainClass, domainInstance, service))
	if p == nil {
		return nil, ErrNotFound
	}
	v := new(pkg)
	err := zk.GetObject(this.conn, registry.Path(*p), v)
	return v, err
}

func (this *Service) ListPkgVersions(c Context, domainClass, domainInstance, service string) (PkgVersions, error) {
	glog.Infoln("ListPkgVersions", "DomainClass=", domainClass, "DomainInstance=", domainInstance, "Service=", service)

	result := make(PkgVersions)
	err := VisitPkgVersions(this.conn, domainClass, domainInstance, service,
		func(data []byte) PkgModel {
			m := pkg{}
			m.FromBytes(data)
			return m
		},
		func(version string, model PkgModel) bool {
			result[version] = false
			return true
		})
	// read the live version
	realpath := zk.GetString(this.conn, GetPkgLivePath(domainClass, domainInstance, service))
	if realpath != nil {
		result[registry.NewPath(*realpath).Dir().Base()] = true
	}

	return result, err
}
