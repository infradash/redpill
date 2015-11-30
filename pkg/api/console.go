package api

import (
	_ "github.com/qorio/maestro/pkg/mqtt"
	"github.com/qorio/maestro/pkg/pubsub"
)

type Console struct {
	Id     string
	Input  pubsub.Topic
	Output pubsub.Topic
}

type ConsoleService interface {
	ListConsoles(c Context, domainClass, domainInstance string) (map[string][]string, error)
	GetConsole(c Context, domainClass, domainInstance, service, id string) (*Console, error)
}
