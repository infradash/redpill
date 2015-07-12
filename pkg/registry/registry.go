package registry

import (
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

func NewService(pool func() zk.ZK) RegistryService {
	s := new(Service)
	s.conn = pool()
	return s
}

func (this *Service) load_and_check(key string, rev Revision) (*zk.Node, error) {
	zn, err := this.conn.Get(key)
	switch {
	case err == zk.ErrNotExist:
		return nil, ErrNotFound
	case err != nil:
		return nil, err
	}
	if zn.Stats.Version != int32(rev) {
		return nil, ErrConflict
	}
	return zn, nil
}

// RegistryService
func (this *Service) GetEntry(c Context, key string) ([]byte, Revision, error) {
	glog.Infoln("GetEntry:", c.UserId(), "Key=", key)
	zn, err := this.conn.Get(key)
	switch {
	case err == zk.ErrNotExist:
		return nil, Revision(0), ErrNotFound
	case err != nil:
		return nil, Revision(0), err
	}
	return zn.Value, Revision(zn.Stats.Version), nil
}

// RegistryService
func (this *Service) UpdateEntry(c Context, key string, value []byte, rev Revision) (Revision, error) {
	glog.Infoln("UpdateEntry:", c.UserId(), "Key=", key, "Rev=", rev)
	zn, err := this.load_and_check(key, rev)
	switch {

	case err == nil:
		err = zn.Set(value)
		if err != nil {
			return Revision(0), err
		}
		return Revision(zn.Stats.Version), nil
	case err == ErrNotFound:
		zn, err = this.conn.Create(key, value)
		if err != nil {
			return Revision(0), err
		}
		return Revision(zn.Stats.Version), nil
	default:
		return Revision(0), err
	}
}

// RegistryService
func (this *Service) DeleteEntry(c Context, key string, rev Revision) error {
	glog.Infoln("DeleteEntry:", c.UserId(), "Key=", key, "Rev=", rev)
	zn, err := this.load_and_check(key, rev)
	if err != nil {
		return err
	}
	return zn.Delete()
}
