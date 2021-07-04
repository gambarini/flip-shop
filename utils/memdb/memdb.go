package memdb

import (
	"github.com/gambarini/flip-shop/utils"
	"sync"
)

type (
	// MemoryKVDatabase
	// Key/value memory database
	// Thread safe for concurrent read/write access
	// This version does not rollback changes if a error occurs in a transaction
	MemoryKVDatabase struct {
		lock sync.RWMutex
		tx *MemoryKVTx
	}

	MemoryKVTx struct {
		data map[utils.StoreName]map[string]interface{}
	}
)

func NewDMemoryKVDatabase() *MemoryKVDatabase {

	dm := make(map[utils.StoreName]map[string]interface{})

	return &MemoryKVDatabase{
		lock: sync.RWMutex{},
		tx: &MemoryKVTx{
			data: dm,
		},
	}
}

func (tx MemoryKVTx) Read(name utils.StoreName, key string) (v interface{}, err error) {
	v, ok := tx.data[name][key]

	if !ok {
		return v, utils.ErrValueNotFound
	}

	return tx.data[name][key], nil

}

func (tx MemoryKVTx) Write (name utils.StoreName, key string, v interface{}) {

	_, ok := tx.data[name]

	if !ok {
		tx.data[name] = map[string]interface{}{}
	}

	tx.data[name][key] = v

}

func (mDb *MemoryKVDatabase) WithTx(txHandler utils.TxHandler) error {
	mDb.lock.Lock()
	defer mDb.lock.Unlock()

	err := txHandler(mDb.tx)

	if err != nil {
		//TODO: Rollback changes
		return err
	}

	return nil

}

func (mDb *MemoryKVDatabase) Read(name utils.StoreName, key string) (v interface{}, err error) {
	mDb.lock.RLock()
	defer mDb.lock.RUnlock()
	v, ok := mDb.tx.data[name][key]

	if !ok {
		return v, utils.ErrValueNotFound
	}

	return mDb.tx.data[name][key], nil
}

