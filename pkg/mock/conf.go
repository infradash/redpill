package mock

import (
	"github.com/infradash/redpill/pkg/conf"
	"os"
)

func init() {
	current_dir, err := os.Getwd()
	if err != nil {
		panic(err)
	}
	conf_db, err = init_conf_db(current_dir, "mock-conf.db")
	if err != nil {
		panic(err)
	}
}

type conf_base int

func ConfStorage() conf.ConfStorage {
	return conf_base(1)
}

func (this conf_base) ListAll(domainClass, service string) ([]string, []int, error) {
	return list_conf_domain_service(conf_db, domainClass, service)
}

func (this conf_base) Save(domainClass, service, name string, content []byte) error {
	return save_conf_domain_service_name(conf_db, domainClass, service, name, content)
}

func (this conf_base) Get(domainClass, service, name string) ([]byte, error) {
	return get_conf_domain_service(conf_db, domainClass, service, name)
}

func (this conf_base) Delete(domainClass, service, name string) error {
	return delete_conf_domain_service(conf_db, domainClass, service, name)
}

func (this conf_base) SaveVersion(domainClass, domainInstance, service, version, name string, content []byte) error {
	return save_conf_domain_instance_service_name_version(conf_db,
		domainClass, domainInstance, service, name, version, content)
}

func (this conf_base) GetVersion(domainClass, domainInstance, service, version, name string) ([]byte, error) {
	return get_conf_domain_service_version(conf_db, domainClass, domainInstance, service, name, version)
}

func (this conf_base) DeleteVersion(domainClass, domainInstance, service, version, name string) error {
	return delete_conf_domain_instance_service_name_version(conf_db,
		domainClass, domainInstance, service, name, version)
}
