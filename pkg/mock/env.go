package mock

import (
	. "github.com/infradash/redpill/pkg/api"
)

func ListEnvs(domainClass string) ([]Env, error) {
	return []Env{
		Env{
			Domain:    domainClass,
			Service:   "blinker",
			Instances: []string{"dev", "staging", "production"},
			Versions:  []string{"develop", "v1.0", "v1.1"},
			Live: map[string]string{
				"dev":        "develop",
				"staging":    "v1.1",
				"production": "v1.0",
			},
		},
	}, nil
}
