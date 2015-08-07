package mock

import (
	"github.com/boltdb/bolt"
	"github.com/golang/glog"
	"path/filepath"
)

const (
	dbBucketConfDomainClassServiceName = "conf_domainClass_service_name"
)

func init_conf_db(dir, file string) (*bolt.DB, error) {
	glog.Infoln("Db dir=", dir, "file=", file)
	db, err := bolt.Open(filepath.Join(dir, file), 0600, nil)
	if err != nil {
		return nil, err
	}

	db.Update(func(tx *bolt.Tx) error {
		for _, b := range []string{dbBucketConfDomainClassServiceName} {
			_, err := tx.CreateBucketIfNotExists([]byte(b))
			if err != nil {
				return err
			}
		}
		return nil
	})
	return db, nil
}

func save_conf_domain_service_name(boltdb *bolt.DB, domain, service, name string, conf []byte) error {
	return boltdb.Update(func(tx *bolt.Tx) error {
		return write_nested(tx, conf, domain, service, name)
	})
}
