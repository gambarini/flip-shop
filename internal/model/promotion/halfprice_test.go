package promotion

import (
	"testing"

	"github.com/gambarini/flip-shop/internal/model/item"
)

func TestItemQtyHalfPricePromotion_Apply(t *testing.T) {
	tests := []struct {
		name         string
		purchasedQty int
		cartQty      int
		price        int64
		wantDiscount int64
	}{
		// buy 2 get 1 at half: 1 unit at price/2
		{"buy 2 get 1 half off", 2, 2, 1000, 500},
		// buy 4: 2 groups of 2 → 2 units at half price
		{"buy 4 get 2 half off", 2, 4, 1000, 1000},
		// buy 3 with group of 2: 1 group → 1 unit at half
		{"buy 3 get 1 half off", 2, 3, 1000, 500},
		// below threshold
		{"below threshold", 2, 1, 1000, 0},
		// odd price truncates
		{"odd price truncates", 2, 2, 999, 499},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := ItemQtyHalfPricePromotion{PurchasedItemSku: "SKU", PurchasedQty: tt.purchasedQty}
			var got int64
			ctx := PromotionContext{
				GetPurchased: func(sku item.Sku) (PurchasedItem, bool) {
					return PurchasedItem{Sku: "SKU", Price: tt.price, Qty: tt.cartQty}, true
				},
				AddPromo:    func(s item.Sku, q int) (int64, error) { return 0, nil },
				AddDiscount: func(s item.Sku, d int64) (int64, error) { got = d; return d, nil },
			}
			if err := p.Apply(ctx); err != nil {
				t.Fatalf("unexpected error: %v", err)
			}
			if got != tt.wantDiscount {
				t.Errorf("discount = %d, want %d", got, tt.wantDiscount)
			}
		})
	}
}
