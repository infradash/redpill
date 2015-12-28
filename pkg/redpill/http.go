package redpill

import (
	"bufio"
	"encoding/json"
	"fmt"
	"github.com/golang/glog"
	"github.com/gorilla/websocket"
	. "github.com/infradash/redpill/pkg/api"
	"github.com/qorio/maestro/pkg/pubsub"
	"github.com/qorio/omni/auth"
	"github.com/qorio/omni/rest"
	"github.com/qorio/omni/version"
	"io"
	"net/http"
	"net/http/pprof"
	"os/exec"
	"path/filepath"
	"runtime"
	pp "runtime/pprof"
	"strconv"
	"strings"
	"sync"
	"time"
)

type Api struct {
	options     Options
	authService auth.Service
	engine      rest.Engine

	event                EventService
	env                  EnvService
	domain               DomainService
	registry             RegistryService
	orchestrate          OrchestrateService
	conf                 ConfService
	pkg                  PkgService
	dockerapi            DockerProxyService
	console              ConsoleService
	CreateServiceContext CreateContextFunc
}

var ServiceId = "redpill"

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

type handler string

func (name handler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "text/plain; charset=utf-8")
	debug, _ := strconv.Atoi(r.FormValue("debug"))
	p := pp.Lookup(string(name))
	if p == nil {
		w.WriteHeader(404)
		fmt.Fprintf(w, "Unknown profile: %s\n", name)
		return
	}
	gc, _ := strconv.Atoi(r.FormValue("gc"))
	if name == "heap" && gc > 0 {
		runtime.GC()
	}
	p.WriteTo(w, debug)
	return
}

func NewApi(options Options, auth auth.Service,
	env EnvService,
	domain DomainService,
	event EventService,
	registry RegistryService,
	orchestrate OrchestrateService,
	conf ConfService,
	pkg PkgService,
	console ConsoleService,
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
		console:     console,
	}

	ep.CreateServiceContext = ServiceContext(ep.engine)

	ep.engine.Handle("/debug/pprof/", http.HandlerFunc(pprof.Index))

	for _, pf := range pp.Profiles() {
		h := handler(pf.Name())
		ep.engine.Handle("/debug/pprof/"+pf.Name(), h)
	}

	ep.engine.Handle("/debug/pprof/cmdline", http.HandlerFunc(pprof.Cmdline))
	ep.engine.Handle("/debug/pprof/profile", http.HandlerFunc(pprof.Profile))
	ep.engine.Handle("/debug/pprof/symbol", http.HandlerFunc(pprof.Symbol))
	ep.engine.Handle("/debug/pprof/trace", http.HandlerFunc(pprof.Trace))

	ep.engine.Bind(

		rest.SetHandler(Methods[ServerInfo], ep.GetServerInfo),
		rest.SetHandler(Methods[RunScript], ep.WsRunScript),
		rest.SetHandler(Methods[EventFeed], ep.WsEventFeed),
		rest.SetHandler(Methods[SubscribeTopic], ep.WsSubscribeTopic),
		rest.SetHandler(Methods[DuplexTopic], ep.WsDuplexTopic),

		rest.SetAuthenticatedHandler(ServiceId, Methods[LogFeed], ep.LogFeed),

		rest.SetAuthenticatedHandler(ServiceId, Methods[PrototypeRunScript], ep.PrototypeRunScript),
		rest.SetAuthenticatedHandler(ServiceId, Methods[PrototypeEventFeed], ep.PrototypeEventFeed),

		rest.SetAuthenticatedHandler(ServiceId, Methods[PrototypeListConsoles], ep.PrototypeListConsoles),
		rest.SetAuthenticatedHandler(ServiceId, Methods[PrototypeConnectConsole], ep.PrototypeConnectConsole),

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

		// Util
		rest.SetAuthenticatedHandler(ServiceId, Methods[UtilTopicSubscribe], ep.UtilTopicSubscribe),
		rest.SetAuthenticatedHandler(ServiceId, Methods[UtilTopicPublish], ep.UtilTopicPublish),
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
		message, err := json.Marshal(event)
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

func (this *Api) WsSubscribeTopic(resp http.ResponseWriter, req *http.Request) {
	queries, err := this.engine.GetUrlQueries(req, Methods[SubscribeTopic].UrlQueries)
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

func (this *Api) PrototypeListConsoles(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	start := time.Now()
	defer func() { glog.Infoln("Elapsed", time.Now().Sub(start).Nanoseconds()) }()

	request := this.CreateServiceContext(context, req)
	domainClass := request.UrlParameter("domain_class")
	domainInstance := request.UrlParameter("domain_instance")
	glog.Infoln("ListConsoles:", domainClass, domainInstance)

	result, err := this.console.ListConsoles(request, domainClass, domainInstance)
	if err != nil {
		this.engine.HandleError(resp, req, err.Error(), http.StatusInternalServerError)
		return
	}
	err = this.engine.MarshalJSON(req, result, resp)
	if err != nil {
		this.engine.HandleError(resp, req, "malformed-response", http.StatusInternalServerError)
		return
	}

}

func (this *Api) PrototypeConnectConsole(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	request := this.CreateServiceContext(context, req)
	domainClass := request.UrlParameter("domain_class")
	domainInstance := request.UrlParameter("domain_instance")
	service := request.UrlParameter("service")
	id := request.UrlParameter("id")
	console, err := this.console.GetConsole(request, domainClass, domainInstance, service, id)
	if err != nil {
		this.engine.HandleError(resp, req, err.Error(), http.StatusInternalServerError)
		return
	}
	glog.Infoln("ConnectConsole", "console=", console)
	this.ws_connect_mqtt_topics(console.Input, console.Output, resp, req)
}

func (this *Api) WsDuplexTopic(resp http.ResponseWriter, req *http.Request) {
	glog.Infoln("DISABLED TEMPORARILY")
	if true {
		this.engine.HandleError(resp, req, "not-implemented", http.StatusNotImplemented)
		return
	}

	queries, err := this.engine.GetUrlQueries(req, Methods[DuplexTopic].UrlQueries)
	if err != nil {
		this.engine.HandleError(resp, req, "error-bad-request", http.StatusBadRequest)
		return
	}

	topic := queries["topic"].(string)
	glog.Infoln("Topic = ", topic)

	backend := queries["backend"].(bool)
	glog.Infoln("Backend = ", backend)

	topicIn := pubsub.Topic(topic + ".in")
	topicOut := pubsub.Topic(topic + ".out")
	if backend {
		// reverse the in and out
		topicIn = pubsub.Topic(topic + ".out")
		topicOut = pubsub.Topic(topic + ".in")
	}
	this.ws_connect_mqtt_topics(topicIn, topicOut, resp, req)
}

func (this *Api) ws_connect_mqtt_topics(topicIn, topicOut pubsub.Topic, resp http.ResponseWriter, req *http.Request) {

	glog.Infoln("Connecting ws to topic:", topicIn, "and", topicOut)
	if !topicIn.Valid() {
		glog.Warningln("Topic", topicIn, "is not valid")
		this.engine.HandleError(resp, req, "bad-topic", http.StatusBadRequest)
		return
	}
	if !topicOut.Valid() {
		glog.Warningln("Topic", topicOut, "is not valid")
		this.engine.HandleError(resp, req, "bad-topic", http.StatusBadRequest)
		return
	}

	conn, err := upgrader.Upgrade(resp, req, nil)
	if err != nil {
		glog.Infoln("ERROR", err)
		return
	}

	glog.Infoln("Topic using broker", topicIn.Broker())
	clientIn, err := topicIn.Broker().PubSub("test")
	if err != nil {
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, err.Error(), http.StatusInternalServerError)
		return
	}
	// Process incoming messages from the connection
	go func() {
		defer conn.Close()
		out := pubsub.GetWriter(topicIn, clientIn)
		for {
			_, in, err := conn.NextReader()
			if err != nil {
				report_error(conn, err, "ws read error")
				break
			}
			_, err = io.Copy(out, in)
			if err != nil {
				report_error(conn, err, "publish write error")
				break
			}
		}
		glog.Infoln("Completed")
	}()

	glog.Infoln("Topic using broker", topicOut.Broker())
	clientOut, err := topicOut.Broker().PubSub("test")
	if err != nil {
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, err.Error(), http.StatusInternalServerError)
		return
	}
	// Read from topic and write out to connection
	go func() {
		defer conn.Close()
		in := pubsub.GetReader(topicOut, clientOut)
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

var mock_events = map[string]chan interface{}{}

func event_source(key string) chan interface{} {
	if m, has := mock_events[key]; !has {
		messages := make(chan interface{})
		mock_events[key] = messages
		go func() {
			for i := 0; ; i++ {
				if strings.Index(key, ".json") > 0 {
					messages <- map[string]interface{}{
						"key":     key,
						"message": "this is from " + key,
					}
				} else {
					messages <- fmt.Sprintf("%d, %s", i, key)
				}
				time.Sleep(1 * time.Second)
				if i == 45 {
					glog.Infoln(">>> Time's up. Closing event source")
					messages <- nil
					close(messages)
					return
				}
			}
		}()
		return messages
	} else {
		return m
	}
}

func (this *Api) LogFeed(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	request := this.CreateServiceContext(context, req)

	domainClass := request.UrlParameter("domain_class")
	domainInstance := request.UrlParameter("domain_instance")
	target := request.UrlParameter("target")

	key := req.URL.Path
	glog.Infoln("LogFeed:", domainClass, domainInstance, target, "key=", key)
	messages := event_source(key)
	if strings.Index(target, ".json") > 0 {
		this.engine.MergeHttpStream(resp, req, "application/json", "TestEvent", key, messages)
	} else {
		this.engine.MergeHttpStream(resp, req, "text/plain", "text/plain", key, messages)
	}
}

func (this *Api) PrototypeEventFeed(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	request := this.CreateServiceContext(context, req)

	glog.Infoln("PrototypeEventFeed", request)

	this.engine.MergeHttpStream(resp, req, "application/json", "Event", "event", this.event.EventFeed())
}

func (this *Api) PrototypeRunScript(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	request := this.CreateServiceContext(context, req)
	script := request.UrlParameter("script")

	glog.Infoln("PrototypeRunScript", script)

	output, err := this.exec_script(script)
	if err != nil {
		this.engine.HandleError(resp, req, err.Error(), http.StatusInternalServerError)
		return
	}
	this.engine.MergeHttpStream(resp, req, "text/plain", "text/plain", script, output)
}

func (this *Api) exec_script(script string) (<-chan interface{}, error) {
	script_file := filepath.Join(this.options.WorkingDir, "scripts", script)
	glog.Infoln("Running script:", script, "file=", script_file)

	command, err := exec.LookPath(script_file)
	if err != nil {
		return nil, err
	}

	cmd := exec.Command(command)
	glog.Infoln("Running command", *cmd)

	stdout, err := cmd.StdoutPipe()
	if err != nil {
		return nil, err
	}

	stderr, err := cmd.StderrPipe()
	if err != nil {
		return nil, err
	}

	output := make(chan interface{})

	go func() {
		scanner := bufio.NewScanner(io.MultiReader(stdout, stderr))
		for scanner.Scan() {
			output <- scanner.Bytes()
		}
	}()

	go func() {
		glog.Infoln(">>>>> Running", *cmd)
		err = cmd.Run()
		if err != nil {
			output <- "===> error:" + err.Error()
			return
		}
		glog.Infoln("Completed", *cmd)
	}()

	return output, nil
}
