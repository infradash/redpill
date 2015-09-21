package redpill

import (
	"github.com/golang/glog"
	_ "github.com/qorio/maestro/pkg/mqtt"
	"github.com/qorio/maestro/pkg/pubsub"
	"github.com/qorio/omni/auth"
	"net/http"
	"strings"
)

func (this *Api) UtilTopicSubscribe(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	request := this.CreateServiceContext(context, req)
	topicStr := request.UrlParameter("topic")

	glog.Infoln("UtilTopicSubscribe", topicStr)

	topicStr = strings.Replace(topicStr, "mqtt:", "mqtt://", 1)
	topicStr = strings.Replace(topicStr, "*", "#", 1)

	topic := pubsub.Topic(topicStr)
	if !topic.Valid() {
		http.Error(resp, "bad-topic:"+topicStr, http.StatusBadRequest)
		return
	}

	pubsub, err := topic.Broker().PubSub(req.RemoteAddr)
	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}

	source, err := pubsub.Subscribe(topic)
	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}

	pipe := make(chan interface{})

	// Listen to the closing of the http connection via the CloseNotifier
	notify := make(chan int)
	go func() {
		for {
			select {
			case <-notify:
				glog.Infoln("Client disconnected")
				return
			case m, open := <-source:
				if !open {
					return
				} else {
					pipe <- m
				}
			}
		}
	}()

	contentType, eventType := "text/plain", "text/plain"
	this.engine.StreamServerEvents(resp, req, contentType, eventType, topicStr, pipe)
	notify <- 1
}

func (this *Api) UtilTopicPublish(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	request := this.CreateServiceContext(context, req)
	topic := request.UrlParameter("topic")

	glog.Infoln("UtilTopicPublish", topic)

}
