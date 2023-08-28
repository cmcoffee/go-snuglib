package kvlite

import (
	"bytes"
	"crypto/aes"
	"crypto/cipher"
	"crypto/sha256"
	"encoding/gob"
	"errors"
	"fmt"
	"github.com/boltdb/bolt"
	"strings"
	"time"
)

var ErrLocked = errors.New("Database is currently in use by an exisiting instance, please close it and try again.")

// Main Store Interface
type Store interface {
	// Tables provides a list of all tables.
	Tables() (tables []string, err error)
	// Table creats a key/val direct to a specified Table.
	Table(table string) Table
	// SubStore Creates a new bucket with a different namespace, tied to
	Sub(name string) Store
	// SyncStore Creates a new bucket for shared tenants.
	Bucket(name string) Store
	// Drop drops the specified table.
	Drop(table string) (err error)
	// CountKeys provides a total of keys in table.
	CountKeys(table string) (count int, err error)
	// Keys provides a listing of all keys in table.
	Keys(table string) (keys []string, err error)
	// CryptSet encrypts the value within the key/value pair in table.
	CryptSet(table, key string, value interface{}) (err error)
	// Set sets the key/value pair in table.
	Set(table, key string, value interface{}) (err error)
	// Unset deletes the key/value pair in table.
	Unset(table, key string) (err error)
	// Get retrieves value at key in table.
	Get(table, key string, output interface{}) (found bool, err error)
	// Close closes the kvliter.Store.
	Close() (err error)
	// Buckets lists all bucket namespaces, limit_depth limits to first-level buckets
	buckets(limit_depth bool) (stores []string, err error)
}

// Table Interface follows the Main Store Interface, but directly to a table.
type Table interface {
	Keys() (keys []string, err error)
	CountKeys() (count int, err error)
	Set(key string, value interface{}) (err error)
	CryptSet(key string, value interface{}) (err error)
	Get(key string, value interface{}) (found bool, err error)
	Unset(key string) (err error)
	Drop() (err error)
}

type focused struct {
	table string
	store Store
}

func (s focused) Get(key string, value interface{}) (found bool, err error) {
	return s.store.Get(s.table, key, value)
}

func (s focused) Keys() (keys []string, err error) {
	return s.store.Keys(s.table)
}

func (s focused) CountKeys() (count int, err error) {
	return s.store.CountKeys(s.table)
}

func (s focused) Set(key string, value interface{}) (err error) {
	return s.store.Set(s.table, key, value)
}

func (s focused) CryptSet(key string, value interface{}) (err error) {
	return s.store.CryptSet(s.table, key, value)
}

func (s focused) Unset(key string) (err error) {
	return s.store.Unset(s.table, key)
}

func (s focused) Drop() (err error) {
	return s.store.Drop(s.table)
}

// Bolt Backend
type boltDB struct {
	db      *bolt.DB
	encoder encoder
}

type encoder []byte

// Get all buckets on system.
func (K *boltDB) buckets(limit_depth bool) (buckets []string, err error) {
	bmap := make(map[string]struct{})

	err = K.db.View(func(tx *bolt.Tx) error {
		add_bucket := func(name []byte, b *bolt.Bucket) error {
			name_str := string(name)
			if name_str == "KVLite" {
				return nil
			}
			if !limit_depth {
				buckets = append(buckets, name_str)
			} else {
				name_str = strings.Split(name_str, string(sepr))[0]
				if _, ok := bmap[name_str]; !ok {
					bmap[name_str] = struct{}{}
					buckets = append(buckets, name_str)
				}
			}
			return nil
		}
		return tx.ForEach(add_bucket)
	})
	return buckets, err
}

// Perform sha256.Sum256 against input byte string.
func hashBytes(input []byte) []byte {
	sum := sha256.Sum256(input)
	var output []byte
	output = append(output[0:], sum[0:]...)
	return output
}

// Encrypts bytes.
func (e encoder) encrypt(input []byte) []byte {

	key := hashBytes([]byte(e))
	block, _ := aes.NewCipher([]byte(e))

	buff := make([]byte, len(input))
	copy(buff, input)

	cipher.NewCFBEncrypter(block, key[0:block.BlockSize()]).XORKeyStream(buff, buff)

	return buff
}

// Decryps bytes.
func (e encoder) decrypt(input []byte) []byte {

	key := hashBytes([]byte(e))

	buff := make([]byte, len(input))
	copy(buff, input)

	block, _ := aes.NewCipher([]byte(e))
	cipher.NewCFBDecrypter(block, key[0:block.BlockSize()]).XORKeyStream(buff, buff)

	return buff
}

// Decodes input in to object.
func (e encoder) decode(input []byte, output interface{}) (err error) {
	var i []byte

	if input == nil {
		return nil
	}

	if input[0] == 1 {
		i = e.decrypt(input[1:])
	} else {
		i = input[1:]
	}

	x := gob.NewDecoder(bytes.NewBuffer(i))

	return x.Decode(output)
}

// Encodes input to bytes
func (e *encoder) encode(input interface{}) (output []byte, err error) {
	buff := bytes.NewBuffer(nil)
	x := gob.NewEncoder(buff)
	err = x.Encode(input)
	return buff.Bytes(), err
}

// Creates a bucket with a common namespace.
func (K *boltDB) Bucket(name string) Store {
	return K.Sub(name)
}

// Created a bucket using a different name space.
func (K *boltDB) Sub(name string) Store {
	return &substore{fmt.Sprintf("%s%c", name, sepr), K}
}

// Counts keys in table.
func (K *boltDB) CountKeys(table string) (count int, err error) {
	err = K.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(table))
		if bucket == nil {
			return nil
		}
		count = bucket.Stats().KeyN
		return nil
	})
	return
}

// Lists keys in table.
func (K *boltDB) Keys(table string) (keys []string, err error) {
	err = K.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(table))
		if bucket == nil {
			return nil
		}
		add_key := func(k, v []byte) error {
			keys = append(keys, string(k))
			return nil
		}
		return bucket.ForEach(add_key)
	})
	return keys, err
}

// Delete a key/value.
func (K *boltDB) Unset(table, key string) (err error) {
	return K.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(table))
		if bucket == nil {
			return nil
		}
		if err = bucket.Delete([]byte(key)); err != nil {
			return err
		}
		return nil
	})
}

// Drops table
func (K *boltDB) Drop(table string) (err error) {
	tmp, e := K.buckets(false)
	if e != nil {
		return e
	}

	var tables []string
	for _, v := range tmp {
		if strings.HasPrefix(v, fmt.Sprintf("%s%c", table, sepr)) || v == table {
			tables = append(tables, v)
		}
	}

	if len(tables) == 0 {
		return nil
	}

	for _, v := range tables {
		err = K.db.Update(func(tx *bolt.Tx) error {
			return tx.DeleteBucket([]byte(v))
		})
	}
	return
}

// Lists all tables
func (K *boltDB) Tables() (tables []string, err error) {
	tmp, e := K.buckets(true)
	if e != nil {
		return tables, e
	}
	for _, v := range tmp {
		if !strings.ContainsRune(v, sepr) {
			tables = append(tables, v)
		}
	}
	return tables, err
}

// Returns sub of table.
func (K *boltDB) Table(table string) Table {
	return focused{table: table, store: K}
}

// Retrieve value from bolt db.
func (K *boltDB) Get(table, key string, output interface{}) (found bool, err error) {
	return found, K.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(table))
		if bucket == nil {
			found = false
			return nil
		}
		data := bucket.Get([]byte(key))
		if data != nil {
			found = true
			if output == nil {
				return nil
			}
		}
		return K.encoder.decode(data, output)
	})
}

func (K *boltDB) Close() (err error) {
	return K.db.Close()
}

// Stores encrypted key/value pair.
func (K *boltDB) CryptSet(table, key string, value interface{}) (err error) {
	return K.set(table, key, value, true)
}

// Stores unencrypted key/value pair.
func (K *boltDB) Set(table, key string, value interface{}) (err error) {
	return K.set(table, key, value, false)
}

// Stores key/value pair in bolt.
func (K *boltDB) set(table, key string, value interface{}, encrypt_value bool) (err error) {
	return K.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(table))
		if err != nil {
			return err
		}

		v, err := K.encoder.encode(value)
		if err != nil {
			return err
		}

		if encrypt_value {
			v = K.encoder.encrypt(v)
			v = append([]byte{1}, v[0:]...)
		} else {
			v = append([]byte{0}, v[0:]...)
		}

		return bucket.Put([]byte(key), v)
	})
}

// Resets encryption key on database, removing all encrypted keys in the process.
func CryptReset(filename string) (err error) {
	db, err := open(filename)
	if err != nil {
		return err
	}

	db.Set("KVLite", "Reset", true)

	tables, err := db.buckets(false)
	if err != nil {
		return err
	}

	for _, t := range tables {
		var crypted_keys []string
		keys, err := db.Keys(t)
		if err != nil {
			return err
		}
		for _, k := range keys {
			err = db.db.View(func(tx *bolt.Tx) error {
				bucket := tx.Bucket([]byte(t))
				if bucket == nil {
					return nil
				}
				o := bucket.Get([]byte(k))
				if o != nil && o[0] == 1 {
					crypted_keys = append(crypted_keys, k)
				}
				return nil
			})
			if err != nil {
				return err
			}
		}
		for _, k := range crypted_keys {
			err = db.Unset(t, k)
			if err != nil {
				return err
			}
		}
	}
	err = db.Drop("KVLite")
	if err != nil {
		return err
	}
	return db.Close()
}

// Opens bolt keystore.
func open(filename string) (DB *boltDB, err error) {
	db, err := bolt.Open(filename, 0600, &bolt.Options{Timeout: 1 * time.Second})
	if err != nil {
		if err == bolt.ErrTimeout {
			err = ErrLocked
		}
		return nil, err
	}
	return &boltDB{db: db}, nil
}

// Opens BoltDB backed kvlite.Store.
func Open(filename string, padlock ...byte) (Store, error) {
	db, err := open(filename)
	if err != nil {
		return nil, err
	}

	found, err := db.Get("KVLite", "Reset", nil)
	if err != nil {
		return nil, err
	}

	if found {
		db.Close()
		err = CryptReset(filename)
		if err != nil {
			return nil, err
		}
		db, err = open(filename)
		if err != nil {
			return nil, err
		}
	}

	var X *xLock
	_, err = db.Get("KVLite", "X", &X)
	if err != nil {
		return nil, err
	}
	if X == nil {
		X = new(xLock)
	}

	db.encoder, err = X.dbunlocker(padlock)
	if err != nil {
		db.Close()
		return nil, err
	}
	err = db.Set("KVLite", "X", &X)
	return db, err
}
