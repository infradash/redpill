package conf

import (
	"github.com/golang/glog"
	. "github.com/infradash/redpill/pkg/api"
)

type Service struct {
	storage ConfStorage
}

type conf struct {
	Domain      string `json:"domain"`
	Service     string `json:"service"`
	Name        string `json:"name"`
	Size        int    `json:"size"`
	ContentType string `json:"content_type"`
}

func (this conf) IsConf(other interface{}) bool {
	_, isa := other.(conf)
	return isa
}

func NewService(storage func() ConfStorage) ConfService {
	return &Service{
		storage: storage(),
	}
}

func (this *Service) SaveConf(c Context, domainClass, service, name string, buff []byte) error {
	glog.Infoln("Saving conf", "DomainClass=", domainClass, "Service=", service, "Name=", name, "Content=", string(buff))
	return this.storage.Save(domainClass, service, name, buff)
}

func (this *Service) ListConfs(c Context, domainClass, service string) ([]Conf, error) {
	glog.Infoln("Listing confs", "DomainClass=", domainClass, "Service=", service)
	keys, sizes, err := this.storage.ListAll(domainClass, service)
	if err != nil {
		return nil, err
	}
	confs := []Conf{}
	for i, k := range keys {
		c := conf{
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
