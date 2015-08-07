package mock

import (
	"github.com/boltdb/bolt"
	"github.com/golang/glog"
)

var (
	boltdb *bolt.DB
)

func write_bucket(tx *bolt.Tx, bucket, key string, value []byte, subbucket ...string) error {
	var err error
	b := tx.Bucket([]byte(bucket))
	if b == nil {
		return nil
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

func read_nested(tx *bolt.Tx, bucket, key string, subbucket ...string) ([]byte, error) {
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
	return b.Get([]byte(key)), nil
}

func delete_nested(tx *bolt.Tx, bucket, key string, subbucket ...string) error {
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

func write_nested(tx *bolt.Tx, value []byte, subs ...string) error {
	var err error
	var b *bolt.Bucket
	for _, bn := range subs[0 : len(subs)-1] {
		b, err = b.CreateBucketIfNotExists([]byte(bn))
		if err != nil {
			return err
		}
	}
	if err := b.Put([]byte(subs[len(subs)-1]), value); err != nil {
		glog.Warningln(err)
		return err
	}
	return nil
}
