package pkg

import (
	"github.com/qorio/maestro/pkg/docker"
	"github.com/qorio/omni/common"
)

type pkg struct {
	DockerImageUrl *string `json:"docker_image"`
}

func (this pkg) IsPkg(other interface{}) bool {
	return common.TypeMatch(this, other)
}

func (this pkg) AsDockerImage() *docker.Image {
	if this.DockerImageUrl == nil {
		return nil
	}
	image := docker.ParseImageUrl(*this.DockerImageUrl)
	return &image
}
