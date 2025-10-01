package utils

import "errors"

type (
	StoreName string

	// KVDatabase
	// Interface defining a key/value database
	// It's expected that the implementation is thread safe
	KVDatabase interface {
		// WithTx
		// Enclosures logic that requires a transaction
		WithTx(txHandler TxHandler) error
		// Read
		// Return the value for a key
		// return ErrValueNotFound if key/value does not exist
		Read(name StoreName, key string) (interface{}, error)
	}

	// KVRepository
	// Interface defining repositories for key/value database
	// Enable a repository to handle transactions explicitly
	KVRepository interface {
		WithTx(txHandler TxHandler) error
	}

	// Tx
	// Defines the database read and write inside a transactional boundary
	Tx interface {
		// Read
		// Return the value for a key within a transaction
		// return ErrValueNotFound if key does not exist
		Read(name StoreName, key string) (interface{}, error)
		// Write
		// Write a value for a key within a transaction
		Write(name StoreName, key string, v interface{})
	}

	// TxHandler
	// Handles database operations in a transactional boundary
	TxHandler func(tx Tx) error
)

var (
	ErrValueNotFound = errors.New("value not found for key")
)
