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

	env      EnvService
	domain   DomainService
	registry RegistryService

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
	registry RegistryService) (*Api, error) {
	ep := &Api{
		options:     options,
		authService: auth,
		engine:      rest.NewEngine(&Methods, auth, nil),
		env:         env,
		domain:      domain,
		registry:    registry,
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
		rest.SetAuthenticatedHandler(ServiceId, Methods[GetEnvironmentVars], ep.GetEnvironmentVars),
		rest.SetAuthenticatedHandler(ServiceId, Methods[UpdateEnvironmentVars], ep.UpdateEnvironmentVars),

		// Registry
		rest.SetAuthenticatedHandler(ServiceId, Methods[GetRegistry], ep.GetRegistry),
		rest.SetAuthenticatedHandler(ServiceId, Methods[UpdateRegistry], ep.UpdateRegistry),
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
}

func (this *Api) WsRunScript(resp http.ResponseWriter, req *http.Request) {
	conn, err := upgrader.Upgrade(resp, req, nil)
	if err != nil {
		glog.Infoln("ERROR", err)
		return
	}

	script := this.engine.GetUrlParameter(req, "script")
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
	case err != nil:
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, "save-env-fails", http.StatusInternalServerError)
	}
}

func (this *Api) GetRegistry(context auth.Context, resp http.ResponseWriter, req *http.Request) {

}

func (this *Api) UpdateRegistry(context auth.Context, resp http.ResponseWriter, req *http.Request) {

}
