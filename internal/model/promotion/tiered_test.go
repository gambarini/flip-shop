package promotion

import (
	"testing"

	"github.com/gambarini/flip-shop/internal/model/item"
)

func TestItemQtyTieredDiscountPromotion_Apply(t *testing.T) {
	tiers := []DiscountTier{
		{MinQty: 2, PercentageDiscount: 0.10},
		{MinQty: 5, PercentageDiscount: 0.20},
	}

	tests := []struct {
		name         string
		cartQty      int
		wantDiscount int64
	}{
		{"below all tiers", 1, 0},
		{"hits tier 1 exactly", 2, 200},  // 2*1000=2000 → 2000/100=20 → 20*0.10*100=200
		{"between tiers", 3, 300},         // 3*1000=3000 → 30 → 30*0.10*100=300
		{"hits tier 2 exactly", 5, 1000}, // 5*1000=5000 → 50 → 50*0.20*100=1000
		{"above tier 2", 6, 1200},         // 6*1000=6000 → 60 → 60*0.20*100=1200
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := ItemQtyTieredDiscountPromotion{PurchasedItemSku: "SKU", Tiers: tiers}
			var got int64
			ctx := PromotionContext{
				GetPurchased: func(sku item.Sku) (PurchasedItem, bool) {
					return PurchasedItem{Sku: "SKU", Price: 1000, Qty: tt.cartQty}, true
				},
				AddPromo:    func(s item.Sku, q int) (int64, error) { return 0, nil },
				AddDiscount: func(s item.Sku, d int64) (int64, error) { got = d; return d, nil },
			}
			if err := p.Apply(ctx); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.wantDiscount {
				t.Errorf("qty=%d: discount = %d, want %d", tt.cartQty, got, tt.wantDiscount)
			}
		})
	}
}

func TestItemQtyTieredDiscountPromotion_Apply_UnsortedTiers(t *testing.T) {
	// Tiers provided in ascending order — Apply must sort them correctly
	p := ItemQtyTieredDiscountPromotion{
		PurchasedItemSku: "SKU",
		Tiers: []DiscountTier{
			{MinQty: 2, PercentageDiscount: 0.10},
			{MinQty: 5, PercentageDiscount: 0.20},
		},
	}
	var got int64
	ctx := PromotionContext{
		GetPurchased: func(sku item.Sku) (PurchasedItem, bool) {
			return PurchasedItem{Sku: "SKU", Price: 1000, Qty: 5}, true
		},
		AddPromo:    func(s item.Sku, q int) (int64, error) { return 0, nil },
		AddDiscount: func(s item.Sku, d int64) (int64, error) { got = d; return d, nil },
	}
	if err := p.Apply(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	// qty=5 should match the 20% tier, not the 10% tier
	if got != 1000 {
		t.Errorf("expected 1000 (20%% tier), got %d", got)
	}
}
