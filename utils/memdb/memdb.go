// Package memdb provides a simple in-memory key/value database with
// mutex-guarded serializable isolation. Each call to WithTx executes inside a
// critical section protected by a global mutex, ensuring no interleaving between
// concurrent transactions. Writes are applied using a copy-on-write snapshot per
// transaction: changes are committed atomically only if the handler returns nil;
// otherwise, they are discarded (rollback).
package memdb

import (
	"sync"

	"github.com/gambarini/flip-shop/utils"
)

type (
	// MemoryKVDatabase
	// Key/value memory database
	// Thread safe for concurrent read/write access
	// Serializable isolation via global mutex and copy-on-write transactional semantics.
	MemoryKVDatabase struct {
		lock sync.RWMutex
		tx   *MemoryKVTx
	}

	// MemoryKVTx represents a transaction view over the underlying data.
	// It holds a map of stores, each a map from key to value.
	MemoryKVTx struct {
		data map[utils.StoreName]map[string]interface{}
	}
)

// NewMemoryKVDatabase creates a new in-memory key/value database.
func NewMemoryKVDatabase() *MemoryKVDatabase {
	return &MemoryKVDatabase{
		lock: sync.RWMutex{},
		tx:   &MemoryKVTx{data: make(map[utils.StoreName]map[string]interface{})},
	}
}

// NewDMemoryKVDatabase is deprecated; use NewMemoryKVDatabase instead.
func NewDMemoryKVDatabase() *MemoryKVDatabase {
	return NewMemoryKVDatabase()
}

func (tx MemoryKVTx) Read(name utils.StoreName, key string) (v interface{}, err error) {
	v, ok := tx.data[name][key]
	if !ok {
		return v, utils.ErrValueNotFound
	}
	return v, nil
}

func (tx MemoryKVTx) Write(name utils.StoreName, key string, v interface{}) {
	if _, ok := tx.data[name]; !ok {
		tx.data[name] = map[string]interface{}{}
	}
	tx.data[name][key] = v
}

// cloneData performs a shallow copy of the top-level store map and each inner
// key/value map. Values are copied by reference, which is acceptable given the
// domain models in this project are treated as immutable within a transaction
// unless overwritten explicitly.
func cloneData(src map[utils.StoreName]map[string]interface{}) map[utils.StoreName]map[string]interface{} {
	copy := make(map[utils.StoreName]map[string]interface{}, len(src))
	for store, kv := range src {
		inner := make(map[string]interface{}, len(kv))
		for k, v := range kv {
			inner[k] = v
		}
		copy[store] = inner
	}
	return copy
}

func (mDb *MemoryKVDatabase) WithTx(txHandler utils.TxHandler) error {
	// Ensure serializable isolation across transactions
	mDb.lock.Lock()
	defer mDb.lock.Unlock()

	// Create a transactional snapshot (copy-on-write)
	snapshot := cloneData(mDb.tx.data)
	tx := &MemoryKVTx{data: snapshot}

	// Execute user handler against the snapshot
	if err := txHandler(tx); err != nil {
		// rollback by discarding snapshot
		return err
	}

	// Commit by replacing the live data with the snapshot
	mDb.tx = tx
	return nil
}

func (mDb *MemoryKVDatabase) Read(name utils.StoreName, key string) (v interface{}, err error) {
	mDb.lock.RLock()
	defer mDb.lock.RUnlock()
	v, ok := mDb.tx.data[name][key]
	if !ok {
		return v, utils.ErrValueNotFound
	}
	return v, nil
}

// List returns a snapshot slice with all values for a given store name.
func (mDb *MemoryKVDatabase) List(name utils.StoreName) ([]interface{}, error) {
	mDb.lock.RLock()
	defer mDb.lock.RUnlock()
	store, ok := mDb.tx.data[name]
	if !ok {
		return []interface{}{}, nil
	}
	res := make([]interface{}, 0, len(store))
	for _, v := range store {
		res = append(res, v)
	}
	return res, nil
}
