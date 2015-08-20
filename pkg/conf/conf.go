package conf

import (
	"fmt"
	"github.com/golang/glog"
	. "github.com/infradash/redpill/pkg/api"
	"github.com/qorio/maestro/pkg/zk"
	"github.com/qorio/omni/common"
	"strings"
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
	_, isa := other.(confInfo)
	return isa
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

func (this *Service) SaveConf(c Context, domainClass, service, name string, buff []byte) error {
	glog.Infoln("Saving conf", "DomainClass=", domainClass, "Service=", service, "Name=", name, "Content=", string(buff))
	return this.storage.Save(domainClass, service, name, buff)
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
				if zversion.GetBasename() == "live" {
					zobjects, err := zversion.Children()
					if err != nil {
						return nil, err
					}
					for _, zobject := range zobjects {
						// get the current live information
						if strings.Index(zobject.GetBasename(), "_") == -1 {
							service_stats[service].set_live(domainInstance, zobject.GetBasename(), zobject.GetValueString())
						}
					}

				} else {
					// a version
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

func (this *Service) GetConf(c Context, domainClass, service, name string) ([]byte, error) {
	glog.Infoln("GetConf DomainClass=", domainClass, "Service=", service, "Name=", name)
	return this.storage.Get(domainClass, service, name)
}

func (this *Service) DeleteConf(c Context, domainClass, service, name string) error {
	glog.Infoln("DeleteConf DomainClass=", domainClass, "Service=", service, "Name=", name)
	return this.storage.Delete(domainClass, service, name)
}

func (this *Service) SaveConfVersion(c Context,
	domainClass, domainInstance, service, name, version string,
	buff []byte) error {
	glog.Infoln("SaveConfVersion DomainClass=", domainClass, "Service=", service, "Name=", name,
		"DomainInstance=", domainInstance, "Version=", version)
	return this.storage.SaveVersion(domainClass, domainInstance, service, name, version, buff)
}
func (this *Service) GetConfVersion(c Context, domainClass, domainInstance, service, name, version string) ([]byte, error) {
	glog.Infoln("GetConfVersion DomainClass=", domainClass, "Service=", service, "Name=", name,
		"DomainInstance=", domainInstance, "Version=", version)
	return this.storage.GetVersion(domainClass, domainInstance, service, name, version)

}
func (this *Service) DeleteConfVersion(c Context,
	domainClass, domainInstance, service, name, version string) error {
	glog.Infoln("DeleteConfVersion DomainClass=", domainClass, "Service=", service, "Name=", name,
		"DomainInstance=", domainInstance, "Version=", version)
	return this.storage.DeleteVersion(domainClass, domainInstance, service, name, version)
}
