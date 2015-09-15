package conf

import (
	"fmt"
	"github.com/golang/glog"
	. "github.com/infradash/redpill/pkg/api"
	"github.com/qorio/maestro/pkg/registry"
	"github.com/qorio/maestro/pkg/zk"
)

type Service struct {
	conn    zk.ZK
	storage ConfStorage
	domains DomainService
}

func NewService(pool func() zk.ZK, storage func() ConfStorage, domains DomainService) ConfService {
	return &Service{
		conn:    pool(),
		storage: storage(),
		domains: domains,
	}
}

type service_stat struct {
	objects   map[string]int
	instances map[string]int
	versions  map[string]int
	live      map[string]map[string]string
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

func (this *service_stat) get_objects() []string {
	l := []string{}
	for k, _ := range this.objects {
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

func (this *service_stat) set_live(instance, object, live string) {
	if _, has := this.live[instance]; !has {
		this.live[instance] = map[string]string{}
	}
	this.live[instance][object] = live
	this.objects[object] += 1
}

func (this *Service) ListDomainConfs(c Context, domainClass string) (map[string]Conf, error) {
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
					objects:   map[string]int{},
					instances: map[string]int{},
					versions:  map[string]int{},
					live:      map[string]map[string]string{},
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
				case "live": // TODO - Remove after transition
				case "_watch":
				case "_live":
					zobjects, err := zversion.Children()
					if err != nil {
						return nil, err
					}
					for _, zobject := range zobjects {
						switch zobject.GetBasename() {
						case "env", "pkg":
						default:
							service_stats[service].set_live(domainInstance, zobject.GetBasename(), zobject.GetValueString())
						}
					}
				default:
					service_stats[service].add_version(zversion.GetBasename())
				}
			}

			// the base versions -- look in the _redpill namespace:
			VisitConfs(this.conn, domainClass, service, func(n string) bool {
				service_stats[service].objects[n] += 1
				return true
			})
		}
	}

	confs := map[string]Conf{}
	// Now generate the metadata output based on the stats
	for service, stats := range service_stats {
		confs[service] = conf{
			Domain:    domainClass,
			Service:   service,
			Instances: stats.get_instances(),
			Versions:  stats.get_versions(),
			Live:      stats.live,
			Objects:   stats.get_objects(),
		}
	}
	return confs, nil
}

func (this *Service) ListConfs(c Context, domainClass, service string) ([]ConfInfo, error) {
	glog.Infoln("Listing confs", "DomainClass=", domainClass, "Service=", service)

	confs := []ConfInfo{}
	err := VisitConfs(this.conn, domainClass, service,
		func(name string) bool {
			c := confInfo{
				Domain:      domainClass,
				Service:     service,
				Name:        name,
				ContentType: "text/plain",
			}
			confs = append(confs, c)
			return true
		})
	return confs, err
}

func (this *Service) CreateConf(c Context, domainClass, service, name string, buff []byte) (Revision, error) {
	p := GetConfPath(domainClass, service, name)
	glog.Infoln("CreateConf", "Path=", p)
	if zk.PathExists(this.conn, p) {
		return -1, ErrConflict
	}

	v, err := zk.VersionLockAndExecute(this.conn, p, 0,
		func() error {
			return this.storage.Save(domainClass, service, name, buff)
		})
	return Revision(v), err
}

func (this *Service) UpdateConf(c Context, domainClass, service, name string, buff []byte, rev Revision) (Revision, error) {
	p := GetConfPath(domainClass, service, name)
	glog.Infoln("UpdateConf", "Path=", p)
	if !zk.PathExists(this.conn, p) {
		return -1, ErrNotFound
	}

	v, err := zk.VersionLockAndExecute(this.conn, p, int(rev),
		func() error {
			return this.storage.Save(domainClass, service, name, buff)
		})
	return Revision(v), err
}

func (this *Service) GetConf(c Context, domainClass, service, name string) ([]byte, Revision, error) {
	p := GetConfPath(domainClass, service, name)
	glog.Infoln("GetConf", "Path=", p)
	version := zk.GetInt(this.conn, p)
	if version == nil {
		return nil, -1, ErrNotFound
	}
	buff, err := this.storage.Get(domainClass, service, name)
	return buff, Revision(*version), err
}

func (this *Service) DeleteConf(c Context, domainClass, service, name string, rev Revision) error {
	p := GetConfPath(domainClass, service, name)
	glog.Infoln("DeleteConf", "Path=", p)
	current := zk.GetInt(this.conn, p)
	if current == nil {
		return ErrNotFound
	}
	if *current != int(rev) {
		return ErrConflict
	}
	if err := this.storage.Delete(domainClass, service, name); err == nil {
		return zk.DeleteObject(this.conn, p)
	} else {
		return err
	}
}

func (this *Service) CreateConfVersion(c Context, domainClass, domainInstance, service, version, name string,
	buff []byte) (Revision, error) {

	p := GetConfVersionPath(domainClass, domainInstance, service, version, name)
	glog.Infoln("CreateConfVersion", "Path=", p)
	if zk.PathExists(this.conn, p) {
		return -1, ErrConflict
	}

	v, err := zk.VersionLockAndExecute(this.conn, p, 0,
		func() error {
			return this.storage.SaveVersion(domainClass, domainInstance, service, version, name, buff)
		})
	return Revision(v), err
}

func (this *Service) UpdateConfVersion(c Context, domainClass, domainInstance, service, version, name string,
	buff []byte, rev Revision) (Revision, error) {

	p := GetConfVersionPath(domainClass, domainInstance, service, version, name)
	glog.Infoln("UpdateConfVersion", "Path=", p)
	if !zk.PathExists(this.conn, p) {
		if int(rev) != 0 {
			return -1, ErrNotFound
		}
	}

	v, err := zk.VersionLockAndExecute(this.conn, p, int(rev),
		func() error {
			return this.storage.SaveVersion(domainClass, domainInstance, service, version, name, buff)
		})
	return Revision(v), err
}

func (this *Service) GetConfVersion(c Context, domainClass, domainInstance, service, version, name string) ([]byte, Revision, error) {
	p := GetConfVersionPath(domainClass, domainInstance, service, version, name)
	glog.Infoln("GetConfVersion", "Path=", p)
	rev := zk.GetInt(this.conn, p)
	if rev == nil {
		rev = new(int) // Use virtual copy
		*rev = 0
	}
	buff, err := this.storage.GetVersion(domainClass, domainInstance, service, version, name)
	return buff, Revision(*rev), err
}

func (this *Service) DeleteConfVersion(c Context, domainClass, domainInstance, service, version, name string,
	rev Revision) error {
	p := GetConfVersionPath(domainClass, domainInstance, service, version, name)
	glog.Infoln("DeleteConfVersion", "Path=", p)
	current := zk.GetInt(this.conn, p)
	if current == nil {
		return ErrNotFound
	}
	if *current != int(rev) {
		return ErrConflict
	}
	if err := this.storage.DeleteVersion(domainClass, domainInstance, service, version, name); err != nil {
		return zk.DeleteObject(this.conn, p)
	} else {
		return err
	}
}

func (this *Service) SetLive(c Context, domainClass, domainInstance, service, version, name string) error {
	if !zk.PathExists(this.conn, GetConfVersionPath(domainClass, domainInstance, service, version, name)) {

		// Copy on write - when we set live and this version is a virtual version, make a copy.
		if copy, _, err := this.GetConfVersion(c, domainClass, domainInstance, service, version, name); err == nil {
			if _, err := this.CreateConfVersion(c, domainClass, domainInstance, service, version, name, copy); err != nil {
				return err
			}
		} else {
			return err
		}

		// check again
		if !zk.PathExists(this.conn, GetConfVersionPath(domainClass, domainInstance, service, version, name)) {
			return ErrCannotCreateCopy
		}
	}
	p := GetConfLivePath(domainClass, domainInstance, service, name)
	glog.Infoln("Setlive", "Path=", p)
	err := zk.CreateOrSetString(this.conn, p, version)
	if err != nil {
		return err
	}
	// Watch nodes
	err = zk.Increment(this.conn, GetConfWatchPath(domainClass, domainInstance, service, name), 1)
	return err
}

func (this *Service) ListConfVersions(c Context, domainClass, domainInstance, service, name string) (ConfVersions, error) {
	glog.Infoln("ListConfVersions", "DomainClass=", domainClass, "DomainInstance=", domainInstance, "Service=", service)

	result := make(ConfVersions)
	err := VisitConfVersions(this.conn, domainClass, domainInstance, service, name,
		func(version string) bool {
			result[version] = false
			return true
		})
	// read the live version
	version := zk.GetString(this.conn, GetConfLivePath(domainClass, domainInstance, service, name))
	if version != nil {
		result[*version] = true
		return result, err
	} else {
		return nil, ErrNotFound
	}

}

func (this *Service) ListConfLiveVersions(c Context, domainClass, domainInstance, service string) (ConfLiveVersions, error) {
	domain := fmt.Sprintf("/%s.%s", domainInstance, domainClass)

	glog.Infoln("ListConfLiveVersions", domain, service)

	result := make(ConfLiveVersions)
	err := zk.Visit(this.conn, registry.NewPath(domain, service, "_live"),
		func(p registry.Path, v []byte) bool {
			switch p.Base() {
			case "_env", "_pkg":
			default:
				result[p.Base()] = LiveVersion(string(v))
			}
			return true
		})
	return result, err
}

func (this *Service) GetConfLiveVersion(c Context, domainClass, domainInstance, service, name string) ([]byte, error) {
	version := zk.GetString(this.conn, GetConfLivePath(domainClass, domainInstance, service, name))
	if version != nil {
		buff, _, err := this.GetConfVersion(c, domainClass, domainInstance, service, *version, name)
		return buff, err
	}
	return nil, ErrNotFound
}
