package redpill

import (
	"bufio"
	"github.com/golang/glog"
	"github.com/gorilla/websocket"
	. "github.com/infradash/redpill/pkg/api"
	"github.com/qorio/maestro/pkg/pubsub"
	"github.com/qorio/omni/auth"
	"github.com/qorio/omni/rest"
	"github.com/qorio/omni/version"
	"io"
	"net/http"
	"os/exec"
	"path/filepath"
	"sync"
)

type Api struct {
	options     Options
	authService auth.Service
	engine      rest.Engine

	event       EventService
	env         EnvService
	domain      DomainService
	registry    RegistryService
	orchestrate OrchestrateService
	conf        ConfService
	pkg         PkgService
	dockerapi   DockerProxyService

	CreateServiceContext CreateContextFunc
}

var ServiceId = "redpill"

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func NewApi(options Options, auth auth.Service,
	env EnvService,
	domain DomainService,
	event EventService,
	registry RegistryService,
	orchestrate OrchestrateService,
	conf ConfService,
	pkg PkgService,
	dockerapi DockerProxyService) (*Api, error) {
	ep := &Api{
		options:     options,
		authService: auth,
		engine:      rest.NewEngine(&Methods, auth, nil),
		env:         env,
		domain:      domain,
		registry:    registry,
		orchestrate: orchestrate,
		conf:        conf,
		event:       event,
		pkg:         pkg,
		dockerapi:   dockerapi,
	}

	ep.CreateServiceContext = ServiceContext(ep.engine)

	ep.engine.Bind(

		rest.SetHandler(Methods[ServerInfo], ep.GetServerInfo),
		rest.SetHandler(Methods[RunScript], ep.WsRunScript),
		rest.SetHandler(Methods[EventFeed], ep.WsEventFeed),
		rest.SetHandler(Methods[PubSubTopic], ep.WsPubSubTopic),

		// Domains
		rest.SetAuthenticatedHandler(ServiceId, Methods[ListDomains], ep.ListDomains),
		rest.SetAuthenticatedHandler(ServiceId, Methods[CreateDomain], ep.CreateDomain),
		rest.SetAuthenticatedHandler(ServiceId, Methods[UpdateDomain], ep.UpdateDomain),
		rest.SetAuthenticatedHandler(ServiceId, Methods[GetDomain], ep.GetDomain),

		// Environments
		rest.SetAuthenticatedHandler(ServiceId, Methods[ListDomainEnvs], ep.ListDomainEnvs),
		rest.SetAuthenticatedHandler(ServiceId, Methods[CreateEnv], ep.CreateEnv),
		rest.SetAuthenticatedHandler(ServiceId, Methods[UpdateEnv], ep.UpdateEnv),
		rest.SetAuthenticatedHandler(ServiceId, Methods[GetEnv], ep.GetEnv),
		rest.SetAuthenticatedHandler(ServiceId, Methods[DeleteEnv], ep.DeleteEnv),
		rest.SetAuthenticatedHandler(ServiceId, Methods[SetEnvLiveVersion], ep.SetEnvLiveVersion),
		rest.SetAuthenticatedHandler(ServiceId, Methods[ListEnvVersions], ep.ListEnvVersions),
		rest.SetAuthenticatedHandler(ServiceId, Methods[GetEnvLiveVersion], ep.GetEnvLiveVersion),

		// ConfigFiles
		rest.SetAuthenticatedHandler(ServiceId, Methods[ListDomainConfs], ep.ListDomainConfs),
		rest.SetAuthenticatedHandler(ServiceId, Methods[CreateConfFile], ep.CreateConfFile),
		rest.SetAuthenticatedHandler(ServiceId, Methods[UpdateConfFile], ep.UpdateConfFile),
		rest.SetAuthenticatedHandler(ServiceId, Methods[GetConfFile], ep.GetConfFile),
		rest.SetAuthenticatedHandler(ServiceId, Methods[DeleteConfFile], ep.DeleteConfFile),
		rest.SetAuthenticatedHandler(ServiceId, Methods[ListConfFiles], ep.ListConfFiles),
		rest.SetAuthenticatedHandler(ServiceId, Methods[ListConfLiveVersions], ep.ListConfLiveVersions),

		rest.SetAuthenticatedHandler(ServiceId, Methods[GetConfFileVersion], ep.GetConfFileVersion),
		rest.SetAuthenticatedHandler(ServiceId, Methods[CreateConfFileVersion], ep.CreateConfFileVersion),
		rest.SetAuthenticatedHandler(ServiceId, Methods[UpdateConfFileVersion], ep.UpdateConfFileVersion),
		rest.SetAuthenticatedHandler(ServiceId, Methods[DeleteConfFileVersion], ep.DeleteConfFileVersion),

		rest.SetAuthenticatedHandler(ServiceId, Methods[SetConfLiveVersion], ep.SetConfLiveVersion),
		rest.SetAuthenticatedHandler(ServiceId, Methods[ListConfVersions], ep.ListConfVersions),
		rest.SetAuthenticatedHandler(ServiceId, Methods[GetConfLiveVersion], ep.GetConfLiveVersion),

		// Packages
		rest.SetAuthenticatedHandler(ServiceId, Methods[ListDomainPkgs], ep.ListDomainPkgs),
		rest.SetAuthenticatedHandler(ServiceId, Methods[CreatePkg], ep.CreatePkg),
		rest.SetAuthenticatedHandler(ServiceId, Methods[UpdatePkg], ep.UpdatePkg),
		rest.SetAuthenticatedHandler(ServiceId, Methods[GetPkg], ep.GetPkg),
		rest.SetAuthenticatedHandler(ServiceId, Methods[DeletePkg], ep.DeletePkg),
		rest.SetAuthenticatedHandler(ServiceId, Methods[SetPkgLiveVersion], ep.SetPkgLiveVersion),
		rest.SetAuthenticatedHandler(ServiceId, Methods[GetPkgLiveVersion], ep.GetPkgLiveVersion),
		rest.SetAuthenticatedHandler(ServiceId, Methods[ListPkgVersions], ep.ListPkgVersions),

		// Registry
		rest.SetAuthenticatedHandler(ServiceId, Methods[GetRegistryEntry], ep.GetRegistryEntry),
		rest.SetAuthenticatedHandler(ServiceId, Methods[UpdateRegistryEntry], ep.UpdateRegistryEntry),
		rest.SetAuthenticatedHandler(ServiceId, Methods[DeleteRegistryEntry], ep.DeleteRegistryEntry),

		// Orchestration
		rest.SetAuthenticatedHandler(ServiceId, Methods[ListOrchestrations], ep.ListOrchestrations),
		rest.SetAuthenticatedHandler(ServiceId, Methods[StartOrchestration], ep.StartOrchestration),
		rest.SetAuthenticatedHandler(ServiceId, Methods[WatchOrchestration], ep.WatchOrchestration),
		rest.SetAuthenticatedHandler(ServiceId, Methods[GetOrchestrationInstance], ep.GetOrchestrationInstance),
		rest.SetAuthenticatedHandler(ServiceId, Methods[ListOrchestrationInstances], ep.ListOrchestrationInstances),

		// Models
		rest.SetAuthenticatedHandler(ServiceId, Methods[GetOrchestrationModel], ep.GetOrchestrationModel),
		rest.SetAuthenticatedHandler(ServiceId, Methods[CreateOrchestrationModel], ep.CreateOrchestrationModel),
		rest.SetAuthenticatedHandler(ServiceId, Methods[UpdateOrchestrationModel], ep.CreateOrchestrationModel),
		rest.SetAuthenticatedHandler(ServiceId, Methods[DeleteOrchestrationModel], ep.DeleteOrchestrationModel),

		// Docker proxy
		rest.SetAuthenticatedHandler(ServiceId, Methods[ListDockerProxies], ep.ListDockerProxies),
		rest.SetAuthenticatedHandler(ServiceId, Methods[DockerProxyReadonly], ep.DockerProxyReadonly),
		rest.SetAuthenticatedHandler(ServiceId, Methods[DockerProxyUpdate], ep.DockerProxyUpdate),
	)

	return ep, nil
}

func (this *Api) GetEngine() rest.Engine {
	return this.engine
}

func (this *Api) ServeHTTP(resp http.ResponseWriter, request *http.Request) {
	this.engine.ServeHTTP(resp, request)
}

func (this *Api) GetServerInfo(resp http.ResponseWriter, req *http.Request) {
	build := version.BuildInfo()
	glog.Infoln("Build info:", build)
	err := this.engine.MarshalJSON(req, &build, resp)
	if err != nil {
		this.engine.HandleError(resp, req, "malformed-response", http.StatusInternalServerError)
		return
	}
}

func (this *Api) WsRunScript(resp http.ResponseWriter, req *http.Request) {
	script := this.engine.GetUrlParameter(req, "script")
	this.ws_run_script(resp, req, script)
}

func (this *Api) ws_run_script(resp http.ResponseWriter, req *http.Request, script string) {
	conn, err := upgrader.Upgrade(resp, req, nil)
	if err != nil {
		glog.Infoln("ERROR", err)
		return
	}

	script_file := filepath.Join(this.options.WorkingDir, "scripts", script)
	glog.Infoln("Running script:", script, "file=", script_file)

	command, err := exec.LookPath(script_file)
	if err != nil {
		if conn.WriteMessage(websocket.TextMessage, []byte(err.Error())); err != nil {
			glog.Warningln("WriteErr=", err)
			return
		}
	}

	defer conn.Close()
	readOnly(conn) // Ignore incoming messages

	cmd := exec.Command(command)
	glog.Infoln("Running command", *cmd)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		report_error(conn, err, "no stdout")
		return
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		report_error(conn, err, "no stderr")
		return
	}

	w, err := conn.NextWriter(websocket.TextMessage)
	if err != nil {
		report_error(conn, err, "no ws writer")
		return
	}
	defer w.Close()

	go func() {
		scanner := bufio.NewScanner(io.MultiReader(stdout, stderr))
		for scanner.Scan() {
			err := conn.WriteMessage(websocket.TextMessage, scanner.Bytes())
			if err != nil {
				report_error(conn, err, "ws write error")
				return
			}
		}
	}()

	glog.Infoln(">>>>> Running", *cmd)
	err = cmd.Run()
	if err != nil {
		report_error(conn, err, "run() error")
		return
	}
	glog.Infoln("Completed", *cmd)

}

func report_error(conn *websocket.Conn, err error, msg string) {
	glog.Warningln(msg, "Error=", err)
	conn.WriteMessage(websocket.TextMessage, []byte(err.Error()))
	return
}

func readOnly(c *websocket.Conn) {
	go func() {
		for {
			if _, _, err := c.NextReader(); err != nil {
				c.Close()
				break
			}
		}
	}()
}

var (
	feeds = 0
	mutex sync.Mutex
)

func (this *Api) WsEventFeed(resp http.ResponseWriter, req *http.Request) {
	conn, err := upgrader.Upgrade(resp, req, nil)
	if err != nil {
		glog.Infoln("ERROR", err)
		return
	}

	defer conn.Close()
	readOnly(conn) // Ignore incoming messages

	mutex.Lock()
	feeds += 1
	mutex.Unlock()

	glog.Infoln("Feed #", feeds)

	events := this.event.EventFeed()
	for {
		event := <-events
		if event == nil {
			break
		}
		message, err := event.Marshal()
		if err != nil {
			glog.Warningln("ERROR Mashal:", err)
			continue
		}
		err = conn.WriteMessage(websocket.TextMessage, message)
		if err != nil {
			report_error(conn, err, "ws write error")
			return
		}
	}
	glog.Infoln("Completed")
}

func (this *Api) WsPubSubTopic(resp http.ResponseWriter, req *http.Request) {
	queries, err := this.engine.GetUrlQueries(req, Methods[PubSubTopic].UrlQueries)
	if err != nil {
		this.engine.HandleError(resp, req, "error-bad-request", http.StatusBadRequest)
		return
	}
	topic := pubsub.Topic(queries["topic"].(string))
	glog.Infoln("Connecting ws to topic:", topic)

	if !topic.Valid() {
		glog.Warningln("Topic", topic, "is not valid")
		this.engine.HandleError(resp, req, "bad-topic", http.StatusBadRequest)
		return
	}

	glog.Infoln("Topic using broker", topic.Broker())
	sub, err := topic.Broker().PubSub("test")
	if err != nil {
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, err.Error(), http.StatusInternalServerError)
		return
	}

	conn, err := upgrader.Upgrade(resp, req, nil)
	if err != nil {
		glog.Infoln("ERROR", err)
		return
	}
	readOnly(conn) // Ignore incoming messages
	go func() {
		defer conn.Close()
		in := pubsub.GetReader(topic, sub)
		buff := make([]byte, 4096)
		for {
			n, err := in.Read(buff)
			if err != nil {
				break
			}
			err = conn.WriteMessage(websocket.TextMessage, buff[0:n])
			if err != nil {
				report_error(conn, err, "ws write error")
				return
			}
		}
		glog.Infoln("Completed")
	}()
}
