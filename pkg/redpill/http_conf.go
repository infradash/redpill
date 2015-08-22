package redpill

import (
	"fmt"
	"github.com/golang/glog"
	. "github.com/infradash/redpill/pkg/api"
	"github.com/qorio/omni/auth"
	"io/ioutil"
	"net/http"
	"strconv"
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

	// TODO - FIX ME
	err = this.conf.SaveConf(c, domain_class, service, name, buff, Revision(1))
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

	// TODO - FIX ME
	err = this.conf.SaveConfVersion(c, domain_class, domain_instance, service, name, version, buff, Revision(1))
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

	buff, rev, err := this.conf.GetConf(c, domain_class, service, name)
	if err != nil {
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, err.Error(), http.StatusInternalServerError)
		return
	}
	if len(buff) > 0 {
		resp.Header().Add("Content-Type", "text/plain")
		resp.Header().Set("X-Dash-Version", fmt.Sprintf("%d", rev))
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

	buff, rev, err := this.conf.GetConfVersion(c, domain_class, domain_instance, service, name, version)
	switch {
	case err == ErrNotFound:
		this.engine.HandleError(resp, req, err.Error(), http.StatusNotFound)
		return
	case err != nil:
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

	resp.Header().Set("X-Dash-Version", fmt.Sprintf("%d", rev))
}

func (this *Api) DeleteConfFile(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	c := this.CreateServiceContext(context, req)
	domain_class := c.UrlParameter("domain_class")
	service := c.UrlParameter("service")
	name := c.UrlParameter("name")

	version_header := req.Header.Get("X-Dash-Version")
	if version_header == "" {
		this.engine.HandleError(resp, req, "missing-header-X-Dash-Version", http.StatusBadRequest)
		return
	}
	rev, err := strconv.Atoi(version_header)
	if err != nil {
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, "bad-value-X-Dash-Version", http.StatusBadRequest)
		return
	}

	glog.Infoln("DomainClass=", domain_class, "Service=", service, "Name=", name, "Rev=", rev)

	err = this.conf.DeleteConf(c, domain_class, service, name, Revision(rev))
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

	version_header := req.Header.Get("X-Dash-Version")
	if version_header == "" {
		this.engine.HandleError(resp, req, "missing-header-X-Dash-Version", http.StatusBadRequest)
		return
	}
	rev, err := strconv.Atoi(version_header)
	if err != nil {
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, "bad-value-X-Dash-Version", http.StatusBadRequest)
		return
	}

	glog.Infoln("DomainClass=", domain_class, "Service=", service, "Name=", name,
		"DomainInstance=", domain_instance, "Version=", version, "Rev=", rev)

	err = this.conf.DeleteConfVersion(c, domain_class, domain_instance, service, name, version, Revision(rev))
	if err != nil {
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, err.Error(), http.StatusInternalServerError)
		return
	}
	return
}

func (this *Api) SetConfLiveVersion(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	request := this.CreateServiceContext(context, req)

	domainClass := request.UrlParameter("domain_class")
	domainInstance := request.UrlParameter("domain_instance")
	service := request.UrlParameter("service")
	version := request.UrlParameter("version")
	name := request.UrlParameter("name")

	glog.Infoln("SetConfLiveVersion:", "DomainClass=", domainClass, "DomainInstance=", domainInstance,
		"Service=", service, "Version=", version, "Name=", name)

	err := this.conf.SetLive(request, domainClass, domainInstance, service, version, name)

	switch {
	case err == ErrNotFound:
		this.engine.HandleError(resp, req, "not-found", http.StatusNotFound)
		return
	case err != nil:
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, "setlive-fails", http.StatusInternalServerError)
		return
	}
}

func (this *Api) ListConfVersions(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	request := this.CreateServiceContext(context, req)

	domainClass := request.UrlParameter("domain_class")
	domainInstance := request.UrlParameter("domain_instance")
	service := request.UrlParameter("service")
	name := request.UrlParameter("name")

	glog.Infoln("SetConfLiveVersion:", "DomainClass=", domainClass, "DomainInstance=", domainInstance,
		"Service=", service, "Name=", name)

	confVersions, err := this.conf.ListConfVersions(request, domainClass, domainInstance, service, name)

	switch {
	case confVersions == nil:
		this.engine.HandleError(resp, req, "not-found", http.StatusNotFound)
		return
	case err != nil:
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, "list-conf-versions-fails", http.StatusInternalServerError)
		return
	}

	err = this.engine.MarshalJSON(req, confVersions, resp)
	if err != nil {
		this.engine.HandleError(resp, req, "malformed-result", http.StatusInternalServerError)
		return
	}
}

func (this *Api) GetConfLiveVersion(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	request := this.CreateServiceContext(context, req)

	domainClass := request.UrlParameter("domain_class")
	domainInstance := request.UrlParameter("domain_instance")
	service := request.UrlParameter("service")
	name := request.UrlParameter("name")

	glog.Infoln("GetConfLiveVersion:", "DomainClass=", domainClass, "DomainInstance=", domainInstance,
		"Service=", service, "Name=", name)

	buff, err := this.conf.GetConfLiveVersion(request, domainClass, domainInstance, service, name)
	switch {
	case err == ErrNotFound:
		this.engine.HandleError(resp, req, err.Error(), http.StatusNotFound)
		return
	case err != nil:
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
}
