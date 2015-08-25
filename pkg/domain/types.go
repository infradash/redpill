package domain

import (
	"github.com/qorio/omni/common"
)

type Info struct {
	Id    string `json:"id"`
	Class string `json:"class"`
	Name  string `json:"name"`
	Url   string `json:"url"`
}

func (this *Info) IsDomainInfo(other interface{}) bool {
	return common.TypeMatch(this, other)
}

type Model struct {
	Info
	Instances []string `json:"instances"`
	Services  []string `json:"services"`
}

func (this *Model) IsDomainModel(other interface{}) bool {
	return common.TypeMatch(this, other)
}

func (this *Model) DomainClass() string {
	return this.Class
}

func (this *Model) DomainInstances() []string {
	return this.Instances
}
