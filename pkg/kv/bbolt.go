package kv

import (
	"context"
	"path"

	"github.com/hashicorp/go-hclog"
	bolt "go.etcd.io/bbolt"
)

// Bolt implements the http.KV interface for storage to an on-disk
// bolt-db.
type Bolt struct {
	db *bolt.DB
	l  hclog.Logger
}

// NewBolt initializes and connects the boltdb.
func NewBolt(l hclog.Logger) (*Bolt, error) {
	db, err := bolt.Open("pitman.db", 0600, nil)
	if err != nil {
		return nil, err
	}
	err = db.Update(func(tx *bolt.Tx) error {
		_, err := tx.CreateBucketIfNotExists([]byte("pitman"))
		return err
	})
	if err != nil {
		return nil, err
	}
	return &Bolt{db: db, l: l.Named("bbolt")}, nil
}

// Keys performs a table scan prefixed to the provided value.
func (b *Bolt) Keys(_ context.Context, prefix string) ([]string, error) {
	b.l.Debug("Performing key prefix scan", "prefix", prefix)
	keys := []string{}
	err := b.db.View(func(tx *bolt.Tx) error {
		bk := tx.Bucket([]byte("pitman"))
		bk.ForEach(func(k, _ []byte) error {
			if matched, _ := path.Match(prefix, string(k)); matched {
				b.l.Trace("Matched Key", "key", string(k))
				keys = append(keys, string(k))
			}
			return nil
		})
		return nil
	})
	return keys, err
}

// Get returns a single value pointed to by the designated key.
func (b *Bolt) Get(_ context.Context, key string) ([]byte, error) {
	b.l.Debug("Reading value from key", "key", key)
	var val []byte
	err := b.db.View(func(tx *bolt.Tx) error {
		bk := tx.Bucket([]byte("pitman"))
		v := bk.Get([]byte(key))
		val = make([]byte, len(v))
		copy(val, v)
		return nil
	})
	return val, err
}

// Put stores a value into the specified key.  This is an exclusive
// transaction and blocks all other write transactions until it
// returns.
func (b *Bolt) Put(_ context.Context, key string, val []byte) error {
	b.l.Debug("Putting value", "key", key, "value", string(val))
	return b.db.Update(func(tx *bolt.Tx) error {
		bk := tx.Bucket([]byte("pitman"))
		return bk.Put([]byte(key), val)
	})
}

// Ping is used by other storage implementations to determine if the
// remote endpoint is connected.
func (b *Bolt) Ping(_ context.Context) error {
	return nil
}

// Close flushes the database and makes it invalid for future
// operations.
func (b *Bolt) Close() error {
	return b.db.Close()
}
