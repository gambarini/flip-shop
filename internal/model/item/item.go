package item

import (
	"errors"
)

var (
	ErrItemNotAvailableReservation = errors.New("item not available for reservation")
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
		return ErrItemNotAvailableReservation
	}

	i.QtyReserved += toReserveQty

	return nil
}

// ReleaseItem
// Release reserved quantity of the item
func (i *Item) ReleaseItem(toReleaseQty int) error {

	if toReleaseQty > i.QtyReserved {
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
