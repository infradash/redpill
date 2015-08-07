package redpill

import (
	"github.com/golang/glog"
	_ "github.com/infradash/redpill/pkg/api"
	"github.com/qorio/omni/auth"
	"io/ioutil"
	"net/http"
)

func (this *Api) CreateConfigFileBase(context auth.Context, resp http.ResponseWriter, req *http.Request) {
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

	glog.Infoln(string(buff))
}
