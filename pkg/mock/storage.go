package mock

import (
	"github.com/boltdb/bolt"
	"github.com/golang/glog"
	"path/filepath"
)

var (
	boltdb *bolt.DB
)

type db_builder struct {
	db  *bolt.DB
	err error
}

func init_db(dir, file string) *db_builder {
	db, err := bolt.Open(filepath.Join(dir, file), 0600, nil)
	if err != nil {
		return &db_builder{db: nil, err: err}
	}

	err = db.Update(func(tx *bolt.Tx) error {
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
	return &db_builder{db: db, err: err}
}

func (builder *db_builder) done() (*bolt.DB, error) {
	return builder.db, builder.err
}

func (builder *db_builder) buckets(bucket string, more ...string) (*bolt.DB, error) {
	if builder.err != nil {
		return nil, builder.err
	}

	builder.err = builder.db.Update(func(tx *bolt.Tx) error {
		for _, b := range append(more, bucket) {
			_, err := tx.CreateBucketIfNotExists([]byte(b))
			if err != nil {
				return err
			}
		}
		return nil
	})
	return builder.db, builder.err
}

func write_boltdb(tx *bolt.Tx, key string, value []byte, bucket string, subbucket ...string) error {
	b, err := tx.CreateBucketIfNotExists([]byte(bucket))
	if err != nil {
		return err
	}
	if len(subbucket) > 0 {
		for _, sb := range subbucket {
			b, err = b.CreateBucketIfNotExists([]byte(sb))
			if err != nil {
				return err
			}
		}
	}
	if err := b.Put([]byte(key), value); err != nil {
		glog.Warningln(err)
		return err
	}
	return nil
}

func read_boltdb(tx *bolt.Tx, key string, bucket string, subbucket ...string) ([]byte, error) {
	b := tx.Bucket([]byte(bucket))
	if b == nil {
		return nil, nil
	}
	if len(subbucket) > 0 {
		for _, sb := range subbucket {
			b = b.Bucket([]byte(sb))
			if b == nil {
				return nil, nil
			}
		}
	}
	if v := b.Get([]byte(key)); len(v) == 0 {
		return nil, nil
	} else {
		return v, nil
	}
}

func delete_boltdb(tx *bolt.Tx, key string, bucket string, subbucket ...string) error {
	b := tx.Bucket([]byte(bucket))
	if b == nil {
		return nil
	}
	if len(subbucket) > 0 {
		for _, sb := range subbucket {
			b = b.Bucket([]byte(sb))
			if b == nil {
				return nil
			}
		}
	}
	return b.Delete([]byte(key))
}

func list_keys_sizes_boltdb(tx *bolt.Tx, bucket string, subbucket ...string) ([][]byte, []int, error) {
	b := tx.Bucket([]byte(bucket))
	if b == nil {
		return nil, nil, nil
	}
	if len(subbucket) > 0 {
		for _, sb := range subbucket {
			b = b.Bucket([]byte(sb))
			if b == nil {
				return nil, nil, nil
			}
		}
	}

	keys := [][]byte{}
	sizes := []int{}
	cur := b.Cursor()
	k, v := cur.First()
	for {
		if k == nil && v == nil {
			break
		}
		keys = append(keys, k)
		sizes = append(sizes, len(v))

		k, v = cur.Next()
	}
	return keys, sizes, nil
}
