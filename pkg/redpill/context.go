package redpill

import (
	. "github.com/infradash/redpill/pkg/api"
	"github.com/qorio/omni/auth"
	"github.com/qorio/omni/rest"
	"net/http"
)

type context struct {
	userId       func() string
	urlParameter func(string) string
	context      auth.Context
	request      *http.Request
}

func (c *context) UserId() string {
	return c.userId()
}

func (c *context) UrlParameter(k string) string {
	return c.UrlParameter(k)
}

func ServiceContext(engine rest.Engine) func(c auth.Context, req *http.Request) Context {
	return func(c auth.Context, req *http.Request) Context {
		ctx := &context{context: c, request: req}
		ctx.userId = func() string {
			return c.GetStringForService(ServiceId, "@id")
		}
		ctx.urlParameter = func(k string) string {
			return engine.GetUrlParameter(req, k)
		}
		return ctx
	}
}
