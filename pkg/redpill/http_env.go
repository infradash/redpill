package redpill

import (
	"fmt"
	"github.com/golang/glog"
	. "github.com/infradash/redpill/pkg/api"
	"github.com/infradash/redpill/pkg/env"
	"github.com/qorio/omni/auth"
	"net/http"
	"strconv"
)

func (this *Api) ListDomainEnvs(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	request := this.CreateServiceContext(context, req)

	domain_class := request.UrlParameter("domain_class")
	result, err := this.env.ListDomainEnvs(request, domain_class)
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

func (this *Api) GetEnvironmentVars(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	request := this.CreateServiceContext(context, req)

	vars, rev, err := this.env.GetEnv(request,
		fmt.Sprintf("%s.%s", request.UrlParameter("domain_instance"), request.UrlParameter("domain_class")),
		request.UrlParameter("service"),
		request.UrlParameter("version"))

	switch {
	case err == env.ErrNoEnv:
		this.engine.HandleError(resp, req, err.Error(), http.StatusNotFound)
		return
	case err != nil:
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, "get-env-fails", http.StatusInternalServerError)
		return
	}
	err = this.engine.MarshalJSON(req, vars, resp)
	if err != nil {
		this.engine.HandleError(resp, req, "malformed-result", http.StatusInternalServerError)
		return
	}
	resp.Header().Set("X-Dash-Version", fmt.Sprintf("%d", rev))
}

func (this *Api) CreateEnvironmentVars(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	request := this.CreateServiceContext(context, req)

	vars := Methods[CreateEnvironmentVars].RequestBody(req).(*EnvList)
	err := this.engine.UnmarshalJSON(req, vars)
	if err != nil {
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, "bad-json", http.StatusBadRequest)
		return
	}

	rev, err := this.env.NewEnv(request,
		fmt.Sprintf("%s.%s", request.UrlParameter("domain_instance"), request.UrlParameter("domain_class")),
		request.UrlParameter("service"),
		request.UrlParameter("version"),
		vars)

	switch {
	case err == ErrConflict:
		this.engine.HandleError(resp, req, "version-conflict", http.StatusConflict)
		return
	case err != nil:
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, "save-env-fails", http.StatusInternalServerError)
		return
	}
	resp.Header().Set("X-Dash-Version", fmt.Sprintf("%d", rev))
}

func (this *Api) UpdateEnvironmentVars(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	request := this.CreateServiceContext(context, req)

	change := Methods[UpdateEnvironmentVars].RequestBody(req).(*EnvChange)
	err := this.engine.UnmarshalJSON(req, change)
	if err != nil {
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, "bad-json", http.StatusBadRequest)
		return
	}

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

	_, err = this.env.SaveEnv(request,
		fmt.Sprintf("%s.%s", request.UrlParameter("domain_instance"), request.UrlParameter("domain_class")),
		request.UrlParameter("service"),
		request.UrlParameter("version"),
		change,
		Revision(rev))

	switch {
	case err == env.ErrNoChanges:
		this.engine.HandleError(resp, req, "", http.StatusNotModified)
		return
	case err == env.ErrBadVarName:
		this.engine.HandleError(resp, req, "err-bad-input", http.StatusBadRequest)
		return
	case err == env.ErrNoEnv:
		this.engine.HandleError(resp, req, "not-found", http.StatusNotFound)
		return
	case err == ErrConflict:
		this.engine.HandleError(resp, req, "version-conflict", http.StatusConflict)
		return
	case err != nil:
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, "save-env-fails", http.StatusInternalServerError)
		return
	}
}

func (this *Api) UpdateLiveVersionEnv(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	request := this.CreateServiceContext(context, req)

	err := this.env.SetLive(request,
		fmt.Sprintf("%s.%s", request.UrlParameter("domain_instance"), request.UrlParameter("domain_class")),
		request.UrlParameter("service"),
		request.UrlParameter("version"))

	switch {
	case err == env.ErrNoEnv:
		this.engine.HandleError(resp, req, "not-found", http.StatusNotFound)
		return
	case err != nil:
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, "setlive-fails", http.StatusInternalServerError)
		return
	}
}

func (this *Api) ListEnvVersions(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	request := this.CreateServiceContext(context, req)

	envVersions, err := this.env.ListEnvVersions(request,
		fmt.Sprintf("%s.%s", request.UrlParameter("domain_instance"), request.UrlParameter("domain_class")),
		request.UrlParameter("service"))

	switch {
	case err == env.ErrNoEnv:
		this.engine.HandleError(resp, req, "not-found", http.StatusNotFound)
		return
	case err != nil:
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, "list-env-versions-fails", http.StatusInternalServerError)
		return
	}

	err = this.engine.MarshalJSON(req, envVersions, resp)
	if err != nil {
		this.engine.HandleError(resp, req, "malformed-result", http.StatusInternalServerError)
		return
	}
}

func (this *Api) GetEnvLiveVersion(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	request := this.CreateServiceContext(context, req)

	vars, err := this.env.GetEnvLiveVersion(request,
		fmt.Sprintf("%s.%s", request.UrlParameter("domain_instance"), request.UrlParameter("domain_class")),
		request.UrlParameter("service"))

	switch {
	case err == env.ErrNoEnv:
		this.engine.HandleError(resp, req, err.Error(), http.StatusNotFound)
		return
	case err != nil:
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, "get-env-fails", http.StatusInternalServerError)
		return
	}
	err = this.engine.MarshalJSON(req, vars, resp)
	if err != nil {
		this.engine.HandleError(resp, req, "malformed-result", http.StatusInternalServerError)
		return
	}
}
