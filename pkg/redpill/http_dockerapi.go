package redpill

import (
	"github.com/golang/glog"
	. "github.com/infradash/redpill/pkg/api"
	"github.com/qorio/omni/auth"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func (this *Api) ListDockerProxies(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	request := this.CreateServiceContext(context, req)
	domainClass := request.UrlParameter("domain_class")
	domainInstance := request.UrlParameter("domain_instance")

	glog.Infoln("ListDockerProxies:", "DomainClass=", domainClass, "DomainInstance=", domainInstance)
	result, err := this.dockerapi.ListProxies(request, domainClass, domainInstance)
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

func (this *Api) DockerProxyReadonly(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	request := this.CreateServiceContext(context, req)
	domainClass := request.UrlParameter("domain_class")
	domainInstance := request.UrlParameter("domain_instance")
	target := request.UrlParameter("target")

	proxy, err := this.dockerapi.GetProxy(request, domainClass, domainInstance, target)
	switch {
	case err == ErrNotFound:
		this.engine.HandleError(resp, req, "not-found", http.StatusNotFound)
		return
	case err != nil:
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, err.Error(), http.StatusInternalServerError)
		return
	}

	ep, err := proxy.GetApiProxy()
	switch {
	case err == ErrNoApiProxy:
		this.engine.HandleError(resp, req, "no-proxy", http.StatusNotFound)
		return
	case err != nil:
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, err.Error(), http.StatusInternalServerError)
		return
	}

	strip := "/v1/dockerapi/" + domainClass + "/" + domainInstance + "/" + target
	glog.Infoln("DockerProxyReadonly", domainClass, domainInstance, target, "strip=", strip, "endpoint=", ep)
	h := http.StripPrefix(strip, createTcpHandler(ep))
	h.ServeHTTP(resp, req)
}

func (this *Api) DockerProxyUpdate(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	request := this.CreateServiceContext(context, req)
	domainClass := request.UrlParameter("domain_class")
	domainInstance := request.UrlParameter("domain_instance")
	target := request.UrlParameter("target")

	proxy, err := this.dockerapi.GetProxy(request, domainClass, domainInstance, target)
	switch {
	case err == ErrNotFound:
		this.engine.HandleError(resp, req, "not-found", http.StatusNotFound)
		return
	case err != nil:
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, err.Error(), http.StatusInternalServerError)
		return
	}

	ep, err := proxy.GetApiProxy()
	switch {
	case err == ErrNoApiProxy:
		this.engine.HandleError(resp, req, "no-proxy", http.StatusNotFound)
		return
	case err != nil:
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, err.Error(), http.StatusInternalServerError)
		return
	}

	strip := "/v1/dockerapi/" + domainClass + "/" + domainInstance + "/" + target
	glog.Infoln("DockerProxyUpdate", domainClass, domainInstance, target, "strip=", strip, "endpoint=", ep)
	h := http.StripPrefix(strip, createTcpHandler(ep))
	h.ServeHTTP(resp, req)
}

func createTcpHandler(e string) http.Handler {
	u, err := url.Parse(e)
	if err != nil {
		glog.Fatal(err)
	}
	return httputil.NewSingleHostReverseProxy(u)
}
