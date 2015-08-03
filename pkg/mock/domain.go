package mock

import (
	"github.com/golang/glog"
	. "github.com/infradash/redpill/pkg/api"
)

type domainService struct {
}

func NewDomainService() DomainService {
	return &domainService{}
}

func (this *domainService) ListDomains(c Context) ([]Domain, error) {
	glog.Infoln("ListDomains", "UserId=", c.UserId())
	return []Domain{
		Domain{
			Id:    "blinker.com",
			Class: "blinker.com",
			Name:  "API",
			Url:   "/v1/domain/blinker.com",
		},
	}, nil
}

func (this *domainService) GetDomain(c Context, domainClass string) (*DomainDetail, error) {
	glog.Infoln("GetDomain", "UserId=", c.UserId())
	if domainClass == "blinker.com" {
		return &DomainDetail{
			Id:    "blinker.com",
			Class: "blinker.com",
			Name:  "API",
			Instances: []string{
				"dev",
				"staging",
				"production",
			},
		}, nil
	}
	return nil, nil
}
