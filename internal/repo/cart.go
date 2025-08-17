package repo

import (
	"errors"

	"github.com/gambarini/flip-shop/internal/model/cart"
	"github.com/gambarini/flip-shop/utils"
)

const (
	// CartStoreName is the store name for carts in the KV database.
	CartStoreName = utils.StoreName("Cart")
)

type (
	// ICartRepository exposes cart persistence operations against a KV database.
	ICartRepository interface {
		utils.KVRepository
		// FindCartByID loads a cart by its identifier.
		FindCartByID(id string) (c cart.Cart, err error)
		// Store persists the given cart within the provided transaction.
		Store(tx utils.Tx, c cart.Cart) (err error)
	}

	// CartRepository is a concrete implementation of ICartRepository backed by a KVDatabase.
	CartRepository struct {
		utils.KVDatabase
	}
)

var (
	// ErrCartNotFound is returned when a cart cannot be found in the store.
	ErrCartNotFound = errors.New("cart not found")
)

// NewCartRepository creates a new CartRepository using the provided KV database.
func NewCartRepository(kvDb utils.KVDatabase) *CartRepository {
	return &CartRepository{
		kvDb,
	}
}

// FindCartByID reads a cart from the underlying KV database.
func (repo CartRepository) FindCartByID(id string) (c cart.Cart, err error) {

	v, err := repo.KVDatabase.Read(CartStoreName, id)

	switch {
	case errors.Is(err, utils.ErrValueNotFound):
		return c, ErrCartNotFound
	case err != nil:
		return c, err
	default:
		return v.(cart.Cart), nil
	}
}

// Store writes a cart into the KV database within the given transaction.
func (repo CartRepository) Store(tx utils.Tx, c cart.Cart) (err error) {

	tx.Write(CartStoreName, string(c.CartID), c)

	return nil

}
