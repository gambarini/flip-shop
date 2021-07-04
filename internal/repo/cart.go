package repo

import (
	"errors"
	"github.com/gambarini/flip-shop/internal/model/cart"
	"github.com/gambarini/flip-shop/utils"
)

const (
	cartStoreName = utils.StoreName("Cart")
)

type (
	ICartRepository interface {
		utils.KVRepository
		FindCartByID(id string) (c cart.Cart, err error)
		Store(tx utils.Tx, c cart.Cart) (err error)
	}

	CartRepository struct {
		utils.KVDatabase
	}

)

var (
	ErrCartNotFound = errors.New("cart not found")
)

func NewCartRepository(kvDb utils.KVDatabase) *CartRepository {
	return &CartRepository{
		kvDb,
	}
}

func (repo CartRepository) FindCartByID(id string) (c cart.Cart, err error) {

	v, err := repo.KVDatabase.Read(cartStoreName, id)

	switch {
	case err == utils.ErrValueNotFound:
		return c, ErrCartNotFound
	case err != nil:
		return c, err
	default:
		return v.(cart.Cart), nil
	}
}

func (repo CartRepository) Store(tx utils.Tx, c cart.Cart) (err error) {

	tx.Write(cartStoreName, string(c.CartID), c)

	return nil

}

