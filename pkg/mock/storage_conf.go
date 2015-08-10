package mock

import (
	"github.com/boltdb/bolt"
	"github.com/golang/glog"
)

const (
	dbBucketConfDomainClassServiceName = "conf_domainClass_service_name"
)

var (
	conf_db *bolt.DB
)

func init_conf_db(dir, file string) (*bolt.DB, error) {
	glog.Infoln("Db dir=", dir, "file=", file)
	return init_db(dir, file).buckets(dbBucketConfDomainClassServiceName)
}

func save_conf_domain_service_name(db *bolt.DB, domainClass, service, name string, conf []byte) error {
	return db.Update(func(tx *bolt.Tx) error {
		return write_boltdb(tx, name, conf, domainClass, service)
	})
}

func get_conf_domain_service(db *bolt.DB, domainClass, service, name string) ([]byte, error) {
	var err error
	var result []byte
	db.View(func(tx *bolt.Tx) error {
		result, err = read_boltdb(tx, name, domainClass, service)
		return err
	})
	return result, err
}

func list_conf_domain_service(db *bolt.DB, domainClass, service string) ([]string, []int, error) {
	keys := []string{}
	sizes := []int{}
	err := db.View(func(tx *bolt.Tx) error {
		k, s, e := list_keys_sizes_boltdb(tx, domainClass, service)
		if e != nil {
			return e
		}
		for i, kk := range k {
			keys = append(keys, string(kk))
			sizes = append(sizes, s[i])
		}
		return nil
	})
	return keys, sizes, err
}

func delete_conf_domain_service(db *bolt.DB, domainClass, service, name string) error {
	return db.Update(func(tx *bolt.Tx) error {
		return delete_boltdb(tx, name, domainClass, service)
	})
}

func save_conf_domain_instance_service_name_version(db *bolt.DB,
	domainClass, domainInstance, service, name, version string, conf []byte) error {
	return db.Update(func(tx *bolt.Tx) error {
		return write_boltdb(tx, name, conf, domainClass, service, domainInstance, version)
	})
}

func get_conf_domain_service_version(db *bolt.DB,
	domainClass, domainInstance, service, name, version string) ([]byte, error) {

	var err error
	var result []byte

	err = db.View(func(tx *bolt.Tx) error {
		result, err = read_boltdb(tx, name, domainClass, service, domainInstance, version)
		return err
	})
	if err != nil {
		return nil, err
	}

	if len(result) == 0 {
		result, err = get_conf_domain_service(db, domainClass, service, name)
	}
	return result, err
}

func delete_conf_domain_instance_service_name_version(db *bolt.DB,
	domainClass, domainInstance, service, name, version string) error {
	return db.Update(func(tx *bolt.Tx) error {
		return delete_boltdb(tx, name, domainClass, service, domainInstance, version)
	})
}
