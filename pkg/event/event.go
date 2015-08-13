package event

import (
	"encoding/json"
	"github.com/golang/glog"
	"github.com/infradash/redpill/pkg/api"
	"github.com/qorio/omni/common"
)

type Event struct {
	Status      string `json:"status"`
	Title       string `json:"title,omitempty"`
	Description string `json:"description,omitempty"`
	Note        string `json:"note,omitempty"`
	User        string `json:"user,omitempty"`
	Type        string `json:"type,omitempty"`
	Url         string `json:"url,omitempty"`
	Timestamp   int64  `json:"timestamp,omitempty"`
	ObjectId    string `json:"object_id"`
	ObjectType  string `json:"object_type"`
}

func (this *Event) IsEvent(other interface{}) bool {
	return common.TypeMatch(this, other)
}

func (this *Event) Marshal() ([]byte, error) {
	return json.Marshal(this)
}

type Service struct {
	feed func() <-chan Event
}

func NewService(feed func() <-chan Event) api.EventService {
	s := new(Service)
	s.feed = feed
	return s
}

func (this *Service) EventFeed() <-chan api.Event {
	glog.Infoln("EventFeed")
	out := make(chan api.Event)
	go func() {
		for {
			select {
			case event := <-this.feed():
				out <- &event
			}
		}
	}()
	return out

}
