package aws

import (
	"encoding/json"
)

type Children func() []interface{}

func (c Children) MarshalJSON() ([]byte, error) {
	if c == nil {
		return []byte("\"\""), nil
	}
	return json.Marshal(c())
}

func (c Children) UnmarshalJSON(m []byte) error {
	list := []interface{}{}
	err := json.Unmarshal(m, list)
	if err != nil {
		return err
	}
	c = func() []interface{} {
		return list
	}
	return nil
}

type Node struct {
	Type        string   `json:"type"`
	Id          string   `json:"id"`
	Name        string   `json:"name"`
	Description string   `json:"description"`
	Children    Children `json:"children"`
}

type VirtualPrivateCloud struct {
	Node

	Region        string `json:"region"`
	State         string `json:"state"`
	CIDR          string `json:"cidr"`
	Tenancy       string `json:"tenancy"`
	DNSResolution bool   `json:"dns_resolution"`
	DNSHostnames  bool   `json:"dns_hostnames"`

	Tags map[string]string `json:"tags"`
}

type AvailabilityZone struct {
	Node

	Status string `json:"status"`
}

type Subnet struct {
	Node

	Status  string `json:"status"`
	CIDR    string `json:"cidr"`
	Private bool   `json:"private"`
}

type NetTraffic struct {
	Type      string `json:"type"`
	Protocol  string `json:"protocol"`
	PortRange []int  `json:"port_range"`
}

type Inbound struct {
	NetTraffic
	Source string `json:"source"`
}

type Outbound struct {
	NetTraffic
	Destination string `json:"destination"`
}

type SecurityGroup struct {
	Node

	Inbound  []Inbound  `json:"inbound"`
	Outbound []Outbound `json:"outbound"`
}

type LogStreamUrl string // a ws url

type Host struct {
	Node

	Logs map[string]LogStreamUrl `json:"logs"`
}

type Service struct {
	Node

	Logs map[string]LogStreamUrl `json:"logs"`
}

type Container struct {
	Node

	Logs map[string]LogStreamUrl `json:"logs"`
}

type Process struct {
	Node

	Logs map[string]LogStreamUrl `json:"logs"`
}
