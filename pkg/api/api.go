package api

import (
	"github.com/qorio/omni/api"
	"github.com/qorio/omni/version"
	"net/http"
)

const (
	VersionHeader = "X-Dash-Version"
	PathRegex     = "[:0-9a-zA-Z\\.\\-]+(/[:0-9a-zA-Z\\.\\-]+)*"
)

const (
	ScopeEnvReadonly api.AuthScope = iota
	ScopeEnvUpdate
	ScopeEnvAdmin

	ScopeConfUpdate
	ScopeConfReadonly

	ScopePkgUpdate
	ScopePkgReadonly

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

	ScopeLiveVersionUpdate
	ScopeLiveVersionReadonly

	ScopeDockerProxyUpdate
	ScopeDockerProxyReadonly
)

var AuthScopes = api.AuthScopes{
	ScopeEnvAdmin:     "env-admin",
	ScopeEnvReadonly:  "env-readonly",
	ScopeEnvUpdate:    "env-update",
	ScopeConfUpdate:   "conf-update",
	ScopeConfReadonly: "conf-readonly",
	ScopePkgUpdate:    "pkg-update",
	ScopePkgReadonly:  "pkg-readonly",

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

	ScopeLiveVersionUpdate:   "live-version-update",
	ScopeLiveVersionReadonly: "live-version-readonly",

	ScopeDockerProxyUpdate:   "docker-proxy-update",
	ScopeDockerProxyReadonly: "docker-proxy-readonly",
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
	GetEnv
	CreateEnv
	UpdateEnv
	DeleteEnv

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

	SetEnvLiveVersion
	ListEnvVersions
	GetEnvLiveVersion

	SetConfLiveVersion
	ListConfVersions
	GetConfLiveVersion
	ListConfLiveVersions

	ListDomainPkgs
	CreatePkg
	UpdatePkg
	GetPkg
	DeletePkg
	SetPkgLiveVersion
	GetPkgLiveVersion
	ListPkgVersions

	ListDockerProxies
	DockerProxyReadonly
	DockerProxyUpdate
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

	CreateEnv: api.MethodSpec{
		AuthScope: AuthScopes[ScopeEnvUpdate],
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

	UpdateEnv: api.MethodSpec{
		AuthScope: AuthScopes[ScopeEnvUpdate],
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

	GetEnv: api.MethodSpec{
		AuthScope: AuthScopes[ScopeEnvReadonly],
		Doc: `
Get environment variables
`,
		UrlRoute:     "/v1/env/{domain_class}/{domain_instance}/{service}/{version}",
		HttpMethod:   "GET",
		ContentTypes: []string{"application/json"},
		ResponseBody: func(req *http.Request) interface{} {
			return new(EnvList)
		},
	},

	DeleteEnv: api.MethodSpec{
		AuthScope: AuthScopes[ScopeEnvAdmin],
		Doc: `
Delete
`,
		UrlRoute:   "/v1/env/{domain_class}/{domain_instance}/{service}/{version}",
		HttpMethod: "DELETE",
	},

	SetEnvLiveVersion: api.MethodSpec{
		AuthScope: AuthScopes[ScopeEnvUpdate],
		Doc: `
Set this version to live
`,
		UrlRoute:   "/v1/env/{domain_class}/{domain_instance}/{service}/{version}/live",
		HttpMethod: "POST",
	},

	GetEnvLiveVersion: api.MethodSpec{
		AuthScope: AuthScopes[ScopeEnvReadonly],
		Doc: `
Get environment variables for this instance, live version
`,
		UrlRoute:   "/v1/env/{domain_class}/{domain_instance}/{service}",
		HttpMethod: "GET",
		ResponseBody: func(req *http.Request) interface{} {
			return new(EnvList)
		},
	},

	ListEnvVersions: api.MethodSpec{
		AuthScope: AuthScopes[ScopeEnvReadonly],
		Doc: `
List known versions, including one that's live.
`,
		UrlRoute:     "/v1/env/{domain_class}/{domain_instance}/{service}/",
		HttpMethod:   "GET",
		ContentTypes: []string{"application/json"},
		ResponseBody: func(req *http.Request) interface{} {
			return new(EnvVersions)
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
		AuthScope: AuthScopes[ScopeConfReadonly],
		Doc: `
List config versions in a domain class
`,
		UrlRoute:     "/v1/conf/{domain_class}/",
		HttpMethod:   "GET",
		ContentTypes: []string{"application/json"},
		RequestBody: func(req *http.Request) interface{} {
			return []Conf{}
		},
	},

	ListConfFiles: api.MethodSpec{
		AuthScope: AuthScopes[ScopeConfReadonly],
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

	ListConfLiveVersions: api.MethodSpec{
		AuthScope: AuthScopes[ScopeConfReadonly],
		Doc: `
List config files in a domain instance of a service
`,
		UrlRoute:     "/v1/conf/{domain_class}/{domain_instance}/{service}/",
		HttpMethod:   "GET",
		ContentTypes: []string{"application/json"},
		ResponseBody: func(req *http.Request) interface{} {
			return make(ConfLiveVersions)
		},
	},

	// The ConfFile system works by overrides.  A base is used for all domain_instances, and versions,
	// unless there's a real version to override it.
	CreateConfFile: api.MethodSpec{
		AuthScope: AuthScopes[ScopeConfUpdate],
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
		AuthScope: AuthScopes[ScopeConfReadonly],
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
		AuthScope: AuthScopes[ScopeConfUpdate],
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
		AuthScope: AuthScopes[ScopeConfUpdate],
		Doc: `
Delete a config file base
`,
		UrlRoute:   "/v1/conf/{domain_class}/{service}/{name}",
		HttpMethod: "DELETE",
	},

	//// Versions

	CreateConfFileVersion: api.MethodSpec{
		AuthScope: AuthScopes[ScopeConfUpdate],
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
		AuthScope: AuthScopes[ScopeConfReadonly],
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
		AuthScope: AuthScopes[ScopeConfUpdate],
		Doc: `
Set live of a particular version of conf for a domain instance
`,
		UrlRoute:     "/v1/conf/{domain_class}/{domain_instance}/{service}/{version}/{name}",
		HttpMethod:   "PUT",
		ContentTypes: []string{"text/plain"}, //"application/octet-stream"},
		RequestBody: func(req *http.Request) interface{} {
			return new(ConfFile)
		},
	},

	DeleteConfFileVersion: api.MethodSpec{
		AuthScope: AuthScopes[ScopeConfUpdate],
		Doc: `
Delete a config file version
`,
		UrlRoute:   "/v1/conf/{domain_class}/{domain_instance}/{service}/{version}/{name}",
		HttpMethod: "DELETE",
	},

	SetConfLiveVersion: api.MethodSpec{
		AuthScope: AuthScopes[ScopeConfUpdate],
		Doc: `
Set live version of a conf for a domain instance
`,
		UrlRoute:     "/v1/conf/{domain_class}/{domain_instance}/{service}/{version}/{name}/live",
		HttpMethod:   "POST",
		ContentTypes: []string{"text/plain"}, //"application/octet-stream"},
		ResponseBody: func(req *http.Request) interface{} {
			return new(ConfFile)
		},
	},

	ListConfVersions: api.MethodSpec{
		AuthScope: AuthScopes[ScopeConfReadonly],
		Doc: `
List the versions of a conf in a given domain instance
`,
		UrlRoute:     "/v1/conf/{domain_class}/{domain_instance}/{service}/{name}/",
		HttpMethod:   "GET",
		ContentTypes: []string{"application/json"},
		ResponseBody: func(req *http.Request) interface{} {
			return make(ConfVersions)
		},
	},

	GetConfLiveVersion: api.MethodSpec{
		AuthScope: AuthScopes[ScopeConfReadonly],
		Doc: `
Get the live version of a conf in a given domain instance
`,
		UrlRoute:     "/v1/conf/{domain_class}/{domain_instance}/{service}/{name}",
		HttpMethod:   "GET",
		ContentTypes: []string{"text/plain"}, //"application/octet-stream"},
		ResponseBody: func(req *http.Request) interface{} {
			return new(ConfFile)
		},
	},

	/////////////////////////////////////////////////////////////////////////////////
	// PKG

	ListDomainPkgs: api.MethodSpec{
		AuthScope: AuthScopes[ScopeDomainReadonly],
		Doc: `
List all packages in a domain
`,
		UrlRoute:   "/v1/pkg/{domain_class}/",
		HttpMethod: "GET",
		ResponseBody: func(req *http.Request) interface{} {
			return map[string]Pkg{}
		},
	},

	CreatePkg: api.MethodSpec{
		AuthScope: AuthScopes[ScopePkgUpdate],
		Doc: `
Create a software package
`,
		UrlRoute:     "/v1/pkg/{domain_class}/{domain_instance}/{service}/{version}",
		HttpMethod:   "POST",
		ContentTypes: []string{"application/json"},
		RequestBody: func(req *http.Request) interface{} {
			return new(PkgModel)
		},
	},

	UpdatePkg: api.MethodSpec{
		AuthScope: AuthScopes[ScopePkgUpdate],
		Doc: `
Update a package
`,
		UrlRoute:     "/v1/pkg/{domain_class}/{domain_instance}/{service}/{version}",
		HttpMethod:   "PUT",
		ContentTypes: []string{"application/json"},
		RequestBody: func(req *http.Request) interface{} {
			return new(PkgModel)
		},
	},

	GetPkg: api.MethodSpec{
		AuthScope: AuthScopes[ScopePkgReadonly],
		Doc: `
Get a package
`,
		UrlRoute:     "/v1/pkg/{domain_class}/{domain_instance}/{service}/{version}",
		HttpMethod:   "GET",
		ContentTypes: []string{"application/json"},
		ResponseBody: func(req *http.Request) interface{} {
			return new(PkgModel)
		},
	},

	DeletePkg: api.MethodSpec{
		AuthScope: AuthScopes[ScopePkgUpdate],
		Doc: `
Delete a package
`,
		UrlRoute:   "/v1/pkg/{domain_class}/{domain_instance}/{service}/{version}",
		HttpMethod: "DELETE",
	},

	SetPkgLiveVersion: api.MethodSpec{
		AuthScope: AuthScopes[ScopePkgUpdate],
		Doc: `
Set this version to live
`,
		UrlRoute:   "/v1/pkg/{domain_class}/{domain_instance}/{service}/{version}/live",
		HttpMethod: "POST",
	},

	GetPkgLiveVersion: api.MethodSpec{
		AuthScope: AuthScopes[ScopePkgReadonly],
		Doc: `
Get live package for this instance, live version
`,
		UrlRoute:   "/v1/pkg/{domain_class}/{domain_instance}/{service}",
		HttpMethod: "GET",
		ResponseBody: func(req *http.Request) interface{} {
			return new(PkgModel)
		},
	},

	ListPkgVersions: api.MethodSpec{
		AuthScope: AuthScopes[ScopePkgReadonly],
		Doc: `
List known versions, including one that's live.
`,
		UrlRoute:     "/v1/pkg/{domain_class}/{domain_instance}/{service}/",
		HttpMethod:   "GET",
		ContentTypes: []string{"application/json"},
		ResponseBody: func(req *http.Request) interface{} {
			return new(PkgVersions)
		},
	},

	/////////////////////////////////////////////////////////////////////////////////
	// DOCKER API
	ListDockerProxies: api.MethodSpec{
		AuthScope: AuthScopes[ScopeDockerProxyReadonly],
		Doc: `
List docker proxies
`,
		UrlRoute:     "/v1/dockerapi/{domain_class}/{domain_instance}/",
		HttpMethod:   "GET",
		ContentTypes: []string{"application/json"},
		ResponseBody: func(req *http.Request) interface{} {
			return new(DockerProxies)
		},
	},

	DockerProxyReadonly: api.MethodSpec{
		AuthScope: AuthScopes[ScopeDockerProxyReadonly],
		Doc: `
Docker proxy gets
`,
		UrlRoute:    "/v1/dockerapi/{domain_class}/{domain_instance}/{target}/{docker:.*}",
		HttpMethods: []api.HttpMethod{api.HEAD, api.GET},
	},
	DockerProxyUpdate: api.MethodSpec{
		AuthScope: AuthScopes[ScopeDockerProxyUpdate],
		Doc: `
Docker proxy updates
`,
		UrlRoute:    "/v1/dockerapi/{domain_class}/{domain_instance}/{target}/{docker:.*}",
		HttpMethods: []api.HttpMethod{api.POST, api.PUT, api.DELETE, api.PATCH},
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
