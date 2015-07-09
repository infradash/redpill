package mock

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

	return []Domain{

		Domain{
			Id:   "ops-test.blinker.com",
			Name: "ops-test",
			Url:  "/v1/ops-test.blinker.com",
		},
	}, nil
}

func (this *domainService) GetDomain(c Context, domain string) (DomainDetail, error) {
	glog.Infoln("GetDomain", "UserId=", c.UserId())

	return DomainDetail{}, nil
}
