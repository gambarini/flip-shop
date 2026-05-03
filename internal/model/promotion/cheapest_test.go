package promotion

import (
	"testing"

	"github.com/gambarini/flip-shop/internal/model/item"
)

func TestCheapestItemFreePromotion_Apply_MeetsCount(t *testing.T) {
	p := CheapestItemFreePromotion{MinItemCount: 3}

	items := []PurchasedItem{
		{Sku: "EXP", Price: 5000, Qty: 1},
		{Sku: "MID", Price: 2000, Qty: 1},
		{Sku: "CHE", Price: 1000, Qty: 1},
	}
	discounts := map[item.Sku]int64{}

	ctx := PromotionContext{
		GetPurchased:    func(sku item.Sku) (PurchasedItem, bool) { return PurchasedItem{}, false },
		AddPromo:        func(s item.Sku, q int) (int64, error) { return 0, nil },
		AddDiscount:     func(s item.Sku, d int64) (int64, error) { discounts[s] += d; return d, nil },
		GetAllPurchased: func() []PurchasedItem { return items },
		GetCartTotal:    func() int64 { return 8000 },
	}

	if err := p.Apply(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if discounts["CHE"] != 1000 {
		t.Errorf("cheapest discount = %d, want 1000", discounts["CHE"])
	}
	if discounts["EXP"] != 0 || discounts["MID"] != 0 {
		t.Error("only cheapest item should be discounted")
	}
}

func TestCheapestItemFreePromotion_Apply_BelowCount_Noops(t *testing.T) {
	p := CheapestItemFreePromotion{MinItemCount: 3}
	called := false
	ctx := PromotionContext{
		GetPurchased:    func(sku item.Sku) (PurchasedItem, bool) { return PurchasedItem{}, false },
		AddPromo:        func(s item.Sku, q int) (int64, error) { return 0, nil },
		AddDiscount:     func(s item.Sku, d int64) (int64, error) { called = true; return d, nil },
		GetAllPurchased: func() []PurchasedItem { return []PurchasedItem{{Sku: "A", Price: 1000, Qty: 2}} },
		GetCartTotal:    func() int64 { return 2000 },
	}
	if err := p.Apply(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Fatal("discount should not apply when below min item count")
	}
}

func TestCheapestItemFreePromotion_Apply_UsesQtyForCount(t *testing.T) {
	// qty of a single SKU should count toward the total
	p := CheapestItemFreePromotion{MinItemCount: 3}
	var gotSku item.Sku
	ctx := PromotionContext{
		GetPurchased:    func(sku item.Sku) (PurchasedItem, bool) { return PurchasedItem{}, false },
		AddPromo:        func(s item.Sku, q int) (int64, error) { return 0, nil },
		AddDiscount:     func(s item.Sku, d int64) (int64, error) { gotSku = s; return d, nil },
		GetAllPurchased: func() []PurchasedItem { return []PurchasedItem{{Sku: "A", Price: 500, Qty: 3}} },
		GetCartTotal:    func() int64 { return 1500 },
	}
	if err := p.Apply(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if gotSku != "A" {
		t.Errorf("expected discount on A, got %q", gotSku)
	}
}
