package env

import (
	"errors"
	"fmt"
	"github.com/golang/glog"
	. "github.com/infradash/redpill/pkg/api"
	"github.com/qorio/maestro/pkg/registry"
	"github.com/qorio/maestro/pkg/zk"
	"path/filepath"
	"strconv"
	"strings"
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
	// The revision number is simply stored as the value
	str := zn.GetValueString()
	if len(str) == 0 {
		return Revision(0)
	}
	if rev, err := strconv.Atoi(str); err == nil {
		return Revision(rev)
	} else {
		return Revision(0)
	}
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
func (this *service_stat) set_live_legacy(instance, live string) {
	p := strings.Split(live, ",")
	envPath := p[len(p)-1]
	this.live[instance] = filepath.Base(filepath.Dir(envPath))
}

func (this *service_stat) set_live(instance string, live *string) {
	if live == nil {
		return
	}
	this.live[instance] = filepath.Base(filepath.Dir(*live))
}

func (this *Service) ListDomainEnvs(c Context, domainClass string) (map[string]Env, error) {
	model, err := this.domains.GetDomain(c, domainClass)
	if err != nil {
		return nil, err
	}

	// collect information by service
	service_stats := map[string]*service_stat{}

	// Build the fully qualified name for each domain
	for _, domainInstance := range model.DomainInstances() {
		// Get the services
		zdomain, err := this.conn.Get(GetDomainPath(domainClass, domainInstance).Path())
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

				switch zversion.GetBasename() {
				case "live":
				case "_watch", "_live": // skip
				default:
					// a version
					service_stats[service].add_version(zversion.GetBasename())
				}
			}

			// Get the live version
			service_stats[service].set_live(domainInstance,
				zk.GetString(this.conn, GetEnvLivePath(domainClass, domainInstance, service)))

		}
	}

	envs := map[string]Env{}
	// Now generate the metadata output based on the stats
	for service, stats := range service_stats {
		envs[service] = Env{
			Domain:    domainClass,
			Service:   service,
			Instances: stats.get_instances(),
			Versions:  stats.get_versions(),
			Live:      stats.live,
		}
	}
	return envs, nil
}

// EnvService
func (this *Service) GetEnv(c Context, domainClass, domainInstance, service, version string) (EnvList, Revision, error) {
	envPath := GetEnvPath(domainClass, domainInstance, service, version)
	glog.Infoln("GetEnv:", c.UserId(), "Path=", envPath)

	zn, err := this.conn.Get(envPath.Path())
	switch {
	case err == zk.ErrNotExist:
		return nil, Revision(-1), ErrNotFound
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

func validate(vars *EnvList) error {
	if vars == nil {
		return ErrNoChanges
	}
	if len(*vars) == 0 {
		return ErrNoChanges
	}
	for k, _ := range *vars {
		if len(k) == 0 {
			return ErrBadVarName
		}
	}
	return nil
}

func validate_changes(changes *EnvChange) error {
	if changes == nil {
		return ErrNoChanges
	}
	if len(changes.Update) == 0 && len(changes.Delete) == 0 {
		return ErrNoChanges
	}

	for k, _ := range changes.Update {
		if len(k) == 0 {
			return ErrBadVarName
		}
	}
	for _, k := range changes.Delete {
		if len(k) == 0 {
			return ErrBadVarName
		}
	}
	return nil
}

// EnvService
func (this *Service) CreateEnv(c Context, domainClass, domainInstance, service, version string, vars *EnvList) (Revision, error) {
	if err := validate(vars); err != nil {
		return -1, err
	}

	envPath := GetEnvPath(domainClass, domainInstance, service, version)
	glog.Infoln("CreateEnv:", c.UserId(), "Path=", envPath)

	if zk.PathExists(this.conn, envPath) {
		return -1, ErrConflict
	}

	v, err := zk.VersionLockAndExecute(this.conn, envPath, int(0),
		func() error {
			// everything ok. commit changes.  Note this is not atomic.
			// we do our best with double incrementing version numbers.
			for key, create := range *vars {
				k := envPath.Sub(key).Path()
				v := fmt.Sprintf("%s", create)
				_, err := this.conn.Create(k, []byte(v))
				if err != nil {
					return err
				}
			}
			return nil
		})
	return Revision(v), err
}

func (this *Service) UpdateEnv(c Context, domainClass, domainInstance, service, version string, change *EnvChange, rev Revision) (Revision, error) {
	if err := validate_changes(change); err != nil {
		return -1, err
	}

	envPath := GetEnvPath(domainClass, domainInstance, service, version)
	glog.Infoln("UpdateEnv:", c.UserId(), "Path=", envPath)

	if !zk.PathExists(this.conn, envPath) {
		return -1, ErrNotFound
	}

	// Now make changes by acquiring lock
	v, err := zk.VersionLockAndExecute(this.conn, envPath, int(rev),
		func() error {

			_, err := this.conn.Get(envPath.Path())
			if err != nil && err != zk.ErrNotExist {
				return err
			}

			creates := map[string][]byte{}
			updates := map[*zk.Node][]byte{}

			for key, update := range change.Update {
				zkey := envPath.Sub(key).Path()
				n, err := this.conn.Get(zkey)
				v := []byte(fmt.Sprintf("%s", update))
				switch {
				case err == zk.ErrNotExist:
					creates[zkey] = v
				case err != nil:
					return err
				default:
					updates[n] = v
				}
			}

			deletes := []*zk.Node{}
			for _, delete := range change.Delete {
				k := envPath.Sub(delete).Path()
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
				_, err := this.conn.Create(key, create)
				if err != nil {
					return err
				}
			}
			for n, update := range updates {
				err := n.Set(update)
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
		})
	return Revision(v), err
}

func (this *Service) DeleteEnv(c Context, domainClass, domainInstance, service, version string, rev Revision) error {

	envPath := GetEnvPath(domainClass, domainInstance, service, version)
	glog.Infoln("DeleteEnv:", c.UserId(), "Path=", envPath)

	if !zk.PathExists(this.conn, envPath) {
		return ErrNotFound
	}

	// Now make changes by acquiring lock
	_, err := zk.VersionLockAndExecute(this.conn, envPath, int(rev),
		func() error {

			_, err := this.conn.Get(envPath.Path())
			if err != nil && err != zk.ErrNotExist {
				return err
			}

			return zk.DeleteObject(this.conn, envPath)
		})
	return err
}

func (this *Service) legacy_setlive(domainClass, domainInstance, service, version string) error {
	domain := ToDomainName(domainClass, domainInstance)
	livepath := registry.NewPath(domain, service, "live")
	realpath := registry.NewPath(domain, service, version, "env")
	glog.Infoln("LegacySetLive", domain, service, version, "Path=", realpath)
	if !zk.PathExists(this.conn, realpath) {
		return ErrNotFound
	}
	live := zk.GetString(this.conn, livepath)
	if live == nil {
		*live = realpath.Path()
	} else {
		p1 := strings.Split(*live, ",")[0]
		*live = strings.Join([]string{p1, realpath.Path()}, ",")
	}
	err := zk.CreateOrSetString(this.conn, livepath, *live)
	if err != nil {
		return err
	}
	err = zk.Increment(this.conn, registry.NewPath(domain, service, "live", "watch"), 1)
	if err != nil {
		return err
	}
	return nil
}

func (this *Service) SetLive(c Context, domainClass, domainInstance, service, version string) error {
	realpath := GetEnvPath(domainClass, domainInstance, service, version)
	glog.Infoln("SetLive", "Path=", realpath)
	if !zk.PathExists(this.conn, realpath) {
		return ErrNotFound
	}

	err := zk.CreateOrSetString(this.conn, GetEnvLivePath(domainClass, domainInstance, service), realpath.Path())
	if err != nil {
		return err
	}

	// Update the watch nodes
	err = zk.Increment(this.conn, GetEnvWatchPath(domainClass, domainInstance, service), 1)
	if err != nil {
		return err
	}

	// Legacy
	return this.legacy_setlive(domainClass, domainInstance, service, version)
}

func (this *Service) ListEnvVersions(c Context, domainClass, domainInstance, service string) (EnvVersions, error) {
	glog.Infoln("ListEnvVersions", "DomainClass=", domainClass, "DomainInstance=", domainInstance, "Service=", service)

	result := make(EnvVersions)
	err := VisitEnvVersions(this.conn, domainClass, domainInstance, service,
		func(version string, parent *zk.Node) bool {
			result[version] = false
			return true
		})
	// read the live version
	realpath := zk.GetString(this.conn, GetEnvLivePath(domainClass, domainInstance, service))
	if realpath != nil {
		result[registry.NewPath(*realpath).Dir().Base()] = true
	}

	return result, err
}

func (this *Service) GetEnvLiveVersion(c Context, domainClass, domainInstance, service string) (EnvList, error) {
	domain := ToDomainName(domainClass, domainInstance)
	glog.Infoln("GetEnvLiveVersion", domain, service)

	// read the live version
	realpath := zk.GetString(this.conn, GetEnvLivePath(domainClass, domainInstance, service))
	if realpath == nil {
		return nil, ErrNotFound
	}

	version := registry.NewPath(*realpath).Dir().Base()
	v, _, err := this.GetEnv(c, domainClass, domainInstance, service, version)
	return v, err
}
