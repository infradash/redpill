package conf

import (
	"github.com/qorio/omni/common"
)

type confInfo struct {
	Domain      string `json:"domain"`
	Service     string `json:"service"`
	Name        string `json:"name"`
	Size        int    `json:"size,omitempty"`
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

type ConfStorage interface {
	Save(domainClass, service, name string, content []byte) error
	Get(domainClass, service, name string) ([]byte, error)
	Delete(domainClass, service, name string) error
	//	ListAll(domainClass, service string) ([]string, []int, error)

	SaveVersion(domainClass, domainInstance, service, version, name string, content []byte) error
	GetVersion(domainClass, domainInstance, service, version, name string) ([]byte, error)
	DeleteVersion(domainClass, domainInstance, service, version, name string) error
}
