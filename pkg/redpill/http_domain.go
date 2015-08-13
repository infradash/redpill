package redpill

import (
	"github.com/golang/glog"
	"github.com/infradash/redpill/pkg/domain"
	"github.com/qorio/omni/auth"
	"net/http"
)

func (this *Api) ListDomains(ac auth.Context, resp http.ResponseWriter, req *http.Request) {
	context := this.CreateServiceContext(ac, req)
	userId := context.UserId()

	glog.Infoln("ListDomains", "UserId=", userId)
	list, err := this.domain.ListDomains(context)
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

func (this *Api) CreateDomain(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	request := this.CreateServiceContext(context, req)
	userId := request.UserId()
	glog.Infoln("CreateDomain", "UserId=", userId)
	model, err := this.domain.NewDomainModel(request, req, this.engine.UnmarshalJSON)
	if err != nil {
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, "bad-json", http.StatusBadRequest)
		return
	}

	err = this.domain.CreateDomain(request, model)
	switch {
	case err == domain.ErrAlreadyExists:
		this.engine.HandleError(resp, req, err.Error(), http.StatusConflict)
		return
	case err != nil:
		this.engine.HandleError(resp, req, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}

func (this *Api) UpdateDomain(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	request := this.CreateServiceContext(context, req)
	userId := request.UserId()
	domainClass := request.UrlParameter("domain_class")

	glog.Infoln("CreateDomain", "UserId=", userId)
	model, err := this.domain.NewDomainModel(request, req, this.engine.UnmarshalJSON)
	if err != nil {
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, "bad-json", http.StatusBadRequest)
		return
	}
	err = this.domain.UpdateDomain(request, domainClass, model)
	switch {
	case err == domain.ErrNotExists:
		this.engine.HandleError(resp, req, err.Error(), http.StatusBadRequest)
		return
	case err != nil:
		this.engine.HandleError(resp, req, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}

func (this *Api) GetDomain(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	request := this.CreateServiceContext(context, req)
	userId := request.UserId()

	glog.Infoln("GetDomain", "UserId=", userId)
	domain_class := request.UrlParameter("domain_class")
	detail, err := this.domain.GetDomain(request, domain_class)
	if err != nil {
		this.engine.HandleError(resp, req, "cannot-get-domain", http.StatusInternalServerError)
		return
	}
	err = this.engine.MarshalJSON(req, detail, resp)
	if err != nil {
		this.engine.HandleError(resp, req, "malformed-result", http.StatusInternalServerError)
		return
	}
}
