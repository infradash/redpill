package mock

import (
	dash "github.com/infradash/dash/pkg/executor"
	. "github.com/infradash/redpill/pkg/api"
	. "github.com/infradash/redpill/pkg/orchestrate"
	"github.com/qorio/maestro/pkg/task"
	"os"
)

var (
	mock_models = []Model{
		Model{
			ExecutorConfig: dash.ExecutorConfig{
				Task: task.Task{
					Name: "provision_instance",
				},
			},
			FriendlyName: "Provision minion instances",
			Description:  "Starts new minion instance and add to the pool for a given environment",
			DefaultContext: OrchestrationContext{
				"image":     "aws-ami-1234",
				"instances": 1,
				"type":      "t1micro",
			},
		},
		Model{
			ExecutorConfig: dash.ExecutorConfig{
				Task: task.Task{
					Name: "blinker_db_migrate",
				},
			},
			FriendlyName: "Run db migration (blinker)",
			Description:  "Run database migration for blinker",
			DefaultContext: OrchestrationContext{
				"retry": false,
			},
		},
		Model{
			ExecutorConfig: dash.ExecutorConfig{
				Task: task.Task{
					Name: "blinker_build_image",
				},
			},
			FriendlyName: "Build Docker image (blinker)",
			Description:  "Build docker image for blinker service",
			DefaultContext: OrchestrationContext{
				"git_repo":   "git@github.com:BlinkerGit/test.git",
				"git_branch": "develop",
				"git_tag":    "release1.0",
			},
		},
	}
)

func init() {
	current_dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	boltdb, err = init_db(current_dir, "mock.db")
	if err != nil {
		panic(err)
	}

	for _, domain := range []string{"ops-dev.blinker.com", "staging.blinker.com", "production.blinker.com"} {
		for _, m := range mock_models {
			save_orchestrate_model(boltdb, domain, &m)
		}
	}
}

type orchestrate_models int

func OrchestrationModelStorage() ModelStorage {
	return orchestrate_models(1)
}

func (this orchestrate_models) Save(domain string, model Model) error {
	return save_orchestrate_model(boltdb, domain, &model)
}

func (this orchestrate_models) Get(domain, name string) (*Model, error) {
	return find_model_for_domain_name(boltdb, domain, name)
}

func (this orchestrate_models) GetModels(domain string) ([]Model, error) {
	return load_models_for_domain(boltdb, domain)
}

type orchestrate_instances int

func OrchestrationInstanceStorage() InstanceStorage {
	return orchestrate_instances(1)
}

func (this orchestrate_instances) Save(instance *Instance) error {
	return save_orchestrate_instance(boltdb, instance)
}

func (this orchestrate_instances) Get(id string) (*Instance, error) {
	return find_instance_by_id(boltdb, id)
}

func (this orchestrate_instances) List(domain, orchestration string) ([]Instance, error) {
	return load_instances_for_domain_orchestration(boltdb, domain, orchestration)
}
