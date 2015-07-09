package api

import (
	"github.com/qorio/omni/api"
	"github.com/qorio/omni/version"
	"net/http"
)

const (
	ScopeEnvironmentReadonly api.AuthScope = iota
	ScopeEnvironmentUpdate
	ScopeEnvironmentAdmin

	ScopeRegistryReadonly
	ScopeRegistryUpdate
	ScopeRegistryAdmin

	ScopeDomainReadonly
	ScopeDomainUpdate
	ScopeDomainAdmin
)

var AuthScopes = api.AuthScopes{
	ScopeEnvironmentReadonly: "env-readonly",
	ScopeEnvironmentUpdate:   "env-update",
	ScopeEnvironmentAdmin:    "env-admin",
	ScopeRegistryReadonly:    "registry-readonly",
	ScopeRegistryUpdate:      "registry-update",
	ScopeRegistryAdmin:       "registry-admin",
	ScopeDomainReadonly:      "domain-readonly",
	ScopeDomainUpdate:        "domain-update",
	ScopeDomainAdmin:         "domain-admin",
}

const (
	Info api.ServiceMethod = iota

	// Websocket test
	RunScript
	EventsFeed
	PubSubTopic

	// Domains
	ListDomains
	GetDomain

	// Environments
	GetEnvironmentVars
	UpdateEnvironmentVars

	GetRegistry
	UpdateRegistry
	DeleteRegistry
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

	GetEnvironmentVars: api.MethodSpec{
		AuthScope: AuthScopes[ScopeEnvironmentReadonly],
		Doc: `
Get environment variables
`,
		UrlRoute:   "/v1/env/{domain_class}/{domain_instance}/{service}/{version}",
		HttpMethod: "GET",
		ResponseBody: func(req *http.Request) interface{} {
			return EnvList{}
		},
	},

	UpdateEnvironmentVars: api.MethodSpec{
		AuthScope: AuthScopes[ScopeEnvironmentUpdate],
		Doc: `
Update environment variables
`,
		UrlRoute:     "/v1/env/{domain_class}/{domain_instance}/{service}/{version}",
		HttpMethod:   "POST",
		ContentTypes: []string{"application/json"},
		RequestBody: func(req *http.Request) interface{} {
			return new(EnvChange)
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

	ListDomains: api.MethodSpec{
		AuthScope: AuthScopes[ScopeDomainReadonly],
		Doc: `
List domains that the user has access to.
`,
		UrlRoute:   "/v1/domains",
		HttpMethod: "GET",
		ResponseBody: func(req *http.Request) interface{} {
			return DomainList{}
		},
	},

	GetDomain: api.MethodSpec{
		AuthScope: AuthScopes[ScopeDomainReadonly],
		Doc: `
Get information on the domain
`,
		UrlRoute:   "/v1/{domain}",
		HttpMethod: "GET",
		ResponseBody: func(req *http.Request) interface{} {
			return DomainDetail{}
		},
	},

	GetRegistry: api.MethodSpec{
		AuthScope: AuthScopes[ScopeRegistryReadonly],
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
		AuthScope: AuthScopes[ScopeRegistryUpdate],
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

	DeleteRegistry: api.MethodSpec{
		AuthScope: AuthScopes[ScopeRegistryUpdate],
		Doc: `
Update registry key
`,
		UrlRoute:     "/v1/reg/{key}",
		HttpMethod:   "DELETE",
		ContentTypes: []string{"application/json"},
		RequestBody: func(req *http.Request) interface{} {
			return new(string)
		},
	},

	/////////////////////////////////////////////////////////////////////////////////
	// PROTOTYPING

	PubSubTopic: api.MethodSpec{
		Doc: `
Websocket to a pubsub topic
`,
		UrlRoute:   "/v1/ws/feed/",
		HttpMethod: "GET",
		UrlQueries: api.UrlQueries{
			"topic": "mqtt://iot.eclipse.org:1883/test",
		}, ResponseBody: func(req *http.Request) interface{} {
			return make([]string, 0)
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

type RegistryEntry struct {
	Path  string `json:"path"`
	Value string `json:"value"`
}

type DomainList []Domain
type Domain struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Url  string `json:"url"`
}

type DomainDetail struct {
	Id   string `json:"id"`
	Name string `json:"name"`
	Url  string `json:"url"`
}
