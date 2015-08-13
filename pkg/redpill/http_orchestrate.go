package redpill

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/gorilla/websocket"
	. "github.com/infradash/redpill/pkg/api"
	"github.com/qorio/maestro/pkg/pubsub"
	"github.com/qorio/omni/auth"
	"net/http"
)

func (this *Api) ListOrchestrations(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	c := this.CreateServiceContext(context, req)
	domain_class := c.UrlParameter("domain_class")
	domain_instance := c.UrlParameter("domain_instance")
	glog.Infoln("DomainClass=", domain_class, "DomainInstance=", domain_instance)

	available, err := this.orchestrate.ListOrchestrations(c, domain_class)

	list := OrchestrationList{}
	for _, o := range available {
		list = append(list, OrchestrationDescription{
			Name:         o.GetName(),
			Label:        o.GetFriendlyName(),
			Description:  o.GetDescription(),
			DefaultInput: o.GetDefaultContext(),
			ActivateUrl:  fmt.Sprintf("/v1/orchestrate/%s/%s/%s", domain_class, domain_instance, o.GetName()),
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

	orc, err := this.orchestrate.StartOrchestration(c, domain_class, domain_instance, orchestration, request.Context, request.Note)
	if err != nil {
		glog.Warningln("Err=", err)
		this.engine.HandleError(resp, req, "cannot-start-orchestration", http.StatusInternalServerError)
		return
	}

	response := &StartOrchestrationResponse{
		Id:        orc.Info().Id,
		StartTime: orc.Info().StartTime.Unix(),
		LogWsUrl: fmt.Sprintf("/v1/ws/orchestrate/%s/%s/%s/%s",
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

func (this *Api) GetOrchestrationInstance(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	glog.Infoln("GetOrchestrationInstance")
	c := this.CreateServiceContext(context, req)
	domain_class := c.UrlParameter("domain_class")
	domain_instance := c.UrlParameter("domain_instance")
	domain := fmt.Sprintf("%s.%s", domain_instance, domain_class)
	orchestration := c.UrlParameter("orchestration")
	instance_id := c.UrlParameter("instance_id")

	glog.Infoln("Domain=", domain, "Orchestration=", orchestration, "Instance=", instance_id)
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
	err = this.engine.MarshalJSON(req, orc, resp)
	if err != nil {
		this.engine.HandleError(resp, req, "malformed-orchestration-instance", http.StatusInternalServerError)
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
