package storage

import (
	"fmt"
	"strings"

	"github.com/boltdb/bolt"
	"github.com/pkg/errors"
	"github.com/vmihailenco/msgpack"
)

type bkey []byte

// DBTxn an interface that describes a database storage root allow you to create nested roots
type DBTxn interface {
	Bucket(key []byte) *bolt.Bucket
	CreateBucketIfNotExists(key []byte) (*bolt.Bucket, error)
}

// DBRoot an interface that describes a database storage root allowing you to store and retrieve data
type DBRoot interface {
	DBTxn
	Get(key []byte) []byte
	Put(key []byte, data []byte) error
}

// TxnRunner a function that runs inside of a transaction
// If an error is returned the transaction is rolled back
type TxnRunner func(tx *bolt.Tx) error

// Database a storage interface
type Database struct {
	store *bolt.DB
}

// Collection creates a collection from the root at the given path
// Use a `/` to create a nested collection
func (d *Database) Collection(path string, txn DBTxn) (*Collection, error) {
	return NewCollection(strings.Split(path, "/"), txn, nil)
}

// Transaction starts a transaction
func (d *Database) Transaction(run TxnRunner) error {
	return d.store.Update(run)
}

// Close safely disconnects the database
func (d *Database) Close() error {
	return d.store.Close()
}

// NewCollection recursively creates a collection starting at the given root
func NewCollection(paths []string, txn DBTxn, parent *Collection) (*Collection, error) {
	if len(paths) == 0 {
		return parent, nil
	}
	bucket, err := txn.CreateBucketIfNotExists(bkey(paths[0]))
	if err != nil {
		return parent, err
	}
	return NewCollection(paths[1:], bucket, &Collection{
		ID:     paths[0],
		parent: parent,
		txn:    txn,
		root:   bucket,
	})
}

// Collection represents a location in the database that holds keys and values
// This can be nested inside other Collections
type Collection struct {
	ID     string
	parent *Collection
	txn    DBTxn
	root   DBRoot
}

// String
func (c *Collection) String() string {
	return fmt.Sprintf("Collection: %s", strings.Join(c.Path(), " -> "))
}

// Path returns the path to the collection by traversing it's parents IDs
func (c *Collection) Path(path ...string) []string {
	path = append([]string{c.ID}, path...)
	if c.parent == nil {
		return path
	}
	return c.parent.Path(path...)
}

// Collection creates a nested collection
func (c *Collection) Collection(id string) (*Collection, error) {
	return NewCollection([]string{id}, c.txn, c)
}

// Put sets a value in the collection by its key
func (c *Collection) Put(key string, v interface{}) error {
	data, err := msgpack.Marshal(v)
	if err != nil {
		return err
	}
	return c.root.Put(bkey(key), data)
}

// Get returns a value from the collection by its key
func (c *Collection) Get(key string, v interface{}) (bool, error) {
	data := c.root.Get(bkey(key))
	if data == nil {
		return false, nil
	}
	return true, msgpack.Unmarshal(data, v)
}

// Open creates a connection to the database
func Open(path string) (*Database, error) {
	db, err := bolt.Open(path, 0666, nil)
	if err != nil {
		return nil, err
	}
	return &Database{
		store: db,
	}, nil
}
