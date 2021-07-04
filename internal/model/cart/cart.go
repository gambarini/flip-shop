package cart

import (
	"errors"
	"github.com/gambarini/flip-shop/internal/model/item"
	//"github.com/gambarini/flip-shop/internal/model/promotion"
	"github.com/gofrs/uuid"
)

var (
	ErrICartNotAvailable   = errors.New("cart not Available")
	ErrItemQtyAddedInvalid = errors.New("item quantity invalid")
	ErrItemNotInCart       = errors.New("item is not in the cart")
)

type (
	// Status
	// Supported Cart Statuses
	// For simplicity only 2 states are supported
	Status string

	Cart struct {
		CartID     string
		Purchases  map[item.Sku]Purchase
		CartStatus Status
		Total      int64
	}

	Purchase struct {
		Sku      item.Sku
		Name     string
		Price    int64
		Qty      int
		Discount int64
	}

)

const (
	// CartStatusAvailable
	// Cart is ready to receive purchases
	CartStatusAvailable = Status("Available")

	// CartStatusSubmitted
	// Cart is submitted and does not accept purchases
	CartStatusSubmitted = Status("Submitted")
)

func NewAvailableCart() (cart Cart) {

	id, _ := uuid.NewV4()

	return Cart{
		CartID:     id.String(),
		CartStatus: CartStatusAvailable,
		Purchases:  make(map[item.Sku]Purchase),
	}
}

func (c *Cart) PurchaseItem(i item.Item, qty int) (err error) {

	if c.CartStatus != CartStatusAvailable {
		return ErrICartNotAvailable
	}

	p, ok := c.Purchases[i.Sku]

	if !ok {
		p = Purchase{
			Sku:      i.Sku,
			Name:     i.Name,
			Price:    i.Price,
			Qty:      0,
			Discount: 0,
		}
	}

	fQty := p.Qty + qty

	switch {
	case fQty < 0:
		return ErrItemQtyAddedInvalid
	case fQty == 0:
		delete(c.Purchases, i.Sku)
		return nil
	case fQty > 0:
		p.Qty += qty
		c.Purchases[i.Sku] = p
		return nil
	}

	return nil
}

func (c *Cart) DiscountPurchase(sku item.Sku, discount int64) (err error) {

	p, ok := c.Purchases[sku]

	if !ok {
		return ErrItemNotInCart
	}

	p.Discount += discount

	c.Purchases[sku] = p

	return nil
}

func (c *Cart) SubmitCart() (err error) {

	if c.CartStatus != CartStatusAvailable {

		return ErrICartNotAvailable
	}

	for _, p := range c.Purchases {
		c.Total += (p.Price * int64(p.Qty)) - p.Discount
	}

	c.CartStatus = CartStatusSubmitted

	return nil
}
