package promotion

import (
	"testing"

	"github.com/gambarini/flip-shop/internal/model/item"
)

func TestBundleDiscountPromotion_Apply_AllPresent(t *testing.T) {
	p := BundleDiscountPromotion{
		ItemSkus:           []item.Sku{"A", "B"},
		PercentageDiscount: 0.10,
	}

	cart := map[item.Sku]PurchasedItem{
		"A": {Sku: "A", Price: 2000, Qty: 1},
		"B": {Sku: "B", Price: 1000, Qty: 2},
	}
	discounts := map[item.Sku]int64{}

	ctx := PromotionContext{
		GetPurchased: func(sku item.Sku) (PurchasedItem, bool) {
			pu, ok := cart[sku]
			return pu, ok
		},
		AddPromo: func(s item.Sku, q int) (int64, error) { return 0, nil },
		AddDiscount: func(s item.Sku, d int64) (int64, error) {
			discounts[s] += d
			return d, nil
		},
	}

	if err := p.Apply(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	// A: 2000/100=20 * 0.10 * 100 = 200
	if discounts["A"] != 200 {
		t.Errorf("A discount = %d, want 200", discounts["A"])
	}
	// B: 2*1000=2000/100=20 * 0.10 * 100 = 200
	if discounts["B"] != 200 {
		t.Errorf("B discount = %d, want 200", discounts["B"])
	}
}

func TestBundleDiscountPromotion_Apply_MissingItem_Noops(t *testing.T) {
	p := BundleDiscountPromotion{
		ItemSkus:           []item.Sku{"A", "B"},
		PercentageDiscount: 0.10,
	}

	called := false
	ctx := PromotionContext{
		GetPurchased: func(sku item.Sku) (PurchasedItem, bool) {
			if sku == "A" {
				return PurchasedItem{Sku: "A", Price: 2000, Qty: 1}, true
			}
			return PurchasedItem{}, false // B missing
		},
		AddPromo:    func(s item.Sku, q int) (int64, error) { return 0, nil },
		AddDiscount: func(s item.Sku, d int64) (int64, error) { called = true; return d, nil },
	}

	if err := p.Apply(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Fatal("discount should not be applied when bundle is incomplete")
	}
}
