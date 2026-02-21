package repo

import (
	"errors"

	"github.com/gambarini/flip-shop/internal/model/item"
	"github.com/gambarini/flip-shop/utils"
)

const (
	// ItemStoreName is the store name for items in the KV database.
	ItemStoreName = utils.StoreName("Items")
)

type (
	// IItemRepository exposes item persistence operations against a KV database.
	IItemRepository interface {
		utils.KVRepository
		// FindItemBySku loads an item by its SKU using the provided transaction.
		FindItemBySku(tx utils.Tx, sku item.Sku) (item item.Item, err error)
		// Store persists the given item within the provided transaction.
		Store(tx utils.Tx, item item.Item) (err error)
		// ListItems returns all items currently stored.
		ListItems() ([]item.Item, error)
	}

	// ItemRepository is a concrete implementation of IItemRepository backed by a KVDatabase.
	ItemRepository struct {
		utils.KVDatabase
	}
)

var (
	// ErrItemNotFound is returned when an item cannot be found in the store.
	ErrItemNotFound = errors.New("item not found")
)

// NewItemRepository creates a new ItemRepository using the provided KV database.
func NewItemRepository(kvDb utils.KVDatabase) *ItemRepository {
	return &ItemRepository{
		kvDb,
	}
}

// FindItemBySku reads an item by SKU from the underlying KV database using the transaction.
func (repo ItemRepository) FindItemBySku(tx utils.Tx, sku item.Sku) (i item.Item, err error) {

	v, err := tx.Read(ItemStoreName, string(sku))

	switch {
	case errors.Is(err, utils.ErrValueNotFound):
		return i, ErrItemNotFound
	case err != nil:
		return i, err
	default:
		return v.(item.Item), nil
	}
}

// Store writes an item into the KV database within the given transaction.
func (repo ItemRepository) Store(tx utils.Tx, i item.Item) (err error) {

	tx.Write(ItemStoreName, string(i.Sku), i)

	return nil

}

// ListItems returns all items from the underlying store.
func (repo ItemRepository) ListItems() ([]item.Item, error) {
	vals, err := repo.KVDatabase.List(ItemStoreName)
	if err != nil {
		return nil, err
	}
	items := make([]item.Item, 0, len(vals))
	for _, v := range vals {
		if it, ok := v.(item.Item); ok {
			items = append(items, it)
		}
	}
	return items, nil
}
