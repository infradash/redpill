package mock

import (
	"encoding/json"
	"fmt"
	"github.com/boltdb/bolt"
	"github.com/golang/glog"
	. "github.com/infradash/redpill/pkg/orchestrate"
	"path/filepath"
)

const (
	dbBucketOrchestrateModels            = "orchestrate_models"
	dbBucketOrchestrateInstancesById     = "orchestrate_instances_id"
	dbBucketOrchestrateInstancesByDomain = "orchestrate_instances_domain"
)

func init_orchestrate_db(dir, file string) (*bolt.DB, error) {
	glog.Infoln("Db dir=", dir, "file=", file)
	db, err := bolt.Open(filepath.Join(dir, file), 0600, nil)
	if err != nil {
		return nil, err
	}

	db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte(dbBucketOrchestrateModels))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte(dbBucketOrchestrateInstancesByDomain))
		if err != nil {
			return err
		}
		_, err = tx.CreateBucketIfNotExists([]byte(dbBucketOrchestrateInstancesById))
		return err
	})
	return db, nil
}

func save_orchestrate_model(boltdb *bolt.DB, domain string, model *Model) error {
	return boltdb.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(dbBucketOrchestrateModels))
		if b == nil {
			return nil
		}
		bb, err := b.CreateBucketIfNotExists([]byte(domain))
		if err != nil {
			return err
		}
		key := model.Name
		buff, err := json.Marshal(model)
		if err != nil {
			return err
		}
		return bb.Put([]byte(key), buff)
	})
}

func load_models_for_domain(boltdb *bolt.DB, domain string) ([]Model, error) {
	result := []Model{}
	err := boltdb.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(dbBucketOrchestrateModels))
		if b == nil {
			return nil
		}
		bb := b.Bucket([]byte(domain))
		if bb == nil {
			return nil
		}
		cur := bb.Cursor()
		k, v := cur.First()
		for {
			if k == nil && v == nil {
				break
			}
			m := &Model{}
			err := json.Unmarshal(v, m)
			if err != nil {
				return err
			}
			result = append(result, *m)
			k, v = cur.Next()
		}
		return nil
	})
	return result, err
}

func find_model_for_domain_name(boltdb *bolt.DB, domain, name string) (*Model, error) {
	var result *Model
	err := boltdb.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(dbBucketOrchestrateModels))
		if b == nil {
			return nil
		}
		bb := b.Bucket([]byte(domain))
		if bb == nil {
			return nil
		}
		buff := bb.Get([]byte(name))
		if buff == nil {
			return nil
		}
		m := Model{}
		err := json.Unmarshal(buff, &m)
		if err != nil {
			return err
		}
		result = &m
		return nil
	})
	return result, err
}

func delete_model_for_domain_name(boltdb *bolt.DB, domain, name string) error {
	err := boltdb.Update(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(dbBucketOrchestrateModels))
		if b == nil {
			return nil
		}
		bb := b.Bucket([]byte(domain))
		if bb == nil {
			return nil
		}
		return bb.Delete([]byte(name))
	})
	return err
}

func save_orchestrate_instance(boltdb *bolt.DB, instance *Instance) error {
	buff, err := json.Marshal(instance)
	if err != nil {
		glog.Warningln(err)
		return err
	}
	return boltdb.Update(func(tx *bolt.Tx) error {
		if err := write_bucket(tx, dbBucketOrchestrateInstancesById, instance.Info().Id, buff); err != nil {
			glog.Warningln(err)
			return err
		}
		key := fmt.Sprintf("%d.%s", instance.Info().StartTime.Unix(), instance.Info().Id)
		if err := write_bucket(tx, dbBucketOrchestrateInstancesByDomain, key, buff, instance.Info().Domain); err != nil {
			glog.Warningln(err)
			return err
		}
		return err
	})
}

func find_instance_by_id(boltdb *bolt.DB, id string) (*Instance, error) {
	var result *Instance
	err := boltdb.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(dbBucketOrchestrateInstancesById))
		if b == nil {
			return nil
		}
		buff := b.Get([]byte(id))
		if buff != nil {
			result = &Instance{}
			return json.Unmarshal(buff, result)
		}
		return nil
	})
	return result, err
}

func load_instances_for_domain_orchestration(boltdb *bolt.DB, domain, orchestration string) ([]Instance, error) {
	result := []Instance{}
	err := boltdb.View(func(tx *bolt.Tx) error {
		b := tx.Bucket([]byte(dbBucketOrchestrateInstancesByDomain))
		if b == nil {
			return nil
		}
		bb := b.Bucket([]byte(domain))
		if bb == nil {
			return nil
		}
		cur := bb.Cursor()
		k, v := cur.First()
		for {
			if k == nil && v == nil {
				break
			}
			m := &Instance{}
			err := json.Unmarshal(v, m)
			if err != nil {
				return err
			}
			if string(m.Model().GetName()) == orchestration {
				result = append(result, *m)
			}
			k, v = cur.Next()
		}
		return nil
	})
	return result, err
}
