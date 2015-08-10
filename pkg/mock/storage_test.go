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
	testdb_base = "test_base.db"
)

func (suite *StorageTests) SetUpSuite(c *C) {
	db, err := init_db(os.TempDir(), testdb_base).buckets("test1", "test2", "test3")
	c.Assert(err, Equals, nil)
	c.Log(db)
	defer db.Close()
	return
}

func (suite *StorageTests) TestNested(c *C) {
	db, err := init_db(os.TempDir(), "nested").buckets("test1", "test2", "test3")
	c.Assert(err, Equals, nil)

	err = db.Update(func(tx *bolt.Tx) error {
		write_boltdb(tx, "test111", []byte("test111"), "test1", "test1-1", "test1-1-1")
		write_boltdb(tx, "test121", []byte("test121"), "test1", "test1-2", "test1-2-1")
		write_boltdb(tx, "test11", []byte("test11"), "test1", "test1-1")
		return nil
	})
	c.Assert(err, Equals, nil)

	db.View(func(tx *bolt.Tx) error {
		v, err := read_boltdb(tx, "test111", "test1", "test1-1", "test1-1-1")
		c.Assert(err, Equals, nil)
		c.Assert(v, Not(Equals), nil)
		c.Assert(string(v), Equals, "test111")

		v, err = read_boltdb(tx, "test121", "test1", "test1-2", "test1-2-1")
		c.Assert(err, Equals, nil)
		c.Assert(v, Not(Equals), nil)
		c.Assert(string(v), Equals, "test121")

		v, err = read_boltdb(tx, "test11", "test1", "test1-1")
		c.Assert(err, Equals, nil)
		c.Assert(v, Not(Equals), nil)
		c.Assert(string(v), Equals, "test11")

		v, err = read_boltdb(tx, "test1", "test1-2")
		c.Assert(v, Not(Equals), nil)
		c.Assert(len(v), Equals, 0)
		c.Assert(err, Equals, nil)
		return nil
	})
}

func (suite *StorageTests) TestListAll(c *C) {
	db, err := init_db(os.TempDir(), "listing").buckets("b1")
	c.Assert(err, Equals, nil)

	err = db.Update(func(tx *bolt.Tx) error {
		write_boltdb(tx, "k1", []byte("v1"), "b1")
		write_boltdb(tx, "k2", []byte("v2"), "b1")
		write_boltdb(tx, "k3", []byte("v3"), "b1")
		write_boltdb(tx, "k4", []byte("v4"), "b1")
		write_boltdb(tx, "k5", []byte("v5"), "b1")

		return nil
	})
	c.Assert(err, Equals, nil)

	db.View(func(tx *bolt.Tx) error {
		k, s, e := list_keys_sizes_boltdb(tx, "b1")

		c.Assert(e, Equals, nil)
		c.Assert(len(k), Equals, 5)
		c.Assert(len(s), Equals, 5)
		c.Assert(k, DeepEquals, [][]byte{
			[]byte("k1"),
			[]byte("k2"),
			[]byte("k3"),
			[]byte("k4"),
			[]byte("k5"),
		})
		return nil
	})
}
