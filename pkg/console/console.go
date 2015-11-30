package console

import (
	"errors"
	"github.com/golang/glog"
	. "github.com/infradash/redpill/pkg/api"
	"github.com/qorio/maestro/pkg/pubsub"
	"github.com/qorio/maestro/pkg/registry"
	"github.com/qorio/maestro/pkg/zk"
)

const (
	EnvZkHosts = "REDPILL_ZK_HOSTS"
)

var (
	ErrBadVarName = errors.New("error-bad-env-var-name")
)

type Service struct {
	conn    zk.ZK
	domains DomainService
}

func NewService(pool func() zk.ZK, domains DomainService) ConsoleService {
	s := new(Service)
	s.conn = pool()
	s.domains = domains
	return s
}

func (this *Service) ListConsoles(c Context, domainClass, domainInstance string) (map[string][]string, error) {
	top := GetConsoleListPath(domainClass, domainInstance)
	glog.Infoln("ConsoleListPath=", top)

	result := map[string][]string{}

	zk.Visit(this.conn, top, func(p registry.Path, data []byte) bool {
		// for each service
		list := []string{}
		zk.Visit(this.conn, p, func(i registry.Path, _ []byte) bool {
			if zk.PathExists(this.conn, i.Sub("running")) {
				list = append(list, i.Base())
			}
			return true
		})
		result[p.Base()] = list
		return true
	})
	return result, nil
}

func (this *Service) GetConsole(c Context, domainClass, domainInstance, service, id string) (*Console, error) {
	consolePath := GetConsolePath(domainClass, domainInstance, service, id)
	glog.Infoln("ConsolePath=", consolePath)

	// get the info from the 'running' node
	running := consolePath.Sub("running")

	m := map[string]interface{}{}

	if err := zk.GetObject(this.conn, running, &m); err != nil {

		glog.Infoln(">>>>", running, err, m)
		return nil, err
	}

	return &Console{
		Id:     id,
		Input:  pubsub.Topic(m["stdin"].(string)),
		Output: pubsub.Topic(m["stdout"].(string)),
	}, nil
}
