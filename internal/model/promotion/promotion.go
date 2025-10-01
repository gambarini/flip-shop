package promotion

import (
	"github.com/gambarini/flip-shop/internal/model/item"
)

type (
	// Promotion
	// Defines the promotion interface that allows to apply changes to a cart
	// by adding purchases or discounts through delegate handlers
	Promotion interface {
		Apply(getPurchasedHandler GetPurchasedItemHandler, addPromoHandler AddPromoItemToCartHandler, AddDiscountHandler AddDiscountToCartHandler) (err error)
	}

	PurchasedItem struct {
		Sku      item.Sku
		Name     string
		Price    int64
		Qty      int
		Discount int64
	}

	// GetPurchasedItemHandler
	// Delegates the ability find the purchased item in the cart
	GetPurchasedItemHandler func(sku item.Sku) (PurchasedItem, bool)

	// AddPromoItemToCartHandler
	// Delegates the ability to add items to a cart
	AddPromoItemToCartHandler func(i item.Sku, qty int) error

	// AddDiscountToCartHandler
	// Delegates the ability to add discounts to a cart
	AddDiscountToCartHandler func(discountItemSku item.Sku, discount int64) error
)
