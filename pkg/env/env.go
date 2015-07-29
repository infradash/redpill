package env

import (
	"fmt"
	"github.com/golang/glog"
	. "github.com/infradash/redpill/pkg/api"
	"github.com/qorio/maestro/pkg/zk"
)

const (
	EnvZkHosts = "REDPILL_ZK_HOSTS"
)

type Service struct {
	conn zk.ZK
}

func NewService(pool func() zk.ZK) EnvService {
	s := new(Service)
	s.conn = pool()
	return s
}

/// TODO -- this is actually not very accurate to use only the Cversion, which is a version
/// number associated with the number of children of a znode.  A true version should be
/// calculated based on the content of the children or some hash of all the children's versions
func (this *Service) calculate_rev_from_parent(zn *zk.Node) Revision {
	return Revision(zn.Stats.Cversion)
}

// EnvService
func (this *Service) GetEnv(c Context, domain, service, version string) (EnvList, Revision, error) {
	key := fmt.Sprintf("/%s/%s/%s/env", domain, service, version)
	glog.Infoln("GetEnv:", c.UserId(), "Domain=", domain, "Service=", service, "Version=", version, "Key=", key)
	zn, err := this.conn.Get(key)
	if err != nil {
		return nil, -1, err
	}

	list := EnvList{}
	_, err = zn.VisitChildrenRecursive(func(n *zk.Node) bool {
		if n.IsLeaf() {
			list = append(list, Env{
				Name:  n.GetBasename(),
				Value: n.GetValueString(),
			})
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
	for _, create := range *vars {
		k := fmt.Sprintf("%s/%s", root, create.Name)
		_, err := this.conn.Create(k, []byte(create.Value))
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

	creates := []Env{}
	updates := []struct {
		Node  *zk.Node
		Value []byte
	}{}
	for _, update := range change.Update {
		n, err := this.conn.Get(fmt.Sprintf("%s/%s", root, update.Name))

		switch {
		case err == zk.ErrNotExist:
			creates = append(creates, update)
		case err != nil:
			return err
		default:
			updates = append(updates, struct {
				Node  *zk.Node
				Value []byte
			}{n, []byte(update.Value)})
		}
	}

	deletes := []*zk.Node{}
	for _, delete := range change.Delete {
		k := fmt.Sprintf("%s/%s", root, delete.Name)
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
	for _, create := range creates {
		k := fmt.Sprintf("%s/%s", root, create.Name)
		_, err := this.conn.Create(k, []byte(create.Value))
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
