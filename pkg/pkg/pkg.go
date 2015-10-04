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
	if err == zk.ErrNotExist {
		return v, ErrNotFound
	}
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

type service_stat struct {
	instances map[string]int
	versions  map[string]int
	live      map[string]PkgModel
}

func (this *service_stat) add_instance(instance string) {
	this.instances[instance] += 1
}

func (this *service_stat) get_instances() []string {
	l := []string{}
	for k, _ := range this.instances {
		l = append(l, k)
	}
	return l
}

func (this *service_stat) add_version(version string) {
	this.versions[version] += 1
}

func (this *service_stat) get_versions() []string {
	l := []string{}
	for k, _ := range this.versions {
		l = append(l, k)
	}
	return l
}

func (this *service_stat) set_live(instance string, model PkgModel) {
	this.live[instance] = model
}

func (this *Service) ListDomainPkgs(c Context, domainClass string) (map[string]Pkg, error) {
	model, err := this.domains.GetDomain(c, domainClass)
	if err != nil {
		return nil, err
	}

	// collect information by service
	service_stats := map[string]*service_stat{}

	// Build the fully qualified name for each domain
	for _, domainInstance := range model.DomainInstances() {
		// Get the services
		zdomain, err := this.conn.Get(GetDomainPath(domainClass, domainInstance).Path())
		if err != nil {
			glog.Warningln("Err=", err)
			return nil, err
		}
		zservices, err := zdomain.Children()
		if err != nil {
			glog.Warningln("Err=", err)
			return nil, err
		}
		// get the versions
		for _, zservice := range zservices {
			service := zservice.GetBasename()

			if _, has := service_stats[service]; !has {
				service_stats[service] = &service_stat{
					instances: map[string]int{},
					versions:  map[string]int{},
					live:      map[string]PkgModel{},
				}
			}
			// an instance
			service_stats[service].add_instance(domainInstance)

			zversions, err := zservice.Children()
			if err != nil {
				glog.Warningln("Err=", err)
				return nil, err
			}
			for _, zversion := range zversions {

				switch zversion.GetBasename() {
				case "live":
				case "_watch", "_live": // skip
				default:
					// a version
					service_stats[service].add_version(zversion.GetBasename())
				}
			}

			// Get the live version
			if realpath := zk.GetString(this.conn, GetPkgLivePath(domainClass, domainInstance, service)); realpath != nil {
				m := new(pkg)
				if err := zk.GetObject(this.conn, registry.Path(*realpath), m); err == nil {
					service_stats[service].set_live(domainInstance, m)
				}
			}
		}
	}

	packages := map[string]Pkg{}
	// Now generate the metadata output based on the stats
	for service, stats := range service_stats {
		packages[service] = pkgInfo{
			Domain:    domainClass,
			Service:   service,
			Instances: stats.get_instances(),
			Versions:  stats.get_versions(),
			Live:      stats.live,
		}
	}
	return packages, nil
}
