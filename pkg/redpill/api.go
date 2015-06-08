package redpill

import (
	"github.com/qorio/omni/api"
	"github.com/qorio/omni/version"
	"net/http"
)

const (
	ScopeGetEnvironmentVars api.AuthScope = iota
	ScopeUpdateEnvironmentVars
	ScopeGetRegistry
	ScopeUpdateRegistry
)

var AuthScopes = api.AuthScopes{
	ScopeGetEnvironmentVars:    "get-env",
	ScopeUpdateEnvironmentVars: "update-env",
	ScopeGetRegistry:           "get-registry",
	ScopeUpdateRegistry:        "update-registry",
}

const (
	Info api.ServiceMethod = iota

	// Websocket test
	RunScript
	EventsFeed

	GetEnvironmentVars
	UpdateEnvironmentVars

	GetRegistry
	UpdateRegistry
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
		ResponseBody: func(req *http.Request) interface{} {
			return make([]string, 0)
		},
	},

	EventsFeed: api.MethodSpec{
		Doc: `
Main events feed
`,
		UrlRoute:   "/v1/events",
		HttpMethod: "GET",
		ResponseBody: func(req *http.Request) interface{} {
			return EventList{}
		},
	},

	GetEnvironmentVars: api.MethodSpec{
		AuthScope: AuthScopes[ScopeGetEnvironmentVars],
		Doc: `
Get environment variables
`,
		UrlRoute:   "/v1/{domain}/{service}/{version}/env",
		HttpMethod: "GET",
		ResponseBody: func(req *http.Request) interface{} {
			return EnvList{}
		},
	},

	UpdateEnvironmentVars: api.MethodSpec{
		AuthScope: AuthScopes[ScopeUpdateEnvironmentVars],
		Doc: `
Update environment variables
`,
		UrlRoute:     "/v1/{domain}/{service}/{version}/env",
		HttpMethod:   "POST",
		ContentTypes: []string{"application/json"},
		RequestBody: func(req *http.Request) interface{} {
			return new(EnvChange)
		},
	},

	GetRegistry: api.MethodSpec{
		AuthScope: AuthScopes[ScopeGetRegistry],
		Doc: `
Get registry key
`,
		UrlRoute:   "/v1/reg/{key}",
		HttpMethod: "GET",
		ResponseBody: func(req *http.Request) interface{} {
			return new(string)
		},
	},

	UpdateRegistry: api.MethodSpec{
		AuthScope: AuthScopes[ScopeUpdateRegistry],
		Doc: `
Update registry key
`,
		UrlRoute:     "/v1/reg/{key}",
		HttpMethod:   "POST",
		ContentTypes: []string{"application/json"},
		RequestBody: func(req *http.Request) interface{} {
			return new(string)
		},
	},
}

type EventList []Event
type Event struct {
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	User        string `json:"user,omitempty"`
	Type        string `json:"type,omitempty"`
	Url         string `json:"url,omitempty"`
	Timestamp   int64  `json:"timestamp,omitempty"`
}

type EnvList []Env
type Env struct {
	Name  string `json:"name"`
	Value string `json:"value,omitempty"`
}

type EnvChange struct {
	Update EnvList `json:"update,omitempty"`
	Delete EnvList `json:"delete,omitempty"`
}
