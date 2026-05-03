package promotion

import (
	"github.com/gambarini/flip-shop/internal/model/item"
)

type (
	Promotion interface {
		Apply(ctx PromotionContext) error
	}

	PurchasedItem struct {
		Sku      item.Sku
		Name     string
		Price    int64
		Qty      int
		Discount int64
	}

	GetPurchasedItemHandler     func(sku item.Sku) (PurchasedItem, bool)
	AddPromoItemToCartHandler   func(i item.Sku, qty int) (int64, error)
	AddDiscountToCartHandler    func(discountItemSku item.Sku, discount int64) (int64, error)
	GetAllPurchasedItemsHandler func() []PurchasedItem
	GetCartTotalHandler         func() int64
	// AddAppliedPromotionHandler records a named promotion and its total discount.
	// Callers must check for nil — it is optional in test contexts.
	AddAppliedPromotionHandler func(name string, discount int64)

	PromotionContext struct {
		GetPurchased    GetPurchasedItemHandler
		AddPromo        AddPromoItemToCartHandler
		AddDiscount     AddDiscountToCartHandler
		GetAllPurchased GetAllPurchasedItemsHandler
		GetCartTotal    GetCartTotalHandler
		AddApplied      AddAppliedPromotionHandler
	}
)
