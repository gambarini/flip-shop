package promotion

import (
	"github.com/gambarini/flip-shop/internal/model/item"
	"testing"
)

// Tests rounding behavior of ItemQtyPriceDiscountPercentagePromotion with non-even totals.
func TestItemQtyPriceDiscountPercentagePromotion_Rounding_Truncates(t *testing.T) {
	p := ItemQtyPriceDiscountPercentagePromotion{PurchasedItemSku: "SKU", PurchasedQty: 1, PercentageDiscount: 0.10}

	var got int64
	get := func(sku item.Sku) (PurchasedItem, bool) {
		return PurchasedItem{Sku: "SKU", Price: 10950, Qty: 3}, true // total = 32850
	}
	add := func(sku item.Sku, qty int) error { return nil }
	addDisc := func(sku item.Sku, discount int64) error { got = discount; return nil }

	if err := p.Apply(get, add, addDisc); err != nil {
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

	applyBoth := func() int64 {
		var discount int64
		get := func(sku item.Sku) (PurchasedItem, bool) { return PurchasedItem{Sku: buySku, Price: 1000, Qty: 4}, true }
		add := func(sku item.Sku, qty int) error { return nil }
		addDisc := func(sku item.Sku, d int64) error { discount += d; return nil }
		_ = p1.Apply(get, add, addDisc)
		_ = p2.Apply(get, add, addDisc)
		return discount
	}
	applyBothReverse := func() int64 {
		var discount int64
		get := func(sku item.Sku) (PurchasedItem, bool) { return PurchasedItem{Sku: buySku, Price: 1000, Qty: 4}, true }
		add := func(sku item.Sku, qty int) error { return nil }
		addDisc := func(sku item.Sku, d int64) error { discount += d; return nil }
		_ = p2.Apply(get, add, addDisc)
		_ = p1.Apply(get, add, addDisc)
		return discount
	}

	d1 := applyBoth()
	d2 := applyBothReverse()

	if d1 != 4000 || d2 != 4000 {
		t.Fatalf("expected total discount 4000 in both orders, got d1=%d d2=%d", d1, d2)
	}
}
