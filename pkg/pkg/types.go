package pkg

import (
	"encoding/json"
	. "github.com/infradash/redpill/pkg/api"
	"github.com/qorio/maestro/pkg/docker"
	"github.com/qorio/omni/common"
)

type pkg struct {
	DockerImageUrl *string `json:"docker_image"`
}

func (this *pkg) FromBytes(d []byte) error {
	return json.Unmarshal(d, this)
}

func (this pkg) IsPkgModel(other interface{}) bool {
	return common.TypeMatch(this, other)
}

func (this pkg) AsDockerImage() *docker.Image {
	if this.DockerImageUrl == nil {
		return nil
	}
	image := docker.ParseImageUrl(*this.DockerImageUrl)
	return &image
}

type pkgInfo struct {
	Domain    string              `json:"domain"`
	Service   string              `json:"service"`
	Instances []string            `json:"instances"`
	Versions  []string            `json:"versions"`
	Live      map[string]PkgModel `json:"live"`
}

func (this pkgInfo) IsPkg(other interface{}) bool {
	return common.TypeMatch(this, other)
}
