package redpill

import (
	"flag"
	"github.com/qorio/omni/auth"
	"net/http"
)

var (
	Mock = flag.Bool("mock", true, "True to mock data.")
)

func MockAuthContext(authed bool, context auth.Context) (bool, auth.Context) {
	return true, nil
}

type context struct {
	userId       func() string
	UrlParameter func(string) string
	context      auth.Context
	request      *http.Request
}

func (c *context) UserId() string {
	return c.userId()
}

func (this *Api) Wrap(c auth.Context, req *http.Request) *context {

	ctx := &context{context: c, request: req}

	if *Mock {
		ctx.userId = func() string {
			return "Dash"
		}
		ctx.UrlParameter = func(k string) string {
			switch k {
			// we are always overriding this for mock data
			case "domain":
				return "ops-test.blinker.com"
			default:
				return this.engine.GetUrlParameter(req, k)
			}
		}

	} else {
		ctx.userId = func() string {
			return c.GetStringForService(ServiceId, "@id")
		}
		ctx.UrlParameter = func(k string) string {
			return this.engine.GetUrlParameter(req, k)
		}
	}

	return ctx
}
