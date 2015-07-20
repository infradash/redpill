package orchestrate

import (
	. "github.com/infradash/redpill/pkg/api"
	"github.com/qorio/maestro/pkg/task"
	"github.com/qorio/maestro/pkg/zk"
	"github.com/qorio/omni/common"
	"time"
)

const (
	EnvZkHosts = "REDPILL_ZK_HOSTS"
)

type Service struct {
	conn zk.ZK
}

func NewService(pool func() zk.ZK) OrchestrateService {
	s := new(Service)
	s.conn = pool()
	return s
}

func (this *Service) ListOrchestrations(c Context, domain string) ([]Orchestration, error) {
	return []Orchestration{
		Orchestration{
			Orchestration: task.Orchestration{
				Name:        "provision_instance",
				Label:       "Provision minion instances",
				Description: "Starts new minion instance and add to the pool for a given environment",
			},
			ActivateUrl: "/v1/orchestrate/" + domain + "/provision_instance",
			DefaultInput: map[string]interface{}{
				"image":     "aws-ami-1234",
				"instances": 1,
				"type":      "t1micro",
			},
		},
		Orchestration{
			Orchestration: task.Orchestration{
				Name:        "blinker_db_migrate",
				Label:       "Run db migration (blinker)",
				Description: "Run database migration for blinker",
			},
			ActivateUrl: "/v1/orchestrate/" + domain + "/blinker_db_migrate",
			DefaultInput: map[string]interface{}{
				"retry": false,
			},
		},
		Orchestration{
			Orchestration: task.Orchestration{
				Name:        "blinker_build_image",
				Label:       "Build Docker image (blinker)",
				Description: "Build docker image for blinker service",
			},
			ActivateUrl: "/v1/orchestrate/" + domain + "/blinker_build_image",
			DefaultInput: map[string]interface{}{
				"git_repo":   "git@github.com:BlinkerGit/test.git",
				"git_branch": "develop",
				"git_tag":    "release1.0",
			},
		},
	}, nil
}

func (this *Service) ListRunningOrchestrations(c Context, domain string) ([]Orchestration, error) {
	return []Orchestration{}, nil
}

func (this *Service) StartOrchestration(c Context, domain, orchestration string, input map[string]interface{}) (*Orchestration, error) {
	now := time.Now()
	return &Orchestration{
		Orchestration: task.Orchestration{
			Id:        common.NewUUID().String(),
			StartTime: &now,
		},
	}, nil
}

func (this *Service) GetOrchestration(c Context, domain, orchestration, instance string) (*Orchestration, error) {
	return nil, ErrNotFound
}
