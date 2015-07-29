package api

import (
	"encoding/json"
	"github.com/qorio/omni/api"
	"github.com/qorio/omni/version"
	"net/http"
)

const (
	PathRegex = "[:0-9a-zA-Z\\.\\-]+(/[:0-9a-zA-Z\\.\\-]+)*"
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

	ScopeOrchestrateStart
	ScopeOrchestrateReadonly
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
	ScopeOrchestrateStart:    "orchestrate-start",
	ScopeOrchestrateReadonly: "orchestrate-readonly",
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
	CreateEnvironmentVars
	UpdateEnvironmentVars

	GetRegistryEntry
	UpdateRegistryEntry
	DeleteRegistryEntry

	ListOrchestrations
	ListRunningOrchestrations
	StartOrchestration
	WatchOrchestration
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
			return new(EnvList)
		},
	},

	CreateEnvironmentVars: api.MethodSpec{
		AuthScope: AuthScopes[ScopeEnvironmentUpdate],
		Doc: `
Create environment variables for a new domain/ environment
`,
		UrlRoute:     "/v1/env/{domain_class}/{domain_instance}/{service}/{version}",
		HttpMethod:   "POST",
		ContentTypes: []string{"application/json"},
		RequestBody: func(req *http.Request) interface{} {
			return new(EnvList)
		},
	},

	UpdateEnvironmentVars: api.MethodSpec{
		AuthScope: AuthScopes[ScopeEnvironmentUpdate],
		Doc: `
Update environment variables
`,
		UrlRoute:     "/v1/env/{domain_class}/{domain_instance}/{service}/{version}",
		HttpMethod:   "PATCH",
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
			return new(EventList)
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
			return new(DomainList)
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
			return new(DomainDetail)
		},
	},

	GetRegistryEntry: api.MethodSpec{
		AuthScope: AuthScopes[ScopeRegistryReadonly],
		Doc: `
Get registry key
`,
		UrlRoute:   "/v1/reg/{path:" + PathRegex + "}",
		HttpMethod: "GET",
		ResponseBody: func(req *http.Request) interface{} {
			return new(RegistryEntry)
		},
	},

	UpdateRegistryEntry: api.MethodSpec{
		AuthScope: AuthScopes[ScopeRegistryUpdate],
		Doc: `
Update registry key
`,
		UrlRoute:     "/v1/reg/{path:" + PathRegex + "}",
		HttpMethod:   "POST",
		ContentTypes: []string{"application/json"},
		RequestBody: func(req *http.Request) interface{} {
			return new(RegistryEntry)
		},
	},

	DeleteRegistryEntry: api.MethodSpec{
		AuthScope: AuthScopes[ScopeRegistryUpdate],
		Doc: `
Update registry key
`,
		UrlRoute:     "/v1/reg/{path:" + PathRegex + "}",
		HttpMethod:   "DELETE",
		ContentTypes: []string{"application/json"},
	},

	ListOrchestrations: api.MethodSpec{
		AuthScope: AuthScopes[ScopeOrchestrateReadonly],
		Doc: `
List available orchestrations
`,
		UrlRoute:     "/v1/orchestrate/{domain_class}/{domain_instance}/",
		HttpMethod:   "GET",
		ContentTypes: []string{"application/json"},
		ResponseBody: func(req *http.Request) interface{} {
			return new(OrchestrationList)
		},
	},

	ListRunningOrchestrations: api.MethodSpec{
		AuthScope: AuthScopes[ScopeOrchestrateReadonly],
		Doc: `
List all running orchestrations
`,
		UrlRoute:     "/v1/orchestrate/",
		HttpMethod:   "GET",
		ContentTypes: []string{"application/json"},
		ResponseBody: func(req *http.Request) interface{} {
			return new(OrchestrationList)
		},
	},

	StartOrchestration: api.MethodSpec{
		AuthScope: AuthScopes[ScopeOrchestrateStart],
		Doc: `
Start an orchestration instance
`,
		UrlRoute:     "/v1/orchestrate/{domain}/{orchestration}",
		HttpMethod:   "POST",
		ContentTypes: []string{"application/json"},
		RequestBody: func(req *http.Request) interface{} {
			return new(StartOrchestrationRequest)
		},
		ResponseBody: func(req *http.Request) interface{} {
			return new(StartOrchestrationResponse)
		},
	},

	WatchOrchestration: api.MethodSpec{
		AuthScope: AuthScopes[ScopeOrchestrateReadonly],
		Doc: `
Watch the feed of an orchestration instance
`,
		UrlRoute:     "/v1/ws/feed/{domain}/{orchestration}/{instance_id}",
		HttpMethod:   "GET",
		ContentTypes: []string{"application/json"},
		ResponseBody: func(req *http.Request) interface{} {
			return []string{}
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
		},
		ResponseBody: func(req *http.Request) interface{} {
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

type OrchestrationList []OrchestrationDescription
type OrchestrationDescription struct {
	Name         string      `json:"name,omitempty"`
	Label        string      `json:"label,omitempty"`
	Description  string      `json:"description,omitempty"`
	ActivateUrl  string      `json:"activate_url"`
	DefaultInput interface{} `json:"default_input,omitempty"`
}

type StartOrchestrationRequest json.RawMessage

type StartOrchestrationResponse struct {
	Id        string                 `json:"id"`
	StartTime int64                  `json:"start_timestamp"`
	LogWsUrl  string                 `json:"log_ws_url"`
	Context   map[string]interface{} `json:"context,omitempty"`
}
