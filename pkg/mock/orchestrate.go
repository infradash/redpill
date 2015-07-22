package mock

import (
	. "github.com/infradash/redpill/pkg/api"
	. "github.com/infradash/redpill/pkg/orchestrate"
)

type orchestrate_models int

func OrchestrationModelStorage() ModelStorage {
	return orchestrate_models(1)
}

func (this orchestrate_models) Save(model Model) error {
	return nil
}

func (this orchestrate_models) GetModels(domain string) ([]Model, error) {
	return []Model{
		Model{
			Name:         "provision_instance",
			FriendlyName: "Provision minion instances",
			Description:  "Starts new minion instance and add to the pool for a given environment",
			DefaultContext: OrchestrationContext{
				"image":     "aws-ami-1234",
				"instances": 1,
				"type":      "t1micro",
			},
		},
		Model{
			Name:         "blinker_db_migrate",
			FriendlyName: "Run db migration (blinker)",
			Description:  "Run database migration for blinker",
			DefaultContext: OrchestrationContext{
				"retry": false,
			},
		},
		Model{
			Name:         "blinker_build_image",
			FriendlyName: "Build Docker image (blinker)",
			Description:  "Build docker image for blinker service",
			DefaultContext: OrchestrationContext{
				"git_repo":   "git@github.com:BlinkerGit/test.git",
				"git_branch": "develop",
				"git_tag":    "release1.0",
			},
		},
	}, nil
}
