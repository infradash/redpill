package redpill

import (
	"github.com/golang/glog"
	_ "github.com/infradash/redpill/pkg/api"
	"github.com/qorio/omni/auth"
	"io/ioutil"
	"net/http"
)

/// Lists confs by domain -- metadata for domain instances, versions, live versions, etc.
func (this *Api) ListDomainConfs(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	request := this.CreateServiceContext(context, req)

	domain_class := request.UrlParameter("domain_class")
	result, err := this.conf.ListDomainConfs(request, domain_class)
	if err != nil {
		this.engine.HandleError(resp, req, "query-failed", http.StatusInternalServerError)
		return
	}
	err = this.engine.MarshalJSON(req, result, resp)
	if err != nil {
		this.engine.HandleError(resp, req, "malformed-result", http.StatusInternalServerError)
		return
	}
}

func (this *Api) ListConfFiles(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	c := this.CreateServiceContext(context, req)
	domain_class := c.UrlParameter("domain_class")
	service := c.UrlParameter("service")

	result, err := this.conf.ListConfs(c, domain_class, service)
	if err != nil {
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, err.Error(), http.StatusInternalServerError)
		return
	}

	err = this.engine.MarshalJSON(req, result, resp)
	if err != nil {
		this.engine.HandleError(resp, req, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (this *Api) CreateConfFile(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	c := this.CreateServiceContext(context, req)
	domain_class := c.UrlParameter("domain_class")
	service := c.UrlParameter("service")
	name := c.UrlParameter("name")
	glog.Infoln("DomainClass=", domain_class, "Service=", service, "Name=", name)

	defer req.Body.Close()
	buff, err := ioutil.ReadAll(req.Body)
	if err != nil {
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, "bad-json", http.StatusBadRequest)
		return
	}

	err = this.conf.SaveConf(c, domain_class, service, name, buff)
	if err != nil {
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (this *Api) CreateConfFileVersion(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	c := this.CreateServiceContext(context, req)
	domain_class := c.UrlParameter("domain_class")
	domain_instance := c.UrlParameter("domain_instance")
	service := c.UrlParameter("service")
	version := c.UrlParameter("version")
	name := c.UrlParameter("name")
	glog.Infoln("DomainClass=", domain_class, "Service=", service, "Name=", name,
		"DomainInstance=", domain_instance, "Version=", version)

	defer req.Body.Close()
	buff, err := ioutil.ReadAll(req.Body)
	if err != nil {
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, "bad-json", http.StatusBadRequest)
		return
	}

	err = this.conf.SaveConfVersion(c, domain_class, domain_instance, service, name, version, buff)
	if err != nil {
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (this *Api) GetConfFile(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	c := this.CreateServiceContext(context, req)
	domain_class := c.UrlParameter("domain_class")
	service := c.UrlParameter("service")
	name := c.UrlParameter("name")
	glog.Infoln("DomainClass=", domain_class, "Service=", service, "Name=", name)

	buff, err := this.conf.GetConf(c, domain_class, service, name)
	if err != nil {
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(buff) > 0 {
		resp.Header().Add("Content-Type", "text/plain")
		resp.Write(buff)
	} else {
		this.engine.HandleError(resp, req, "not-found", http.StatusNotFound)
	}
	return
}

func (this *Api) GetConfFileVersion(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	c := this.CreateServiceContext(context, req)
	domain_class := c.UrlParameter("domain_class")
	domain_instance := c.UrlParameter("domain_instance")
	service := c.UrlParameter("service")
	version := c.UrlParameter("version")
	name := c.UrlParameter("name")
	glog.Infoln("DomainClass=", domain_class, "Service=", service, "Name=", name,
		"DomainInstance=", domain_instance, "Version=", version)

	buff, err := this.conf.GetConfVersion(c, domain_class, domain_instance, service, name, version)
	if err != nil {
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(buff) > 0 {
		resp.Header().Add("Content-Type", "text/plain")
		resp.Write(buff)
	} else {
		this.engine.HandleError(resp, req, "not-found", http.StatusNotFound)
	}
	return
}

func (this *Api) DeleteConfFile(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	c := this.CreateServiceContext(context, req)
	domain_class := c.UrlParameter("domain_class")
	service := c.UrlParameter("service")
	name := c.UrlParameter("name")
	glog.Infoln("DomainClass=", domain_class, "Service=", service, "Name=", name)

	err := this.conf.DeleteConf(c, domain_class, service, name)
	if err != nil {
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}

func (this *Api) DeleteConfFileVersion(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	c := this.CreateServiceContext(context, req)
	domain_class := c.UrlParameter("domain_class")
	domain_instance := c.UrlParameter("domain_instance")
	service := c.UrlParameter("service")
	version := c.UrlParameter("version")
	name := c.UrlParameter("name")
	glog.Infoln("DomainClass=", domain_class, "Service=", service, "Name=", name,
		"DomainInstance=", domain_instance, "Version=", version)

	err := this.conf.DeleteConfVersion(c, domain_class, domain_instance, service, name, version)
	if err != nil {
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}
