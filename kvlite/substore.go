package kvlite

import (
	"fmt"
	"strings"
)

type substore struct {
	prefix string
	db     Store
}

const sepr = '\x1f'

// applies prefix of table to calls.
func (d substore) apply_prefix(name string) string {
	return string(append([]rune(d.prefix), sepr))
}

func (d *substore) Sub(name string) Store {
	return &substore{fmt.Sprintf("%s%s%c", d.prefix, name, sepr), d.db}
}

// Creates a bucket with a common namespace.
func (d *substore) Shared(name string) Store {
	return &substore{fmt.Sprintf("__shared__%c%s%c", sepr, name, sepr), d.db}
}

func (d substore) Close() (err error) {
	return d.db.Close()
}

// DB Wrappers to perform fatal error checks on each call.
func (d substore) Drop(table string) (err error) {
	return d.db.Drop(d.apply_prefix(table))
}

// Encrypt value to go-kvlie, fatal on error.
func (d substore) CryptSet(table, key string, value interface{}) error {
	return d.db.CryptSet(d.apply_prefix(table), key, value)
}

// Save value to go-kvlite.
func (d substore) Set(table, key string, value interface{}) error {
	return d.db.Set(d.apply_prefix(table), key, value)
}

// Retrieve value from go-kvlite.
func (d substore) Get(table, key string, output interface{}) (bool, error) {
	return d.db.Get(d.apply_prefix(table), key, output)
}

// List keys in go-kvlite.
func (d substore) Keys(table string) ([]string, error) {
	return d.db.Keys(d.apply_prefix(table))
}

// Count keys in table.
func (d substore) CountKeys(table string) (int, error) {
	return d.db.CountKeys(d.apply_prefix(table))
}

func (d substore) Buckets(limit_depth bool) (buckets []string, err error) {
	bmap := make(map[string]struct{})

	tmp, e := d.db.Buckets(false)
	if e != nil {
		return buckets, e
	}
	for _, t := range tmp {
		if strings.HasPrefix(t, d.prefix) {
			if name := strings.TrimPrefix(t, d.prefix); strings.ContainsRune(name, sepr) {
				if !limit_depth {
					buckets = append(buckets, name)
				} else {
					name = strings.Split(name, string(sepr))[0]
					if _, ok := bmap[name]; !ok {
						bmap[name] = struct{}{}
						buckets = append(buckets, name)
					}
				}
			}
		}
	}
	return buckets, err
}

// List Tables in DB
func (d substore) Tables() (buckets []string, err error) {
	tmp, e := d.db.Buckets(true)
	if e != nil {
		return buckets, e
	}
	for _, t := range tmp {
		if strings.HasPrefix(t, d.prefix) {
			if name := strings.TrimPrefix(t, d.prefix); !strings.ContainsRune(name, sepr) {
				buckets = append(buckets, name)
			}
		}
	}
	return buckets, err
}

// Delete value from go-kvlite.
func (d substore) Unset(table, key string) error {
	return d.db.Unset(d.apply_prefix(table), key)
}

// Drill in to specific table.
func (d substore) Table(table string) Table {
	return d.db.Table(d.apply_prefix(table))
}
