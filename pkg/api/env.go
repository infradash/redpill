package api

type Env struct {
	Domain    string            `json:"domain"`
	Service   string            `json:"service"`
	Instances []string          `json:"instances"`
	Versions  []string          `json:"versions"`
	Live      map[string]string `json:"live"`
}

type EnvList map[string]interface{}
type EnvChange struct {
	Update EnvList  `json:"update,omitempty"`
	Delete []string `json:"delete,omitempty"`
}
