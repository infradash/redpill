package api

import (
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

	ScopeOrchestrateModelUpdate
	ScopeOrchestrateModelReadonly

	ScopeConfFileUpdate
	ScopeConfFileReadonly
)

var AuthScopes = api.AuthScopes{
	ScopeEnvironmentReadonly:      "env-readonly",
	ScopeEnvironmentUpdate:        "env-update",
	ScopeEnvironmentAdmin:         "env-admin",
	ScopeRegistryReadonly:         "registry-readonly",
	ScopeRegistryUpdate:           "registry-update",
	ScopeRegistryAdmin:            "registry-admin",
	ScopeDomainReadonly:           "domain-readonly",
	ScopeDomainUpdate:             "domain-update",
	ScopeDomainAdmin:              "domain-admin",
	ScopeOrchestrateStart:         "orchestrate-start",
	ScopeOrchestrateReadonly:      "orchestrate-readonly",
	ScopeOrchestrateModelUpdate:   "orchestrate-model-update",
	ScopeOrchestrateModelReadonly: "orchestrate-model-readonly",
	ScopeConfFileUpdate:           "config-file-update",
	ScopeConfFileReadonly:         "config-file-readonly",
}

const (
	ServerInfo api.ServiceMethod = iota

	// Websocket test
	RunScript
	EventFeed
	PubSubTopic

	// Domains
	ListDomains
	GetDomain
	CreateDomain
	UpdateDomain

	// Environments
	ListDomainEnvs
	GetEnvironmentVars
	CreateEnvironmentVars
	UpdateEnvironmentVars

	GetRegistryEntry
	UpdateRegistryEntry
	DeleteRegistryEntry

	ListOrchestrations
	StartOrchestration
	WatchOrchestration
	ListOrchestrationInstances
	GetOrchestrationInstance

	GetOrchestrationModel
	CreateOrchestrationModel
	UpdateOrchestrationModel
	DeleteOrchestrationModel

	ListDomainConfs
	CreateConfFile
	UpdateConfFile
	DeleteConfFile
	GetConfFile
	ListConfFiles
	CreateConfFileVersion
	UpdateConfFileVersion
	DeleteConfFileVersion
	GetConfFileVersion
)

var Methods = api.ServiceMethods{

	ServerInfo: api.MethodSpec{
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

	///////////////////////////////////////// DOMAIN /////////////////////////////////////////////
	ListDomains: api.MethodSpec{
		AuthScope: AuthScopes[ScopeDomainReadonly],
		Doc: `
List domains that the user has access to.
`,
		UrlRoute:   "/v1/domain/",
		HttpMethod: "GET",
		ResponseBody: func(req *http.Request) interface{} {
			return []DomainInfo{}
		},
	},

	GetDomain: api.MethodSpec{
		AuthScope: AuthScopes[ScopeDomainReadonly],
		Doc: `
Get information on the domain
`,
		UrlRoute:   "/v1/domain/{domain_class}",
		HttpMethod: "GET",
		ResponseBody: func(req *http.Request) interface{} {
			return new(DomainModel)
		},
	},

	CreateDomain: api.MethodSpec{
		AuthScope: AuthScopes[ScopeDomainAdmin],
		Doc: `
Create a domain
`,
		UrlRoute:   "/v1/domain",
		HttpMethod: "POST",
		RequestBody: func(req *http.Request) interface{} {
			return new(DomainModel)
		},
	},

	UpdateDomain: api.MethodSpec{
		AuthScope: AuthScopes[ScopeDomainAdmin],
		Doc: `
Update a domain
`,
		UrlRoute:   "/v1/domain/{domain_class}",
		HttpMethod: "PUT",
		RequestBody: func(req *http.Request) interface{} {
			return new(DomainModel)
		},
	},

	///////////////////////////////////////// ENV /////////////////////////////////////////////
	ListDomainEnvs: api.MethodSpec{
		AuthScope: AuthScopes[ScopeDomainReadonly],
		Doc: `
List all environment variables in a domain
`,
		UrlRoute:   "/v1/env/{domain_class}/",
		HttpMethod: "GET",
		ResponseBody: func(req *http.Request) interface{} {
			return map[string]Env{}
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

	///////////////////////////////////////// EVENTS /////////////////////////////////////////////

	EventFeed: api.MethodSpec{
		Doc: `
Main events feed
`,
		UrlRoute:   "/v1/events",
		HttpMethod: "GET",
		ResponseBody: func(req *http.Request) interface{} {
			return make(<-chan Event)
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

	ListOrchestrationInstances: api.MethodSpec{
		AuthScope: AuthScopes[ScopeOrchestrateReadonly],
		Doc: `
List all running orchestrations
`,
		UrlRoute:     "/v1/orchestrate/{domain_class}/{domain_instance}/{orchestration}/",
		HttpMethod:   "GET",
		ContentTypes: []string{"application/json"},
		ResponseBody: func(req *http.Request) interface{} {
			return []OrchestrationInfo{}
		},
	},

	StartOrchestration: api.MethodSpec{
		AuthScope: AuthScopes[ScopeOrchestrateStart],
		Doc: `
Start an orchestration instance
`,
		UrlRoute:     "/v1/orchestrate/{domain_class}/{domain_instance}/{orchestration}",
		HttpMethod:   "POST",
		ContentTypes: []string{"application/json"},
		RequestBody: func(req *http.Request) interface{} {
			return new(StartOrchestrationRequest)
		},
		ResponseBody: func(req *http.Request) interface{} {
			return new(StartOrchestrationResponse)
		},
	},

	GetOrchestrationInstance: api.MethodSpec{
		AuthScope: AuthScopes[ScopeOrchestrateStart],
		Doc: `
Get an orchestration instance
`,
		UrlRoute:     "/v1/orchestrate/{domain_class}/{domain_instance}/{orchestration}/{instance_id}",
		HttpMethod:   "GET",
		ContentTypes: []string{"application/json"},
		ResponseBody: func(req *http.Request) interface{} {
			return new(OrchestrationInstance)
		},
	},

	WatchOrchestration: api.MethodSpec{
		AuthScope: AuthScopes[ScopeOrchestrateReadonly],
		Doc: `
Watch the feed of an orchestration instance
`,
		UrlRoute:     "/v1/ws/orchestrate/{domain_class}/{domain_instance}/{orchestration}/{instance_id}",
		HttpMethod:   "GET",
		ContentTypes: []string{"application/json"},
		ResponseBody: func(req *http.Request) interface{} {
			return []string{}
		},
	},

	/////////////////////////////////////  MODELS ////////////////////////////////////////////
	CreateOrchestrationModel: api.MethodSpec{
		AuthScope: AuthScopes[ScopeOrchestrateModelUpdate],
		Doc: `
Create or update the model for an orchestration
`,
		UrlRoute:     "/v1/model/{domain_class}",
		HttpMethod:   "POST",
		ContentTypes: []string{"application/json"},
		RequestBody: func(req *http.Request) interface{} {
			return new(OrchestrationModel)
		},
	},

	UpdateOrchestrationModel: api.MethodSpec{
		AuthScope: AuthScopes[ScopeOrchestrateModelUpdate],
		Doc: `
Create or update the model for an orchestration
`,
		UrlRoute:     "/v1/model/{domain_class}",
		HttpMethod:   "PUT",
		ContentTypes: []string{"application/json"},
		RequestBody: func(req *http.Request) interface{} {
			return new(OrchestrationModel)
		},
	},

	GetOrchestrationModel: api.MethodSpec{
		AuthScope: AuthScopes[ScopeOrchestrateModelReadonly],
		Doc: `
Get the model
`,
		UrlRoute:     "/v1/model/{domain_class}/{orchestration}",
		HttpMethod:   "GET",
		ContentTypes: []string{"application/json"},
		ResponseBody: func(req *http.Request) interface{} {
			return new(OrchestrationModel)
		},
	},

	DeleteOrchestrationModel: api.MethodSpec{
		AuthScope: AuthScopes[ScopeOrchestrateModelUpdate],
		Doc: `
Get the model
`,
		UrlRoute:     "/v1/model/{domain_class}/{orchestration}",
		HttpMethod:   "DELETE",
		ContentTypes: []string{"application/json"},
	},

	/////////////////////////////////////  CONFIGS ////////////////////////////////////////////

	ListDomainConfs: api.MethodSpec{
		AuthScope: AuthScopes[ScopeConfFileReadonly],
		Doc: `
List config versions in a domain class
`,
		UrlRoute:     "/v1/conf/{domain_class}/",
		HttpMethod:   "GET",
		ContentTypes: []string{"application/json"},
		RequestBody: func(req *http.Request) interface{} {
			return []Env{}
		},
	},

	ListConfFiles: api.MethodSpec{
		AuthScope: AuthScopes[ScopeConfFileReadonly],
		Doc: `
List config files
`,
		UrlRoute:     "/v1/conf/{domain_class}/{service}/",
		HttpMethod:   "GET",
		ContentTypes: []string{"application/json"},
		ResponseBody: func(req *http.Request) interface{} {
			return []ConfInfo{}
		},
	},

	// The ConfFile system works by overrides.  A base is used for all domain_instances, and versions,
	// unless there's a real version to override it.
	CreateConfFile: api.MethodSpec{
		AuthScope: AuthScopes[ScopeConfFileUpdate],
		Doc: `
Create a config file
`,
		UrlRoute:     "/v1/conf/{domain_class}/{service}/{name}",
		HttpMethod:   "POST",
		ContentTypes: []string{"text/plain"}, //"application/octet-stream"},
		RequestBody: func(req *http.Request) interface{} {
			return new(ConfFile)
		},
	},

	GetConfFile: api.MethodSpec{
		AuthScope: AuthScopes[ScopeConfFileReadonly],
		Doc: `
Get a config file
`,
		UrlRoute:     "/v1/conf/{domain_class}/{service}/{name}",
		HttpMethod:   "GET",
		ContentTypes: []string{"text/plain"}, //"application/octet-stream"},
		ResponseBody: func(req *http.Request) interface{} {
			return new(ConfFile)
		},
	},

	UpdateConfFile: api.MethodSpec{
		AuthScope: AuthScopes[ScopeConfFileUpdate],
		Doc: `
Update a config file
`,
		UrlRoute:     "/v1/conf/{domain_class}/{service}/{name}",
		HttpMethod:   "PUT",
		ContentTypes: []string{"text/plain"}, //"application/octet-stream"},
		RequestBody: func(req *http.Request) interface{} {
			return new(ConfFile)
		},
	},

	DeleteConfFile: api.MethodSpec{
		AuthScope: AuthScopes[ScopeConfFileUpdate],
		Doc: `
Delete a config file base
`,
		UrlRoute:   "/v1/conf/{domain_class}/{service}/{name}",
		HttpMethod: "DELETE",
	},

	//// Versions

	CreateConfFileVersion: api.MethodSpec{
		AuthScope: AuthScopes[ScopeConfFileUpdate],
		Doc: `
Create a config file version
`,
		UrlRoute:     "/v1/conf/{domain_class}/{domain_instance}/{service}/{version}/{name}",
		HttpMethod:   "POST",
		ContentTypes: []string{"text/plain"}, //"application/octet-stream"},
		ResponseBody: func(req *http.Request) interface{} {
			return new(ConfFile)
		},
	},

	GetConfFileVersion: api.MethodSpec{
		AuthScope: AuthScopes[ScopeConfFileReadonly],
		Doc: `
Get a config file version
`,
		UrlRoute:     "/v1/conf/{domain_class}/{domain_instance}/{service}/{version}/{name}",
		HttpMethod:   "GET",
		ContentTypes: []string{"text/plain"}, //"application/octet-stream"},
		ResponseBody: func(req *http.Request) interface{} {
			return new(ConfFile)
		},
	},

	UpdateConfFileVersion: api.MethodSpec{
		AuthScope: AuthScopes[ScopeConfFileUpdate],
		Doc: `
Update a config file version
`,
		UrlRoute:     "/v1/conf/{domain_class}/{domain_instance}/{service}/{version}/{name}",
		HttpMethod:   "PUT",
		ContentTypes: []string{"text/plain"}, //"application/octet-stream"},
		RequestBody: func(req *http.Request) interface{} {
			return new(ConfFile)
		},
	},

	DeleteConfFileVersion: api.MethodSpec{
		AuthScope: AuthScopes[ScopeConfFileUpdate],
		Doc: `
Delete a config file version
`,
		UrlRoute:   "/v1/conf/{domain_class}/{domain_instance}/{service}/{version}/{name}",
		HttpMethod: "DELETE",
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

type RegistryEntry struct {
	Path  string `json:"path"`
	Value string `json:"value"`
}

type StartOrchestrationRequest struct {
	Note    string               `json:"note"`
	Context OrchestrationContext `json:"context"`
}

type StartOrchestrationResponse struct {
	Id        string                 `json:"id"`
	StartTime int64                  `json:"start_timestamp"`
	LogWsUrl  string                 `json:"log_ws_url"`
	Context   map[string]interface{} `json:"context,omitempty"`
	Note      string                 `json:"note,omitempty"`
}

type ConfFile []byte
