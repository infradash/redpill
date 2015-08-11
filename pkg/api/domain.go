package api

type Domain struct {
	Id    string `json:"id"`
	Class string `json:"class"`
	Name  string `json:"name"`
	Url   string `json:"url"`
}

type DomainDetail struct {
	Id        string   `json:"id"`
	Class     string   `json:"class"`
	Name      string   `json:"name"`
	Instances []string `json:"instances"`
}

type DomainService interface {
	ListDomains(c Context) ([]Domain, error)
	GetDomain(c Context, domainClass string) (*DomainDetail, error)
}
