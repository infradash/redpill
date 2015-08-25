package domain

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/golang/glog"
	. "github.com/infradash/redpill/pkg/api"
	"github.com/qorio/maestro/pkg/registry"
	"github.com/qorio/maestro/pkg/zk"
	"net/http"
	"path"
)

const (
	RedpillNamespace = "/_redpill"
)

var (
	ErrBadModel      = errors.New("error-bad-model")
	ErrAlreadyExists = errors.New("error-exists")
	ErrNotExists     = errors.New("error-not-exists")
	ErrBadInput      = errors.New("error-bad-input")
)

type Service struct {
	conn zk.ZK
}

func NewService(pool func() zk.ZK) DomainService {
	s := new(Service)
	s.conn = pool()
	return s
}

func (this *Service) NewDomainModel(c Context, req *http.Request, um Unmarshaler) (DomainModel, error) {
	m := new(Model)
	err := um(req, m)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func (this *Service) ListDomains(c Context) ([]DomainInfo, error) {
	glog.Infoln("ListDomains", "UserId=", c.UserId())

	zdomains, err := this.conn.Get(path.Join(RedpillNamespace, "domain"))
	if err != nil {
		return nil, err
	}
	list, err := zdomains.Children()
	if err != nil {
		return nil, err
	}
	results := []DomainInfo{}
	for _, zdomain := range list {
		info := new(Info)
		err = json.Unmarshal(zdomain.GetValue(), info)
		if err != nil {
			return nil, err
		}
		results = append(results, info)
	}
	return results, nil
}

func (this *Service) CreateDomain(c Context, model DomainModel) error {
	if !(new(Model)).IsDomainModel(model) {
		return ErrBadModel
	}
	p := registry.Path(path.Join(RedpillNamespace, "domain", model.DomainClass()))
	if zk.PathExists(this.conn, p) {
		return ErrAlreadyExists
	}
	return zk.CreateOrSet(this.conn, p, model)
}

func (this *Service) UpdateDomain(c Context, domainClass string, model DomainModel) error {
	if !(new(Model)).IsDomainModel(model) {
		return ErrBadModel
	}
	if domainClass != model.DomainClass() {
		return ErrBadInput
	}
	p := registry.Path(path.Join(RedpillNamespace, "domain", model.DomainClass()))
	if !zk.PathExists(this.conn, p) {
		return ErrNotExists
	}
	return zk.CreateOrSet(this.conn, p, model)
}

func (this *Service) GetDomain(c Context, domainClass string) (DomainModel, error) {
	glog.Infoln("GetDomain", "UserId=", c.UserId())
	model := new(Model)
	err := zk.GetObject(this.conn, registry.Path(path.Join(RedpillNamespace, "domain", domainClass)), model)
	if err != nil {
		return model, err
	}

	// Given the model, we need to find out the services.
	services := map[string]int{}
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
			services[service] += 1
		}
	}

	model.Services = []string{}
	for s, _ := range services {
		model.Services = append(model.Services, s)
	}
	return model, nil
}
