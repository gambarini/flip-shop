package promotion

import (
	"testing"

	"github.com/gambarini/flip-shop/internal/model/item"
)

func TestCartSpendThresholdPromotion_Apply_MeetsThreshold(t *testing.T) {
	p := CartSpendThresholdPromotion{ThresholdAmount: 5000, PercentageDiscount: 0.10}

	items := []PurchasedItem{
		{Sku: "A", Price: 3000, Qty: 1},
		{Sku: "B", Price: 2000, Qty: 1},
	}
	discounts := map[item.Sku]int64{}

	ctx := PromotionContext{
		GetPurchased:    func(sku item.Sku) (PurchasedItem, bool) { return PurchasedItem{}, false },
		AddPromo:        func(s item.Sku, q int) (int64, error) { return 0, nil },
		AddDiscount:     func(s item.Sku, d int64) (int64, error) { discounts[s] += d; return d, nil },
		GetAllPurchased: func() []PurchasedItem { return items },
		GetCartTotal:    func() int64 { return 5000 },
	}

	if err := p.Apply(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// A: 3000/100=30 * 0.10 * 100 = 300
	if discounts["A"] != 300 {
		t.Errorf("A discount = %d, want 300", discounts["A"])
	}
	// B: 2000/100=20 * 0.10 * 100 = 200
	if discounts["B"] != 200 {
		t.Errorf("B discount = %d, want 200", discounts["B"])
	}
}

func TestCartSpendThresholdPromotion_Apply_BelowThreshold_Noops(t *testing.T) {
	p := CartSpendThresholdPromotion{ThresholdAmount: 10000, PercentageDiscount: 0.10}
	called := false
	ctx := PromotionContext{
		GetPurchased:    func(sku item.Sku) (PurchasedItem, bool) { return PurchasedItem{}, false },
		AddPromo:        func(s item.Sku, q int) (int64, error) { return 0, nil },
		AddDiscount:     func(s item.Sku, d int64) (int64, error) { called = true; return d, nil },
		GetAllPurchased: func() []PurchasedItem { return []PurchasedItem{{Sku: "A", Price: 1000, Qty: 1}} },
		GetCartTotal:    func() int64 { return 1000 },
	}
	if err := p.Apply(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Fatal("discount should not apply when below threshold")
	}
}
