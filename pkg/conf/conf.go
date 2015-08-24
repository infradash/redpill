package conf

import (
	"fmt"
	"github.com/golang/glog"
	. "github.com/infradash/redpill/pkg/api"
	"github.com/qorio/maestro/pkg/registry"
	"github.com/qorio/maestro/pkg/zk"
	"github.com/qorio/omni/common"
)

type Service struct {
	conn    zk.ZK
	storage ConfStorage
	domains DomainService
}

type confInfo struct {
	Domain      string `json:"domain"`
	Service     string `json:"service"`
	Name        string `json:"name"`
	Size        int    `json:"size"`
	ContentType string `json:"content_type"`
}

func (this confInfo) IsConfInfo(other interface{}) bool {
	return common.TypeMatch(this, other)
}

type conf struct {
	Domain    string                       `json:"domain"`
	Service   string                       `json:"service"`
	Instances []string                     `json:"instances"`
	Objects   []string                     `json:"objects"`
	Versions  []string                     `json:"versions"`
	Live      map[string]map[string]string `json:"live"`
}

func (this conf) IsConf(other interface{}) bool {
	return common.TypeMatch(this, other)
}

func NewService(pool func() zk.ZK, storage func() ConfStorage, domains DomainService) ConfService {
	return &Service{
		conn:    pool(),
		storage: storage(),
		domains: domains,
	}
}

func (this *Service) CreateConf(c Context, domainClass, service, name string, buff []byte) (Revision, error) {
	glog.Infoln("CreateConf", "DomainClass=", domainClass,
		"Service=", service, "Name=", name,
		"Content=", string(buff))
	p := registry.NewPath("_redpill", "conf", domainClass, service, name)
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
	glog.Infoln("UpdateConf", "DomainClass=", domainClass,
		"Service=", service, "Name=", name, "Rev=", rev,
		"Content=", string(buff))
	p := registry.NewPath("_redpill", "conf", domainClass, service, name)
	if !zk.PathExists(this.conn, p) {
		return -1, ErrNotFound
	}

	v, err := zk.VersionLockAndExecute(this.conn, p, int(rev),
		func() error {
			return this.storage.Save(domainClass, service, name, buff)
		})
	return Revision(v), err
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
		p := fmt.Sprintf("/%s.%s", domainInstance, domainClass)
		zdomain, err := this.conn.Get(p)
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
						case "_env", "_pkg":
						default:
							service_stats[service].set_live(domainInstance, zobject.GetBasename(), zobject.GetValueString())
						}
					}
				default:
					service_stats[service].add_version(zversion.GetBasename())
				}
			}
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
	keys, sizes, err := this.storage.ListAll(domainClass, service)
	if err != nil {
		return nil, err
	}
	confs := []ConfInfo{}
	for i, k := range keys {
		c := confInfo{
			Domain:      domainClass,
			Service:     service,
			Name:        k,
			ContentType: "text/plain",
			Size:        sizes[i],
		}
		confs = append(confs, c)
	}
	return confs, nil
}

func (this *Service) GetConf(c Context, domainClass, service, name string) ([]byte, Revision, error) {
	glog.Infoln("GetConf DomainClass=", domainClass, "Service=", service, "Name=", name)
	p := registry.NewPath("_redpill", "conf", domainClass, service, name)
	version := zk.GetInt(this.conn, p)
	if version == nil {
		return nil, -1, ErrNotFound
	}
	buff, err := this.storage.Get(domainClass, service, name)
	return buff, Revision(*version), err
}

func (this *Service) DeleteConf(c Context, domainClass, service, name string, rev Revision) error {
	glog.Infoln("DeleteConf DomainClass=", domainClass, "Service=", service, "Name=", name)
	p := registry.NewPath("_redpill", "conf", domainClass, service, name)
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

func (this *Service) CreateConfVersion(c Context,
	domainClass, domainInstance, service, name, version string, buff []byte) (Revision, error) {
	glog.Infoln("CreateConfVersion DomainClass=", domainClass, "Service=", service, "Name=", name,
		"DomainInstance=", domainInstance, "Version=", version)
	domain := fmt.Sprintf("%s.%s", domainInstance, domainClass)

	p := registry.NewPath(domain, service, version, name)
	if zk.PathExists(this.conn, p) {
		return -1, ErrConflict
	}

	v, err := zk.VersionLockAndExecute(this.conn, p, 0,
		func() error {
			return this.storage.SaveVersion(domainClass, domainInstance, service, name, version, buff)
		})
	return Revision(v), err
}

func (this *Service) UpdateConfVersion(c Context,
	domainClass, domainInstance, service, name, version string,
	buff []byte, rev Revision) (Revision, error) {
	glog.Infoln("UpdateConfVersion DomainClass=", domainClass, "Service=", service, "Name=", name,
		"DomainInstance=", domainInstance, "Version=", version, "Rev=", rev)
	domain := fmt.Sprintf("%s.%s", domainInstance, domainClass)

	p := registry.NewPath(domain, service, version, name)
	if !zk.PathExists(this.conn, p) {
		if int(rev) != 0 {
			return -1, ErrNotFound
		}
	}

	v, err := zk.VersionLockAndExecute(this.conn, p, int(rev),
		func() error {
			return this.storage.SaveVersion(domainClass, domainInstance, service, name, version, buff)
		})
	return Revision(v), err
}

func (this *Service) GetConfVersion(c Context, domainClass, domainInstance, service, name, version string) ([]byte, Revision, error) {
	glog.Infoln("GetConfVersion DomainClass=", domainClass, "Service=", service, "Name=", name,
		"DomainInstance=", domainInstance, "Version=", version)
	domain := fmt.Sprintf("%s.%s", domainInstance, domainClass)
	p := registry.NewPath(domain, service, version, name)
	rev := zk.GetInt(this.conn, p)
	if rev == nil {
		rev = new(int) // Use virtual copy
		*rev = 0
	}
	buff, err := this.storage.GetVersion(domainClass, domainInstance, service, name, version)
	return buff, Revision(*rev), err
}

func (this *Service) DeleteConfVersion(c Context,
	domainClass, domainInstance, service, name, version string, rev Revision) error {
	glog.Infoln("DeleteConfVersion DomainClass=", domainClass, "Service=", service, "Name=", name,
		"DomainInstance=", domainInstance, "Version=", version, "Rev=", rev)
	domain := fmt.Sprintf("%s.%s", domainInstance, domainClass)
	p := registry.NewPath(domain, service, version, name)
	current := zk.GetInt(this.conn, p)
	if current == nil {
		return ErrNotFound
	}
	if *current != int(rev) {
		return ErrConflict
	}
	if err := this.storage.DeleteVersion(domainClass, domainInstance, service, name, version); err != nil {
		return zk.DeleteObject(this.conn, p)
	} else {
		return err
	}
}

func (this *Service) SetLive(c Context, domainClass, domainInstance, service, version, name string) error {
	domain := fmt.Sprintf("/%s.%s", domainInstance, domainClass)
	livepath := registry.NewPath(domain, service, "_live", name)
	err := zk.CreateOrSetString(this.conn, livepath, version)
	if err != nil {
		return err
	}
	// Watch nodes
	err = zk.Increment(this.conn, registry.NewPath(domain, service, "_watch", name), 1)
	return err
}

func (this *Service) ListConfVersions(c Context, domainClass, domainInstance, service, name string) (ConfVersions, error) {
	domain := fmt.Sprintf("/%s.%s", domainInstance, domainClass)

	glog.Infoln("ListConfVersions", domain, service)

	result := make(ConfVersions)
	err := zk.Visit(this.conn, registry.NewPath(domain, service),
		func(p registry.Path, v []byte) bool {
			switch p.Base() {
			case "live", "_live", "_watch":
			default:
				result[p.Base()] = false
			}
			return true
		})

	// read the live version
	version := zk.GetString(this.conn, registry.NewPath(domain, service, "_live", name))
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
	domain := fmt.Sprintf("/%s.%s", domainInstance, domainClass)

	livepath := registry.NewPath(domain, service, "_live", name)
	version := zk.GetString(this.conn, livepath)
	if version != nil {
		buff, _, err := this.GetConfVersion(c, domainClass, domainInstance, service, name, *version)
		return buff, err
	}
	return nil, ErrNotFound
}
