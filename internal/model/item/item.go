package item

import (
	"errors"
	"fmt"
)

var (
	ErrItemNotAvailableReservation = func(sku Sku, Qty int) error {
		return errors.New(fmt.Sprintf("there is not %d Item(s) %s available for reservation", Qty, sku))
	}

	ErrInvalidReleaseQuantity = errors.New("cannot release quantity")
	ErrInvalidRemoveQuantity  = errors.New("cannot remove quantity")
)

type (
	Sku string

	// Item
	// Represents an item available for purchase
	// This model controls the quantities and reservations available
	// in the same object for the sake of simplicity
	Item struct {
		Sku          Sku
		Name         string
		Price        int64
		QtyAvailable int
		QtyReserved  int
	}
)

// ReserveItem
// Reserve a quantity of the item if available
func (i *Item) ReserveItem(toReserveQty int) error {

	if toReserveQty > (i.QtyAvailable - i.QtyReserved) {
		return ErrItemNotAvailableReservation(i.Sku, toReserveQty)
	}

	i.QtyReserved += toReserveQty

	return nil
}

// ReleaseItem
// Release reserved quantity of the item
func (i *Item) ReleaseItem(toReleaseQty int) error {

	if toReleaseQty > i.QtyAvailable {
		return ErrInvalidReleaseQuantity
	}

	i.QtyReserved -= toReleaseQty

	return nil
}

// RemoveItem
// Remove available quantity for the item
func (i *Item) RemoveItem(toRemoveQty int) error {

	if toRemoveQty > i.QtyReserved {
		return ErrInvalidRemoveQuantity
	}

	i.QtyAvailable -= toRemoveQty
	i.QtyReserved -= toRemoveQty

	return nil
}
