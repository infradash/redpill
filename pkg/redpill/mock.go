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

type Context struct {
	UserId       func() string
	UrlParameter func(string) string

	context auth.Context
	request *http.Request
}

func (this *Api) Wrap(context auth.Context, req *http.Request) *Context {

	if *Mock {
		return &Context{
			context: context,
			request: req,

			UserId: func() string {
				return "Dash"
			},

			UrlParameter: func(k string) string {
				switch k {
				// we are always overriding this for mock data
				case "domain":
					return "ops-test.blinker.com"
				default:
					return this.engine.GetUrlParameter(req, k)
				}
			},
		}
	}

	// The real thing
	return &Context{
		context: context,
		request: req,

		UserId: func() string {
			return context.GetStringForService(ServiceId, "@id")
		},
		UrlParameter: func(k string) string {
			return this.engine.GetUrlParameter(req, k)
		},
	}
}
