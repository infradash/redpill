package mock

import (
	"github.com/boltdb/bolt"
	. "gopkg.in/check.v1"
	"os"
	"testing"
)

func TestStorage(t *testing.T) { TestingT(t) }

type StorageTests struct {
}

var _ = Suite(&StorageTests{})

const (
	testdb = "test.db"
)

func (suite *StorageTests) SetUpSuite(c *C) {
	db, err := init_orchestrate_db(os.TempDir(), testdb)
	c.Assert(err, Equals, nil)
	c.Log(db)
	defer db.Close()

	// delete the buckets
	err = db.Update(func(tx *bolt.Tx) error {
		for _, b := range []string{dbBucketOrchestrateModels,
			dbBucketOrchestrateInstancesById,
			dbBucketOrchestrateInstancesByDomain} {

			err := tx.DeleteBucket([]byte(b))
			if err != nil {
				panic(err)
			}
		}
		return nil
	})
	c.Assert(err, Equals, nil)
	return
}

func (suite *StorageTests) TestInitDb(c *C) {
	db, err := init_orchestrate_db(os.TempDir(), testdb)
	defer db.Close()
	c.Assert(err, Equals, nil)
	c.Log(db)
}

func (suite *StorageTests) TestOrchestrateModelStorage(c *C) {
	db, err := init_orchestrate_db(os.TempDir(), testdb)
	defer db.Close()
	c.Assert(err, Equals, nil)
	c.Log(db)

	for _, m := range mock_models {
		err = save_orchestrate_model(db, "test", &m)
		c.Assert(err, Equals, nil)
	}

	m, err := find_model_for_domain_name(db, "test", string(mock_models[0].Name))
	c.Assert(err, Equals, nil)
	c.Assert(m, Not(Equals), nil)
	c.Log(m)
	c.Assert(string(m.Name), DeepEquals, string(mock_models[0].Name))

	list, err := load_models_for_domain(db, "test")
	c.Assert(err, Equals, nil)
	c.Assert(len(list), Equals, len(mock_models))
	c.Log(list)
}

func (suite *StorageTests) TestOrchestrateInstanceStorage(c *C) {
	db, err := init_orchestrate_db(os.TempDir(), testdb)
	defer db.Close()
	c.Assert(err, Equals, nil)
	c.Log(db)

	ids := []string{}
	for i := 0; i < 10; i++ {
		instance := mock_models[1].NewInstance("test")
		ids = append(ids, instance.Info().Id)
		err = save_orchestrate_instance(db, instance)
		c.Assert(err, Equals, nil)
	}
	c.Log("orchestration=", mock_models[1].Name, "ids=", ids, "len=", len(ids))

	instance, err := find_instance_by_id(db, ids[2])
	c.Assert(err, Equals, nil)
	c.Log(instance)
	c.Assert(instance, Not(Equals), nil)
	c.Assert(instance.Info().Id, Equals, ids[2])

	instances, err := load_instances_for_domain_orchestration(db, "test", string(mock_models[1].Name))
	c.Assert(err, Equals, nil)
	c.Assert(len(instances), Equals, 10)
	c.Log(instances)
}
