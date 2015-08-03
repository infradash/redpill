package domain

import (
	"github.com/golang/glog"
	. "github.com/infradash/redpill/pkg/api"
)

type domainService struct {
}

func NewService() DomainService {
	return &domainService{}
}

func (this *domainService) ListDomains(c Context) ([]Domain, error) {
	glog.Infoln("ListDomains", "UserId=", c.UserId())
	return nil, nil
}

func (this *domainService) GetDomain(c Context, domainClass string) (*DomainDetail, error) {
	glog.Infoln("GetDomain", "UserId=", c.UserId())
	return nil, nil
}
