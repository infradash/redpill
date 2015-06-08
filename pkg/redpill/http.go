package redpill

import (
	"bufio"
	"github.com/golang/glog"
	"github.com/gorilla/websocket"
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
	service     Service
}

var ServiceId = "redpill"

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin:     func(r *http.Request) bool { return true },
}

func NewApi(options Options, auth auth.Service, service Service) (*Api, error) {
	ep := &Api{
		options:     options,
		authService: auth,
		engine:      rest.NewEngine(&Methods, auth, nil),
		service:     service,
	}

	ep.engine.Bind(
		rest.SetHandler(Methods[Info], ep.GetInfo),
		rest.SetHandler(Methods[RunScript], ep.WsRunScript),
		rest.SetHandler(Methods[EventsFeed], ep.WsEventsFeed),

		rest.SetAuthenticatedHandler(ServiceId, Methods[GetEnvironmentVars], ep.GetEnvironmentVars),
		rest.SetAuthenticatedHandler(ServiceId, Methods[UpdateEnvironmentVars], ep.UpdateEnvironmentVars),
		rest.SetAuthenticatedHandler(ServiceId, Methods[GetRegistry], ep.GetRegistry),
		rest.SetAuthenticatedHandler(ServiceId, Methods[UpdateRegistry], ep.UpdateRegistry),
	)
	return ep, nil
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

func (this *Api) GetEnvironmentVars(context auth.Context, resp http.ResponseWriter, req *http.Request) {

}

func (this *Api) UpdateEnvironmentVars(context auth.Context, resp http.ResponseWriter, req *http.Request) {

}

func (this *Api) GetRegistry(context auth.Context, resp http.ResponseWriter, req *http.Request) {

}

func (this *Api) UpdateRegistry(context auth.Context, resp http.ResponseWriter, req *http.Request) {

}
