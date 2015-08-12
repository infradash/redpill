package env

import (
	"errors"
	"fmt"
	"github.com/golang/glog"
	. "github.com/infradash/redpill/pkg/api"
	"github.com/qorio/maestro/pkg/zk"
	"path/filepath"
	"strings"
)

const (
	EnvZkHosts = "REDPILL_ZK_HOSTS"
)

var (
	ErrNoEnv = errors.New("error-no-envs")
)

type Service struct {
	conn    zk.ZK
	domains DomainService
}

func NewService(pool func() zk.ZK, domains DomainService) EnvService {
	s := new(Service)
	s.conn = pool()
	s.domains = domains
	return s
}

/// TODO -- this is actually not very accurate to use only the Cversion, which is a version
/// number associated with the number of children of a znode.  A true version should be
/// calculated based on the content of the children or some hash of all the children's versions
func (this *Service) calculate_rev_from_parent(zn *zk.Node) Revision {
	return Revision(zn.Stats.Cversion)
}

type service_stat struct {
	instances map[string]int
	versions  map[string]int
	live      map[string]string
}

func (this *service_stat) add_instance(instance string) {
	this.instances[instance] += 1
}

func (this *service_stat) get_instances() []string {
	l := []string{}
	for k, _ := range this.instances {
		l = append(l, k)
	}
	return l
}

func (this *service_stat) add_version(version string) {
	this.versions[version] += 1
}

func (this *service_stat) get_versions() []string {
	l := []string{}
	for k, _ := range this.versions {
		l = append(l, k)
	}
	return l
}

// ex: /integration.foo.com/svc/integration/container/docker/img:integration-7049.1317,/integration.foo.com/svc/integration/env ==> should be 'integration'
func (this *service_stat) set_live(instance, live string) {
	p := strings.Split(live, ",")
	envPath := p[len(p)-1]
	this.live[instance] = filepath.Base(filepath.Dir(envPath))
}

func (this *Service) ListDomainEnvs(c Context, domainClass string) ([]Env, error) {
	model, err := this.domains.GetDomain(c, domainClass)
	if err != nil {
		return nil, err
	}
	// Now we have the environments.  Construct full domain name.
	// In ZK, the services are children of domain znodes.
	// Versions are children of service znodes.
	// By looking at the live version data, we can render the details about the domain
	glog.Infoln("DomainDetail=", model)

	// collect information by service
	service_stats := map[string]*service_stat{}

	// Build the fully qualified name for each domain
	for _, domainInstance := range model.DomainInstances() {
		// Get the services
		p := fmt.Sprintf("/%s.%s", domainInstance, domainClass)
		zdomain, err := this.conn.Get(p)
		if err != nil {
			glog.Warningln("Err=", err)
			return nil, err
		}
		zservices, err := zdomain.Children()
		if err != nil {
			glog.Warningln("Err=", err)
			return nil, err
		}
		// get the versions
		for _, zservice := range zservices {
			service := zservice.GetBasename()
			if _, has := service_stats[service]; !has {
				service_stats[service] = &service_stat{
					instances: map[string]int{},
					versions:  map[string]int{},
					live:      map[string]string{},
				}
			}
			// an instance
			service_stats[service].add_instance(domainInstance)

			zversions, err := zservice.Children()
			if err != nil {
				glog.Warningln("Err=", err)
				return nil, err
			}
			for _, zversion := range zversions {
				if zversion.GetBasename() == "live" {
					// get the current live information
					service_stats[service].set_live(domainInstance, zversion.GetValueString())
				} else {
					// a version
					service_stats[service].add_version(zversion.GetBasename())
				}
			}
		}
	}

	envs := []Env{}
	// Now generate the metadata output based on the stats
	for service, stats := range service_stats {
		envs = append(envs, Env{
			Domain:    domainClass,
			Service:   service,
			Instances: stats.get_instances(),
			Versions:  stats.get_versions(),
			Live:      stats.live,
		})
	}
	return envs, nil //this.listDomainEnvs(domainClass)
}

// EnvService
func (this *Service) GetEnv(c Context, domain, service, version string) (EnvList, Revision, error) {
	key := fmt.Sprintf("/%s/%s/%s/env", domain, service, version)
	glog.Infoln("GetEnv:", c.UserId(), "Domain=", domain, "Service=", service, "Version=", version, "Key=", key)
	zn, err := this.conn.Get(key)
	switch {
	case err == zk.ErrNotExist:
		return nil, Revision(-1), ErrNoEnv
	case err != nil:
		return nil, -1, err
	}

	list := EnvList{}
	_, err = zn.VisitChildrenRecursive(func(n *zk.Node) bool {
		if n.IsLeaf() {
			list[n.GetBasename()] = n.GetValueString()
		}
		return false
	})
	return list, this.calculate_rev_from_parent(zn), nil
}

// EnvService
func (this *Service) NewEnv(c Context, domain, service, version string, vars *EnvList) (Revision, error) {
	glog.Infoln("NewEnv:", c.UserId(), "Domain=", domain, "Service=", service, "Version=", version)

	root := fmt.Sprintf("/%s/%s/%s/env", domain, service, version)

	_, err := this.conn.Get(root)
	switch {
	case err == nil:
		return -1, ErrConflict
	case err != zk.ErrNotExist:
		return -1, err
	case err == zk.ErrNotExist:
		// continue
	}

	// everything ok. commit changes.  Note this is not atomic!
	for key, create := range *vars {
		k := fmt.Sprintf("%s/%s", root, key)
		v := fmt.Sprintf("%s", create)
		_, err := this.conn.Create(k, []byte(v))
		if err != nil {
			return -1, err
		}
	}

	zn, err := this.conn.Get(root)
	switch {
	case err != nil:
		return -1, err
	}
	return this.calculate_rev_from_parent(zn), nil
}

func (this *Service) SaveEnv(c Context, domain, service, version string, change *EnvChange, rev Revision) error {
	glog.Infoln("SaveEnv:", c.UserId(), "Domain=", domain, "Service=", service, "Version=", version, "Rev=", rev)

	root := fmt.Sprintf("/%s/%s/%s/env", domain, service, version)
	if zn, err := this.conn.Get(root); err != nil && err != zk.ErrNotExist {
		return err
	} else if this.calculate_rev_from_parent(zn) != rev {
		return ErrConflict
	}

	creates := []string{}
	updates := []struct {
		Node  *zk.Node
		Value []byte
	}{}
	for key, update := range change.Update {
		n, err := this.conn.Get(fmt.Sprintf("%s/%s", root, key))
		v := fmt.Sprintf("%s", update)
		switch {
		case err == zk.ErrNotExist:
			creates = append(creates, v)
		case err != nil:
			return err
		default:
			updates = append(updates, struct {
				Node  *zk.Node
				Value []byte
			}{n, []byte(v)})
		}
	}

	deletes := []*zk.Node{}
	for _, delete := range change.Delete {
		k := fmt.Sprintf("%s/%s", root, delete)
		n, err := this.conn.Get(k)
		switch {
		case err == zk.ErrNotExist:
		case err != nil:
			return err
		default:
			deletes = append(deletes, n)
		}
	}

	// everything ok. commit changes.  Note this is not atomic!
	for key, create := range creates {
		k := fmt.Sprintf("%s/%s", root, key)
		v := fmt.Sprintf("%s", create)
		_, err := this.conn.Create(k, []byte(v))
		if err != nil {
			return err
		}
	}
	for _, update := range updates {
		err := update.Node.Set(update.Value)
		if err != nil {
			return err
		}
	}
	for _, delete := range deletes {
		err := delete.Delete()
		if err != nil {
			return err
		}
	}
	return nil
}
