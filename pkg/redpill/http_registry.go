package redpill

import (
	"fmt"
	"github.com/golang/glog"
	. "github.com/infradash/redpill/pkg/api"
	"github.com/qorio/omni/auth"
	"net/http"
	"strconv"
)

func (this *Api) GetRegistryEntry(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	c := this.CreateServiceContext(context, req)

	result := Methods[GetRegistryEntry].ResponseBody(req).(*RegistryEntry)
	result.Path = "/" + c.UrlParameter("path")

	glog.Infoln("GetRegistry", "path=", result.Path)

	v, rev, err := this.registry.GetEntry(c, result.Path)
	switch {
	case err == ErrNotFound:
		this.engine.HandleError(resp, req, "not-found", http.StatusNotFound)
		return
	case err != nil:
		this.engine.HandleError(resp, req, err.Error(), http.StatusInternalServerError)
		return
	}

	resp.Header().Set("X-Dash-Version", fmt.Sprintf("%d", rev))
	result.Value = string(v)

	err = this.engine.MarshalJSON(req, result, resp)
	if err != nil {
		this.engine.HandleError(resp, req, "malformed-result", http.StatusInternalServerError)
		return
	}
}

func (this *Api) UpdateRegistryEntry(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	change := Methods[UpdateRegistryEntry].RequestBody(req).(*RegistryEntry)
	err := this.engine.UnmarshalJSON(req, change)
	if err != nil {
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, "bad-json", http.StatusBadRequest)
		return
	}

	c := this.CreateServiceContext(context, req)
	path := "/" + c.UrlParameter("path")

	if change.Path != path {
		this.engine.HandleError(resp, req, "conflict", http.StatusBadRequest)
		return
	}

	rev, err := strconv.Atoi(req.Header.Get("X-Dash-Version"))
	if err != nil {
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, "bad-version", http.StatusBadRequest)
		return
	}
	value := []byte(change.Value)
	new_rev, err := this.registry.UpdateEntry(c, path, value, Revision(rev))
	switch {
	case err == ErrNotFound:
		this.engine.HandleError(resp, req, "not-found", http.StatusNotFound)
		return
	case err != nil:
		this.engine.HandleError(resp, req, err.Error(), http.StatusInternalServerError)
		return
	}
	resp.Header().Set("X-Dash-Version", fmt.Sprintf("%d", new_rev))
}

func (this *Api) DeleteRegistryEntry(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	c := this.CreateServiceContext(context, req)
	path := "/" + c.UrlParameter("path")
	rev, err := strconv.Atoi(req.Header.Get("X-Dash-Version"))
	if err != nil {
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, "bad-version", http.StatusBadRequest)
		return
	}
	err = this.registry.DeleteEntry(c, path, Revision(rev))
	switch {
	case err == ErrNotFound:
		this.engine.HandleError(resp, req, "not-found", http.StatusNotFound)
		return
	case err != nil:
		this.engine.HandleError(resp, req, err.Error(), http.StatusInternalServerError)
		return
	}
}
