package redpill

import (
	"bufio"
	"fmt"
	"github.com/golang/glog"
	"github.com/gorilla/websocket"
	. "github.com/infradash/redpill/pkg/api"
	_ "github.com/qorio/maestro/pkg/mqtt"
	"github.com/qorio/maestro/pkg/pubsub"
	"github.com/qorio/omni/auth"
	"github.com/qorio/omni/rest"
	"github.com/qorio/omni/version"
	"io"
	"net/http"
	"os/exec"
	"path/filepath"
	"strconv"
	"sync"
)

type Api struct {
	options     Options
	authService auth.Service
	engine      rest.Engine

	env         EnvService
	domain      DomainService
	registry    RegistryService
	orchestrate OrchestrateService

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
	registry RegistryService,
	orchestrate OrchestrateService) (*Api, error) {
	ep := &Api{
		options:     options,
		authService: auth,
		engine:      rest.NewEngine(&Methods, auth, nil),
		env:         env,
		domain:      domain,
		registry:    registry,
		orchestrate: orchestrate,
	}

	ep.CreateServiceContext = ServiceContext(ep.engine)

	ep.engine.Bind(
		rest.SetHandler(Methods[Info], ep.GetInfo),
		rest.SetHandler(Methods[RunScript], ep.WsRunScript),
		rest.SetHandler(Methods[EventsFeed], ep.WsEventsFeed),
		rest.SetHandler(Methods[PubSubTopic], ep.WsPubSubTopic),

		// Domains
		rest.SetAuthenticatedHandler(ServiceId, Methods[ListDomains], ep.ListDomains),
		rest.SetAuthenticatedHandler(ServiceId, Methods[GetDomain], ep.GetDomain),

		// Environments
		rest.SetAuthenticatedHandler(ServiceId, Methods[ListEnvironmentVars], ep.ListEnvironmentVars),
		rest.SetAuthenticatedHandler(ServiceId, Methods[GetEnvironmentVars], ep.GetEnvironmentVars),
		rest.SetAuthenticatedHandler(ServiceId, Methods[CreateEnvironmentVars], ep.CreateEnvironmentVars),
		rest.SetAuthenticatedHandler(ServiceId, Methods[UpdateEnvironmentVars], ep.UpdateEnvironmentVars),

		// Registry
		rest.SetAuthenticatedHandler(ServiceId, Methods[GetRegistryEntry], ep.GetRegistryEntry),
		rest.SetAuthenticatedHandler(ServiceId, Methods[UpdateRegistryEntry], ep.UpdateRegistryEntry),
		rest.SetAuthenticatedHandler(ServiceId, Methods[DeleteRegistryEntry], ep.DeleteRegistryEntry),

		// Orchestration
		rest.SetAuthenticatedHandler(ServiceId, Methods[ListOrchestrations], ep.ListOrchestrations),
		rest.SetAuthenticatedHandler(ServiceId, Methods[StartOrchestration], ep.StartOrchestration),
		rest.SetAuthenticatedHandler(ServiceId, Methods[WatchOrchestration], ep.WatchOrchestration),
		rest.SetAuthenticatedHandler(ServiceId, Methods[ListOrchestrationInstances], ep.ListOrchestrationInstances),
	)
	return ep, nil
}

func (this *Api) GetEngine() rest.Engine {
	return this.engine
}

func (this *Api) ServeHTTP(resp http.ResponseWriter, request *http.Request) {
	this.engine.ServeHTTP(resp, request)
}

func (this *Api) GetInfo(resp http.ResponseWriter, req *http.Request) {
	build := version.BuildInfo()
	glog.Infoln("Build info:", build)
	err := this.engine.MarshalJSON(req, &build, resp)
	if err != nil {
		this.engine.HandleError(resp, req, "malformed-response", http.StatusInternalServerError)
		return
	}
}

func (this *Api) ListDomains(ac auth.Context, resp http.ResponseWriter, req *http.Request) {
	context := this.CreateServiceContext(ac, req)
	userId := context.UserId()

	glog.Infoln("ListDomains", "UserId=", userId)
	list, err := this.domain.ListDomains(context)
	if err != nil {
		this.engine.HandleError(resp, req, "list-domain-error", http.StatusInternalServerError)
		return
	}
	err = this.engine.MarshalJSON(req, list, resp)
	if err != nil {
		this.engine.HandleError(resp, req, "malformed-result", http.StatusInternalServerError)
		return
	}
}

func (this *Api) GetDomain(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	request := this.CreateServiceContext(context, req)
	userId := request.UserId()

	glog.Infoln("GetDomain", "UserId=", userId)
	domain_class := request.UrlParameter("domain_class")
	detail, err := this.domain.GetDomain(request, domain_class)
	if err != nil {
		this.engine.HandleError(resp, req, "cannot-get-domain", http.StatusInternalServerError)
		return
	}
	err = this.engine.MarshalJSON(req, detail, resp)
	if err != nil {
		this.engine.HandleError(resp, req, "malformed-result", http.StatusInternalServerError)
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

func (this *Api) WsEventsFeed(resp http.ResponseWriter, req *http.Request) {
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

	events := GetEventFeed()
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

func (this *Api) ListEnvironmentVars(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	request := this.CreateServiceContext(context, req)

	domain_class := request.UrlParameter("domain_class")
	result, err := this.env.ListEnvs(request, domain_class)
	if err != nil {
		this.engine.HandleError(resp, req, "query-failed", http.StatusInternalServerError)
		return
	}
	err = this.engine.MarshalJSON(req, result, resp)
	if err != nil {
		this.engine.HandleError(resp, req, "malformed-result", http.StatusInternalServerError)
		return
	}
}

func (this *Api) GetEnvironmentVars(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	request := this.CreateServiceContext(context, req)

	vars, rev, err := this.env.GetEnv(request,
		fmt.Sprintf("%s.%s", request.UrlParameter("domain_instance"), request.UrlParameter("domain_class")),
		request.UrlParameter("service"),
		request.UrlParameter("version"))
	resp.Header().Set("X-Dash-Version", fmt.Sprintf("%d", rev))

	if err != nil {
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, "get-env-fails", http.StatusInternalServerError)
		return
	}
	err = this.engine.MarshalJSON(req, vars, resp)
	if err != nil {
		this.engine.HandleError(resp, req, "malformed-result", http.StatusInternalServerError)
		return
	}
}

func (this *Api) CreateEnvironmentVars(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	request := this.CreateServiceContext(context, req)

	vars := Methods[CreateEnvironmentVars].RequestBody(req).(*EnvList)
	err := this.engine.UnmarshalJSON(req, vars)
	if err != nil {
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, "bad-json", http.StatusBadRequest)
		return
	}

	rev, err := this.env.NewEnv(request,
		fmt.Sprintf("%s.%s", request.UrlParameter("domain_instance"), request.UrlParameter("domain_class")),
		request.UrlParameter("service"),
		request.UrlParameter("version"),
		vars)

	switch {
	case err == ErrConflict:
		this.engine.HandleError(resp, req, "version-conflict", http.StatusConflict)
		return
	case err != nil:
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, "save-env-fails", http.StatusInternalServerError)
		return
	}

	resp.Header().Set("X-Dash-Version", fmt.Sprintf("%d", rev))
}

func (this *Api) UpdateEnvironmentVars(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	request := this.CreateServiceContext(context, req)

	change := Methods[UpdateEnvironmentVars].RequestBody(req).(*EnvChange)
	err := this.engine.UnmarshalJSON(req, change)
	if err != nil {
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, "bad-json", http.StatusBadRequest)
		return
	}

	rev, err := strconv.Atoi(req.Header.Get("X-Dash-Version"))
	if err != nil {
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, "bad-version", http.StatusBadRequest)
		return
	}

	err = this.env.SaveEnv(request,
		fmt.Sprintf("%s.%s", request.UrlParameter("domain_instance"), request.UrlParameter("domain_class")),
		request.UrlParameter("service"),
		request.UrlParameter("version"),
		change,
		Revision(rev))

	switch {
	case err == ErrConflict:
		this.engine.HandleError(resp, req, "version-conflict", http.StatusConflict)
		return
	case err != nil:
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, "save-env-fails", http.StatusInternalServerError)
		return
	}
}

func (this *Api) GetRegistryEntry(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	c := this.CreateServiceContext(context, req)

	result := Methods[GetRegistryEntry].ResponseBody(req).(*RegistryEntry)
	result.Path = "/" + c.UrlParameter("path")

	glog.Infoln("GetRegistry", "path=", result.Path)

	v, rev, err := this.registry.GetEntry(c, result.Path)
	switch {
	case err == ErrNotFound:
		this.engine.HandleError(resp, req, "not-found", http.StatusNotFound)
		return
	case err != nil:
		this.engine.HandleError(resp, req, err.Error(), http.StatusInternalServerError)
		return
	}

	resp.Header().Set("X-Dash-Version", fmt.Sprintf("%d", rev))
	result.Value = string(v)

	err = this.engine.MarshalJSON(req, result, resp)
	if err != nil {
		this.engine.HandleError(resp, req, "malformed-result", http.StatusInternalServerError)
		return
	}
}

func (this *Api) UpdateRegistryEntry(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	change := Methods[UpdateRegistryEntry].RequestBody(req).(*RegistryEntry)
	err := this.engine.UnmarshalJSON(req, change)
	if err != nil {
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, "bad-json", http.StatusBadRequest)
		return
	}

	c := this.CreateServiceContext(context, req)
	path := "/" + c.UrlParameter("path")

	if change.Path != path {
		this.engine.HandleError(resp, req, "conflict", http.StatusBadRequest)
		return
	}

	rev, err := strconv.Atoi(req.Header.Get("X-Dash-Version"))
	if err != nil {
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, "bad-version", http.StatusBadRequest)
		return
	}
	value := []byte(change.Value)
	new_rev, err := this.registry.UpdateEntry(c, path, value, Revision(rev))
	switch {
	case err == ErrNotFound:
		this.engine.HandleError(resp, req, "not-found", http.StatusNotFound)
		return
	case err != nil:
		this.engine.HandleError(resp, req, err.Error(), http.StatusInternalServerError)
		return
	}
	resp.Header().Set("X-Dash-Version", fmt.Sprintf("%d", new_rev))
}

func (this *Api) DeleteRegistryEntry(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	c := this.CreateServiceContext(context, req)
	path := "/" + c.UrlParameter("path")
	rev, err := strconv.Atoi(req.Header.Get("X-Dash-Version"))
	if err != nil {
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, "bad-version", http.StatusBadRequest)
		return
	}
	err = this.registry.DeleteEntry(c, path, Revision(rev))
	switch {
	case err == ErrNotFound:
		this.engine.HandleError(resp, req, "not-found", http.StatusNotFound)
		return
	case err != nil:
		this.engine.HandleError(resp, req, err.Error(), http.StatusInternalServerError)
		return
	}
}

func (this *Api) ListOrchestrations(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	c := this.CreateServiceContext(context, req)
	domain_class := c.UrlParameter("domain_class")
	domain_instance := c.UrlParameter("domain_instance")
	domain := fmt.Sprintf("%s.%s", domain_instance, domain_class)
	glog.Infoln("DomainClass=", domain_class, "DomainInstance=", domain_instance, "Domain=", domain)

	available, err := this.orchestrate.ListOrchestrations(c, domain)

	list := OrchestrationList{}
	for _, o := range available {
		list = append(list, OrchestrationDescription{
			Name:         o.GetName(),
			Label:        o.GetFriendlyName(),
			Description:  o.GetDescription(),
			DefaultInput: o.GetDefaultContext(),
			ActivateUrl:  fmt.Sprintf("/v1/orchestrate/%s/%s", domain, o.GetName()),
		})
	}
	if err != nil {
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, "list-orchestration-error", http.StatusInternalServerError)
		return
	}
	err = this.engine.MarshalJSON(req, list, resp)
	if err != nil {
		this.engine.HandleError(resp, req, "malformed-result", http.StatusInternalServerError)
		return
	}
}

func (this *Api) ListOrchestrationInstances(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	c := this.CreateServiceContext(context, req)
	domain_class := c.UrlParameter("domain_class")
	domain_instance := c.UrlParameter("domain_instance")
	domain := fmt.Sprintf("%s.%s", domain_instance, domain_class)

	orchestration := c.UrlParameter("orchestration")

	glog.Infoln("Domain=", domain, "Orchestration=", orchestration)

	list, err := this.orchestrate.ListInstances(c, domain, orchestration)
	if err != nil {
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, "list-orchestration-error", http.StatusInternalServerError)
		return
	}
	view := Methods[ListOrchestrationInstances].ResponseBody(req).([]OrchestrationInfo)
	for _, l := range list {
		view = append(view, l.Info())
	}
	err = this.engine.MarshalJSON(req, view, resp)
	if err != nil {
		this.engine.HandleError(resp, req, "malformed-result", http.StatusInternalServerError)
		return
	}
}

func (this *Api) StartOrchestration(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	c := this.CreateServiceContext(context, req)
	domain_class := c.UrlParameter("domain_class")
	domain_instance := c.UrlParameter("domain_instance")
	domain := fmt.Sprintf("%s.%s", domain_instance, domain_class)
	orchestration := c.UrlParameter("orchestration")

	glog.Infoln("Orchestration=", orchestration, "Domain=", domain)

	// Get the payload which contains a context object for running the orchestration
	request := &StartOrchestrationRequest{}
	err := this.engine.UnmarshalJSON(req, request)
	if err != nil {
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, "bad-json", http.StatusBadRequest)
		return
	}

	orc, err := this.orchestrate.StartOrchestration(c, domain, orchestration, request.Context, request.Note)
	if err != nil {
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, "cannot-start-orchestration", http.StatusInternalServerError)
		return
	}

	response := &StartOrchestrationResponse{
		Id:        orc.Info().Id,
		StartTime: orc.Info().StartTime.Unix(),
		LogWsUrl: fmt.Sprintf("/v1/ws/feed/%s/%s/%s/%s",
			domain_class, domain_instance, orc.Model().GetName(), orc.Info().Id),
		Context: request.Context,
		Note:    request.Note,
	}
	err = this.engine.MarshalJSON(req, response, resp)
	if err != nil {
		this.engine.HandleError(resp, req, "malformed-result", http.StatusInternalServerError)
		return
	}
}

func (this *Api) WatchOrchestration(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	glog.Infoln("WatchOrchestration")
	this.ws_run_script(resp, req, "timeline1")
}

func (this *Api) watchOrchestrationReal(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	c := this.CreateServiceContext(context, req)

	domain_class := c.UrlParameter("domain_class")
	domain_instance := c.UrlParameter("domain_instance")
	domain := fmt.Sprintf("%s.%s", domain_instance, domain_class)

	orchestration := c.UrlParameter("orchestration")
	instance_id := c.UrlParameter("instance_id")

	glog.Infoln("Orchestration=", orchestration, "InstanceId=", instance_id, "Domain=", domain)

	orc, err := this.orchestrate.GetOrchestration(c, domain, orchestration, instance_id)
	switch {
	case err == ErrNotFound:
		this.engine.HandleError(resp, req, "not-found", http.StatusNotFound)
		return
	case err != nil:
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, "malformed-result", http.StatusInternalServerError)
		return
	}

	topic := orc.Log()
	if topic == nil {
		this.engine.HandleError(resp, req, "no-feed", http.StatusBadRequest)
		return
	}

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
		in := pubsub.GetReader(*topic, sub)
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
