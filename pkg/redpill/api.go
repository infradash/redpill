package redpill

import (
	"github.com/qorio/omni/api"
	"github.com/qorio/omni/version"
	"net/http"
)

const (
	ReadOnlyScope api.AuthScope = iota
)

var AuthScopes = api.AuthScopes{
	ReadOnlyScope: "readonly",
}

const (
	Info api.ServiceMethod = iota

	// Websocket test
	RunScript
	EventsFeed
)

var Methods = api.ServiceMethods{

	Info: api.MethodSpec{
		Doc: `
Returns build info
`,
		UrlRoute:     "/info",
		HttpMethod:   "GET",
		ContentTypes: []string{"application/json"},
		ResponseBody: func(req *http.Request) interface{} {
			return version.Build{}
		},
	},

	RunScript: api.MethodSpec{
		Doc: `
Websocket run a script
`,
		UrlRoute:   "/v1/ws/run/{script}",
		HttpMethod: "GET",
	},

	EventsFeed: api.MethodSpec{
		Doc: `
Main events feed
`,
		UrlRoute:   "/v1/events",
		HttpMethod: "GET",
	},
}
