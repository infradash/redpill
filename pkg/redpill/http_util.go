package redpill

import (
	"fmt"
	"github.com/golang/glog"
	_ "github.com/qorio/maestro/pkg/mqtt"
	"github.com/qorio/maestro/pkg/pubsub"
	"github.com/qorio/omni/auth"
	"io"
	"net/http"
	"strings"
	"time"
)

// Format:  mqtt:host:port/topic
func mqtt_topic(topicStr string) string {
	s := strings.Replace(topicStr, "mqtt:", "mqtt://", 1)
	return strings.Replace(s, "*", "#", 1)
}

func (this *Api) UtilTopicSubscribe(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	request := this.CreateServiceContext(context, req)
	topicStr := request.UrlParameter("topic")

	topicStr = mqtt_topic(topicStr)
	glog.Infoln("UtilTopicSubscribe", topicStr, req.RemoteAddr)

	topic := pubsub.Topic(topicStr)
	if !topic.Valid() {
		http.Error(resp, "bad-topic:"+topicStr, http.StatusBadRequest)
		return
	}

	pubsub, err := topic.Broker().PubSub(fmt.Sprintf("%d", time.Now().Unix()))
	if err != nil {
		glog.Warningln("Err", err)
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}

	defer pubsub.Close()

	// get stream
	stream, err := this.engine.DirectHttpStream(resp, req)
	if err != nil {
		glog.Warningln("Err", err)
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}

	glog.V(100).Infoln("Connecting to", topic)
	source, err := pubsub.Subscribe(topic)
	if err != nil {
		glog.Warningln("Err", err)
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}
	glog.V(100).Infoln("Connected to", topic)

	// Listen to the closing of the http connection via the CloseNotifier
	quit := make(chan int)
	go func() {
		for {
			select {
			case <-quit:
				glog.Infoln("Client disconnected")
				return
			case m, open := <-source:
				if !open {
					glog.V(100).Infoln("Source closed. Exiting")
					return
				} else {
					stream <- m
				}
			}
		}
	}()

	// Listen to the closing of the http connection via the CloseNotifier
	notify := resp.(http.CloseNotifier).CloseNotify()

	<-notify // block here
	glog.V(100).Infoln("HTTP connection just closed.")
	quit <- 1
	close(stream)
	glog.V(100).Infoln("HTTP stream completed.")
}

func (this *Api) UtilTopicPublish(context auth.Context, resp http.ResponseWriter, req *http.Request) {
	request := this.CreateServiceContext(context, req)
	topicStr := request.UrlParameter("topic")

	glog.Infoln("UtilTopicPublish", topicStr)

	topic := pubsub.Topic(mqtt_topic(topicStr))
	if !topic.Valid() {
		http.Error(resp, "bad-topic:"+topicStr, http.StatusBadRequest)
		return
	}

	b, err := topic.Broker().PubSub(req.RemoteAddr)
	if err != nil {
		http.Error(resp, err.Error(), http.StatusInternalServerError)
		return
	}

	publish := pubsub.GetWriter(topic, b)
	io.Copy(publish, req.Body)
}
