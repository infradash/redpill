package redpill

import (
	"github.com/golang/glog"
	. "github.com/infradash/redpill/pkg/api"
	"github.com/qorio/omni/auth"
	"net/http"
)

func (this *Api) ListDomainPkgs(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	request := this.CreateServiceContext(context, req)

	domain_class := request.UrlParameter("domain_class")
	result, err := this.pkg.ListDomainPkgs(request, domain_class)
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

func (this *Api) CreatePkg(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	request := this.CreateServiceContext(context, req)

	model, err := this.pkg.NewPkgModel(request, req, this.engine.UnmarshalJSON)
	if err != nil {
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, "bad-json", http.StatusBadRequest)
		return
	}

	err = this.pkg.CreatePkg(request,
		request.UrlParameter("domain_class"),
		request.UrlParameter("domain_instance"),
		request.UrlParameter("service"),
		request.UrlParameter("version"),
		model)

	switch {
	case err == ErrConflict:
		this.engine.HandleError(resp, req, "already-exists", http.StatusConflict)
		return
	case err != nil:
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, "save-pkg-fails", http.StatusInternalServerError)
		return
	}
}

func (this *Api) UpdatePkg(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	request := this.CreateServiceContext(context, req)

	change, err := this.pkg.NewPkgModel(request, req, this.engine.UnmarshalJSON)
	if err != nil {
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, "bad-json", http.StatusBadRequest)
		return
	}

	err = this.pkg.UpdatePkg(request,
		request.UrlParameter("domain_class"),
		request.UrlParameter("domain_instance"),
		request.UrlParameter("service"),
		request.UrlParameter("version"),
		change)

	switch {
	case err == ErrNoChanges:
		this.engine.HandleError(resp, req, "", http.StatusNotModified)
		return
	case err == ErrNotFound:
		this.engine.HandleError(resp, req, "not-found", http.StatusNotFound)
		return
	case err == ErrConflict:
		this.engine.HandleError(resp, req, "version-conflict", http.StatusConflict)
		return
	case err != nil:
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, "save-pkg-fails", http.StatusInternalServerError)
		return
	}
}

func (this *Api) GetPkg(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	request := this.CreateServiceContext(context, req)

	model, err := this.pkg.GetPkg(request,
		request.UrlParameter("domain_class"),
		request.UrlParameter("domain_instance"),
		request.UrlParameter("service"),
		request.UrlParameter("version"))

	switch {
	case err == ErrNotFound:
		this.engine.HandleError(resp, req, err.Error(), http.StatusNotFound)
		return
	case err != nil:
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, "get-env-fails", http.StatusInternalServerError)
		return
	}
	err = this.engine.MarshalJSON(req, model, resp)
	if err != nil {
		this.engine.HandleError(resp, req, "malformed-result", http.StatusInternalServerError)
		return
	}
}

func (this *Api) DeletePkg(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	request := this.CreateServiceContext(context, req)

	err := this.pkg.DeletePkg(request,
		request.UrlParameter("domain_class"),
		request.UrlParameter("domain_instance"),
		request.UrlParameter("service"),
		request.UrlParameter("version"))

	switch {
	case err == ErrNoChanges:
		this.engine.HandleError(resp, req, "", http.StatusNotModified)
		return
	case err == ErrNotFound:
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

func (this *Api) SetPkgLiveVersion(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	request := this.CreateServiceContext(context, req)

	err := this.pkg.SetLive(request,
		request.UrlParameter("domain_class"),
		request.UrlParameter("domain_instance"),
		request.UrlParameter("service"),
		request.UrlParameter("version"))

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

func (this *Api) GetPkgLiveVersion(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	request := this.CreateServiceContext(context, req)

	model, err := this.pkg.GetPkgLiveVersion(request,
		request.UrlParameter("domain_class"),
		request.UrlParameter("domain_instance"),
		request.UrlParameter("service"))

	switch {
	case err == ErrNotFound:
		this.engine.HandleError(resp, req, err.Error(), http.StatusNotFound)
		return
	case err != nil:
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, "get-pkg-fails", http.StatusInternalServerError)
		return
	}
	err = this.engine.MarshalJSON(req, model, resp)
	if err != nil {
		this.engine.HandleError(resp, req, "malformed-result", http.StatusInternalServerError)
		return
	}
}

func (this *Api) ListPkgVersions(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	request := this.CreateServiceContext(context, req)

	model, err := this.pkg.ListPkgVersions(request,
		request.UrlParameter("domain_class"),
		request.UrlParameter("domain_instance"),
		request.UrlParameter("service"))

	switch {
	case err == ErrNotFound:
		this.engine.HandleError(resp, req, "not-found", http.StatusNotFound)
		return
	case err != nil:
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, "list-env-versions-fails", http.StatusInternalServerError)
		return
	}

	err = this.engine.MarshalJSON(req, model, resp)
	if err != nil {
		this.engine.HandleError(resp, req, "malformed-result", http.StatusInternalServerError)
		return
	}
}
