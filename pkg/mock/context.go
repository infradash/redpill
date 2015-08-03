package mock

import (
	"flag"
	. "github.com/infradash/redpill/pkg/api"
	"github.com/qorio/omni/auth"
	"github.com/qorio/omni/rest"
	"net/http"
)

var (
	Mock = flag.Bool("mock", true, "True to mock data.")
)

func AuthContext(authed bool, context auth.Context) (bool, auth.Context) {
	return true, nil
}

func ServiceContext(engine rest.Engine) func(c auth.Context, req *http.Request) Context {
	return func(c auth.Context, req *http.Request) Context {
		ctx := &context{context: c, request: req}
		ctx.userId = func() string {
			return "Dash"
		}
		ctx.urlParameter = func(k string) string {
			return engine.GetUrlParameter(req, k)
		}
		return ctx
	}
}

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
	return c.urlParameter(k)
}
