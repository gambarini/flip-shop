package item

import (
	"errors"
)

var (
	// ErrItemNotAvailableReservation indicates the requested reservation exceeds availability.
	ErrItemNotAvailableReservation = errors.New("item not available for reservation")
	// ErrInvalidReleaseQuantity indicates releasing more than reserved.
	ErrInvalidReleaseQuantity = errors.New("cannot release quantity")
	// ErrInvalidRemoveQuantity indicates removing more than reserved.
	ErrInvalidRemoveQuantity = errors.New("cannot remove quantity")
)

type (
	// Sku identifies a product uniquely.
	Sku string

	// Item represents an item available for purchase.
	// This model controls the quantities and reservations available
	// in the same object for the sake of simplicity.
	Item struct {
		Sku          Sku
		Name         string
		Price        int64
		QtyAvailable int
		QtyReserved  int
	}
)

// NewItem is a constructor that creates a new Item with the provided attributes.
// QtyReserved is initialized to 0; validations (non-negative qty/price, non-empty sku) are responsibility of callers.
func NewItem(sku Sku, name string, price int64, qty int) Item {
	return Item{
		Sku:          sku,
		Name:         name,
		Price:        price,
		QtyAvailable: qty,
		QtyReserved:  0,
	}
}

// Restock increases the available inventory by the given amount, preserving QtyReserved.
// Callers should validate non-negative quantities.
func (i *Item) Restock(addQty int) {
	i.QtyAvailable += addQty
}

// AdjustPrice sets the item's price to the provided value.
// Callers should validate that price is >= 0 before calling.
func (i *Item) AdjustPrice(newPrice int64) {
	i.Price = newPrice
}

// ReserveItem reserves a quantity of the item if available.
func (i *Item) ReserveItem(toReserveQty int) error {

	if toReserveQty > (i.QtyAvailable - i.QtyReserved) {
		return ErrItemNotAvailableReservation
	}

	i.QtyReserved += toReserveQty

	return nil
}

// ReleaseItem releases reserved quantity of the item.
func (i *Item) ReleaseItem(toReleaseQty int) error {

	if toReleaseQty > i.QtyReserved {
		return ErrInvalidReleaseQuantity
	}

	i.QtyReserved -= toReleaseQty

	return nil
}

// RemoveItem removes the available quantity for the item and decreases reserved accordingly.
func (i *Item) RemoveItem(toRemoveQty int) error {

	if toRemoveQty > i.QtyReserved {
		return ErrInvalidRemoveQuantity
	}

	i.QtyAvailable -= toRemoveQty
	i.QtyReserved -= toRemoveQty

	return nil
}
