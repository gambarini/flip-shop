package promotion

import (
	"github.com/gambarini/flip-shop/internal/model/item"
	"testing"
)

// Tests rounding behavior of ItemQtyPriceDiscountPercentagePromotion with non-even totals.
func TestItemQtyPriceDiscountPercentagePromotion_Rounding_Truncates(t *testing.T) {
	p := ItemQtyPriceDiscountPercentagePromotion{PurchasedItemSku: "SKU", PurchasedQty: 1, PercentageDiscount: 0.10}

	var got int64
	ctx := PromotionContext{
		GetPurchased: func(sku item.Sku) (PurchasedItem, bool) {
			return PurchasedItem{Sku: "SKU", Price: 10950, Qty: 3}, true // total = 32850
		},
		AddPromo:    func(sku item.Sku, qty int) (int64, error) { return 0, nil },
		AddDiscount: func(sku item.Sku, discount int64) (int64, error) { got = discount; return discount, nil },
	}

	if err := p.Apply(ctx); err != nil {
		t.Fatalf("Apply() error: %v", err)
	}
	// With current algorithm: total/100 = 328, 328 * 0.1 * 100 = 3280 (truncated), not 3285
	if got != 3280 {
		t.Fatalf("expected truncated discount 3280, got %d", got)
	}
}

// Tests combined effect when two promotions affect the same purchased SKU; order should not matter for additive discounts.
func TestPromotions_Interactions_SameSKU_AdditiveAndOrderIndependent(t *testing.T) {
	buySku := item.Sku("SKU")
	p1 := ItemQtyPriceFreePromotion{PurchasedItemSku: buySku, PurchasedQty: 2}
	p2 := ItemQtyPriceDiscountPercentagePromotion{PurchasedItemSku: buySku, PurchasedQty: 1, PercentageDiscount: 0.5}

	makeCtx := func(discountAcc *int64) PromotionContext {
		return PromotionContext{
			GetPurchased: func(sku item.Sku) (PurchasedItem, bool) {
				return PurchasedItem{Sku: buySku, Price: 1000, Qty: 4}, true
			},
			AddPromo:    func(sku item.Sku, qty int) (int64, error) { return 0, nil },
			AddDiscount: func(sku item.Sku, d int64) (int64, error) { *discountAcc += d; return d, nil },
		}
	}

	var d1 int64
	ctx1 := makeCtx(&d1)
	_ = p1.Apply(ctx1)
	_ = p2.Apply(ctx1)

	var d2 int64
	ctx2 := makeCtx(&d2)
	_ = p2.Apply(ctx2)
	_ = p1.Apply(ctx2)

	if d1 != 4000 || d2 != 4000 {
		t.Fatalf("expected total discount 4000 in both orders, got d1=%d d2=%d", d1, d2)
	}
}
