package redpill

import (
	"github.com/golang/glog"
	_ "github.com/infradash/redpill/pkg/api"
	"github.com/qorio/omni/auth"
	"net/http"
)

func (this *Api) CreateOrchestrationModel(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	c := this.CreateServiceContext(context, req)
	domain_class := c.UrlParameter("domain_class")
	glog.Infoln("DomainClass=", domain_class)

	model, err := this.orchestrate.NewOrchestrationModel(c, req, this.engine.UnmarshalJSON)
	if err != nil {
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, "bad-json", http.StatusBadRequest)
		return
	}
	err = this.orchestrate.SaveOrchestrationModel(c, domain_class, model)
	if err != nil {
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (this *Api) GetOrchestrationModel(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	c := this.CreateServiceContext(context, req)
	domain_class := c.UrlParameter("domain_class")
	orchestration := c.UrlParameter("orchestration")
	glog.Infoln("DomainClass=", domain_class, "Orchestration=", orchestration)

	model, err := this.orchestrate.GetOrchestrationModel(c, domain_class, orchestration)
	if err != nil {
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, err.Error(), http.StatusInternalServerError)
		return
	}
	err = this.engine.MarshalJSON(req, model, resp)
	if err != nil {
		this.engine.HandleError(resp, req, "malformed-result", http.StatusInternalServerError)
		return
	}
}

func (this *Api) DeleteOrchestrationModel(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	c := this.CreateServiceContext(context, req)
	domain_class := c.UrlParameter("domain_class")
	orchestration := c.UrlParameter("orchestration")
	glog.Infoln("DomainClass=", domain_class, "Orchestration=", orchestration)

	err := this.orchestrate.DeleteOrchestrationModel(c, domain_class, orchestration)
	if err != nil {
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}
