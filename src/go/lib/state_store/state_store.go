package state_store

import (
	"encoding/json"
	"time"

	"github.com/mitchellh/mapstructure"
	"github.com/spf13/cast"
)

// StateStore a generic data container for buildable target state
type StateStore struct {
	Previous KVStore
	Current  KVStore
}

// NewStateStore initializes an empty state store
func NewStateStore() *StateStore {
	return &StateStore{
		Previous: KVStore{},
		Current:  KVStore{},
	}
}

// KVStore a key value store for arbitrary data. It provides methods to get and set values by type.
type KVStore map[string]interface{}

// Get retrieves a value by its key as an interface which can be cast to any other value
func (kvs KVStore) Get(key string) interface{} {
	return kvs[key]
}

// GetCheck retrieves a value by its key as an interface which can be cast to any other value
func (kvs KVStore) GetCheck(key string) (interface{}, bool) {
	v, exists := kvs[key]
	return v, exists
}

// Set adds an arbitrary value of any type by the specified key to the key value store
func (kvs KVStore) Set(key string, val interface{}) interface{} {
	kvs[key] = val
	return val
}

// GetString retrieves a value by its key as string
func (kvs KVStore) GetString(key string) string {
	return cast.ToString(kvs.Get(key))
}

// GetBool retrieves a value by its key as bool
func (kvs KVStore) GetBool(key string) bool {
	return cast.ToBool(kvs.Get(key))
}

// GetInt retrieves a value by its key as int
func (kvs KVStore) GetInt(key string) int {
	return cast.ToInt(kvs.Get(key))
}

// GetInt32 retrieves a value by its key as int 32
func (kvs KVStore) GetInt32(key string) int32 {
	return cast.ToInt32(kvs.Get(key))
}

// GetInt64 retrieves a value by its key as int 64
func (kvs KVStore) GetInt64(key string) int64 {
	return cast.ToInt64(kvs.Get(key))
}

// GetUint retrieves a value by its key as uint
func (kvs KVStore) GetUint(key string) uint {
	return cast.ToUint(kvs.Get(key))
}

// GetUint32 retrieves a value by its key as uint 32
func (kvs KVStore) GetUint32(key string) uint32 {
	return cast.ToUint32(kvs.Get(key))
}

// GetUint64 retrieves a value by its key as uint 64
func (kvs KVStore) GetUint64(key string) uint64 {
	return cast.ToUint64(kvs.Get(key))
}

// GetFloat64 retrieves a value by its key as float 64
func (kvs KVStore) GetFloat64(key string) float64 {
	return cast.ToFloat64(kvs.Get(key))
}

// GetTime retrieves a value by its key as time
func (kvs KVStore) GetTime(key string) time.Time {
	return cast.ToTime(kvs.Get(key))
}

// GetDuration retrieves a value by its key as duration
func (kvs KVStore) GetDuration(key string) time.Duration {
	return cast.ToDuration(kvs.Get(key))
}

// GetIntSlice retrieves a value by its key as int slice
func (kvs KVStore) GetIntSlice(key string) []int {
	return cast.ToIntSlice(kvs.Get(key))
}

// GetStringSlice retrieves a value by its key as string slice
func (kvs KVStore) GetStringSlice(key string) []string {
	return cast.ToStringSlice(kvs.Get(key))
}

// GetStringMap retrieves a value by its key as string map
func (kvs KVStore) GetStringMap(key string) map[string]interface{} {
	return cast.ToStringMap(kvs.Get(key))
}

// GetStringMapString retrieves a value by its key as string map string
func (kvs KVStore) GetStringMapString(key string) map[string]string {
	return cast.ToStringMapString(kvs.Get(key))
}

// GetStringMapStringSlice retrieves a value by its key as string map string slice
func (kvs KVStore) GetStringMapStringSlice(key string) map[string][]string {
	return cast.ToStringMapStringSlice(kvs.Get(key))
}

// MapStructure maps the value of a key to a pointer to an arbitrary struct
// This will raise an error if the value v is not a pointer
func (kvs KVStore) MapStructure(key string, v interface{}) error {
	return mapstructure.Decode(kvs.Get(key), v)
}

// SetFromJSON decodes JSON bytes into a map[string]interface{} and stores the decoded value
func (kvs KVStore) SetFromJSON(key string, data []byte) error {
	val := map[string]interface{}{}
	if err := json.Unmarshal(data, &val); err != nil {
		return err
	}
	kvs.Set(key, val)
	return nil
}
