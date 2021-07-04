package repo

import (
	"errors"
	"github.com/gambarini/flip-shop/internal/model/item"
	"github.com/gambarini/flip-shop/utils"
)

const (
	itemStoreName = utils.StoreName("Items")
)

type (
	IItemRepository interface {
		utils.KVRepository
		FindItemBySku(tx utils.Tx, sku item.Sku) (item item.Item, err error)
		Store(tx utils.Tx, item item.Item) (err error)
	}

	ItemRepository struct {
		utils.KVDatabase
	}

)

var (
	ErrItemNotFound = errors.New("item not found")
)

func NewItemRepository(kvDb utils.KVDatabase) *ItemRepository {
	return &ItemRepository{
		kvDb,
	}
}

func (repo ItemRepository) FindItemBySku(tx utils.Tx, sku item.Sku) (i item.Item, err error) {

	v, err := tx.Read(itemStoreName, string(sku))

	switch {
	case err == utils.ErrValueNotFound:
		return i, ErrItemNotFound
	case err != nil:
		return i, err
	default:
		return v.(item.Item), nil
	}
}

func (repo ItemRepository) Store(tx utils.Tx, i item.Item) (err error) {

	tx.Write(itemStoreName, string(i.Sku), i)

	return nil

}
