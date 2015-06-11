package redpill

import (
	"github.com/golang/glog"
	"github.com/qorio/omni/auth"
	"net/http"
)

func (this *Api) ListDomains(ac auth.Context, resp http.ResponseWriter, req *http.Request) {
	context := this.Wrap(ac, req)
	userId := context.UserId()

	glog.Infoln("ListDomains", "UserId=", userId)
	list, err := this.service.DomainService(context).ListDomains()
	if err != nil {
		this.engine.HandleError(resp, req, "list-domain-error", http.StatusInternalServerError)
		return
	}
	err = this.engine.MarshalJSON(req, list, resp)
	if err != nil {
		this.engine.HandleError(resp, req, "malformed-result", http.StatusInternalServerError)
		return
	}
}

func (this *Api) GetDomain(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	request := this.Wrap(context, req)
	userId := request.UserId()

	glog.Infoln("GetDomain", "UserId=", userId)
}

type domainService struct {
	context *Context
}

func (this *domainService) ListDomains() ([]Domain, error) {
	glog.Infoln("ListDomains", "UserId=", this.context.UserId())

	return []Domain{

		Domain{
			Id:   "ops-test.blinker.com",
			Name: "ops-test",
			Url:  "/v1/ops-test.blinker.com",
		},
	}, nil
}

func (this *domainService) GetDomain(domain string) (DomainDetail, error) {
	glog.Infoln("GetDomain", "UserId=", this.context.UserId())

	return DomainDetail{}, nil
}
