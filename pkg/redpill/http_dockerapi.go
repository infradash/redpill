package redpill

import (
	"github.com/golang/glog"
	"github.com/qorio/omni/auth"
	"net/http"
	"net/http/httputil"
	"net/url"
)

func (this *Api) DockerProxyReadonly(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	request := this.CreateServiceContext(context, req)
	domainClass := request.UrlParameter("domain_class")
	domainInstance := request.UrlParameter("domain_instance")
	target := request.UrlParameter("target")

	ep := "http://localhost:4040/dockerapi"
	strip := "/v1/dockerapi/" + domainClass + "/" + domainInstance + "/" + target

	glog.Infoln(">>>>>>>> DOCKER READONLY", domainClass, domainInstance, target, "strip=", strip)

	h := http.StripPrefix(strip, createTcpHandler(ep))
	h.ServeHTTP(resp, req)
}
func (this *Api) DockerProxyUpdate(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	request := this.CreateServiceContext(context, req)
	domainClass := request.UrlParameter("domain_class")
	domainInstance := request.UrlParameter("domain_instance")
	target := request.UrlParameter("target")
	glog.Infoln(">>>>>>>> DOCKER UPDATE", domainClass, domainInstance, target)
}

func createTcpHandler(e string) http.Handler {
	u, err := url.Parse(e)
	if err != nil {
		glog.Fatal(err)
	}
	return httputil.NewSingleHostReverseProxy(u)
}
