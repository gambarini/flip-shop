package cart

import (
	"errors"

	"github.com/gambarini/flip-shop/internal/model/item"
	"github.com/gambarini/flip-shop/utils"
	"github.com/gofrs/uuid"
)

var (
	// ErrCartNotAvailable is returned when an operation is attempted on a non-available cart.
	ErrCartNotAvailable = errors.New("cart not available")
	// ErrItemQtyAddedInvalid indicates that a purchase update would result in a negative quantity.
	ErrItemQtyAddedInvalid = errors.New("item quantity invalid")
	// ErrItemNotInCart is returned when applying a discount to a non-existent cart item.
	ErrItemNotInCart = errors.New("item is not in the cart")
)

type (
	// Status represents the state of a cart.
	// For simplicity only 2 states are supported.
	Status string

	// Cart represents a shopping cart with purchases and totals.
	// Total is expressed in integer cents (int64).
	Cart struct {
		CartID     string
		Purchases  map[item.Sku]Purchase
		CartStatus Status
		Total      int64
	}

	// Purchase captures an item purchase in the cart, including discount applied.
	// Price and Discount are expressed in integer cents (int64).
	Purchase struct {
		Sku      item.Sku
		Name     string
		Price    int64
		Qty      int
		Discount int64
	}
)

const (
	// CartStatusAvailable indicates the cart can receive purchases.
	CartStatusAvailable = Status("Available")

	// CartStatusSubmitted indicates the cart has been submitted and no longer accepts purchases.
	CartStatusSubmitted = Status("Submitted")
)

// NewAvailableCart creates a new cart in Available status with an auto-generated ID.
func NewAvailableCart() (cart Cart) {

	id, _ := uuid.NewV4()

	return Cart{
		CartID:     id.String(),
		CartStatus: CartStatusAvailable,
		Purchases:  make(map[item.Sku]Purchase),
	}
}

// PurchaseItem updates the cart with a purchase for the given item and quantity.
// Quantity may be negative to remove items; zero removes the item entry.
func (c *Cart) PurchaseItem(i item.Item, qty int) (err error) {

	if c.CartStatus != CartStatusAvailable {
		return ErrCartNotAvailable
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

	// no-op: all branches handled above
	return nil
}

// DiscountPurchase adds a discount to an existing purchase by SKU.
func (c *Cart) DiscountPurchase(sku item.Sku, discount int64) (err error) {

	p, ok := c.Purchases[sku]

	if !ok {
		return ErrItemNotInCart
	}

	p.Discount += discount

	c.Purchases[sku] = p

	return nil
}

// SubmitCart finalizes the cart total and moves it to Submitted status.
func (c *Cart) SubmitCart() (err error) {

	if c.CartStatus != CartStatusAvailable {

		return ErrCartNotAvailable
	}

	for _, p := range c.Purchases {
		line := utils.SaturatingMulInt64Int(p.Price, p.Qty)
		line = utils.SaturatingSubInt64(line, p.Discount)
		c.Total = utils.SaturatingAddInt64(c.Total, line)
	}

	c.CartStatus = CartStatusSubmitted

	return nil
}
