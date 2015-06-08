package redpill

import (
	"fmt"
	"github.com/golang/glog"
	"github.com/qorio/maestro/pkg/zk"
	"strings"
	"time"
)

const (
	EnvZookeeper = "ZOOKEEPER_HOSTS"
)

type Zk struct {
	Hosts   string        `json:"zk_hosts"`
	Timeout time.Duration `json:"zk_timeout"`
	conn    zk.ZK
}

func (this *Zk) Connect() error {
	if this.conn != nil {
		return nil
	}

	glog.Infoln("Connecting to zookeeper:", this.Hosts)
	zk, err := zk.Connect(strings.Split(this.Hosts, ","), this.Timeout)
	if err != nil {
		return err
	}
	glog.Infoln("Connected to zookeeper:", this.Hosts)
	this.conn = zk
	return nil
}

// EnvService
func (this *Zk) GetEnv(domain, service, version string) (EnvList, Revision, error) {
	key := fmt.Sprintf("/%s/%s/%s/env", domain, service, version)
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
	return list, Revision(zn.Stats.Cversion), nil
}

// EnvService
func (this *Zk) UpdateEnv(domain, service, version string, change EnvChange, rev Revision) error {
	root := fmt.Sprintf("/%s/%s/%s/env", domain, service, version)
	if zn, err := this.conn.Get(root); err != nil && err != zk.ErrNotExist {
		return err
	} else if Revision(zn.Stats.Cversion) != rev {
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
		n, err := this.conn.Get(fmt.Sprintf("/%s/%s", root, delete.Name))
		switch {
		case err == zk.ErrNotExist:
		case err != nil:
			return err
		default:
			deletes = append(deletes, n)
		}
	}

	// everything ok. commit changes.  Note this is not atomic!
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
