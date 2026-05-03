package promotion

import (
	"testing"

	"github.com/gambarini/flip-shop/internal/model/item"
)

func TestItemQtyFixedDiscountPromotion_Apply(t *testing.T) {
	tests := []struct {
		name          string
		purchasedQty  int
		fixedDiscount int64
		cartQty       int
		wantDiscount  int64
	}{
		{"meets threshold", 2, 500, 2, 500},
		{"exceeds threshold", 2, 500, 5, 500},
		{"below threshold", 2, 500, 1, 0},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := ItemQtyFixedDiscountPromotion{
				PurchasedItemSku: "SKU",
				PurchasedQty:     tt.purchasedQty,
				FixedDiscount:    tt.fixedDiscount,
			}
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
				t.Errorf("discount = %d, want %d", got, tt.wantDiscount)
			}
		})
	}
}

func TestItemQtyFixedDiscountPromotion_Apply_NotInCart(t *testing.T) {
	p := ItemQtyFixedDiscountPromotion{PurchasedItemSku: "SKU", PurchasedQty: 1, FixedDiscount: 100}
	called := false
	ctx := PromotionContext{
		GetPurchased: func(sku item.Sku) (PurchasedItem, bool) { return PurchasedItem{}, false },
		AddPromo:     func(s item.Sku, q int) (int64, error) { called = true; return 0, nil },
		AddDiscount:  func(s item.Sku, d int64) (int64, error) { called = true; return d, nil },
	}
	if err := p.Apply(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Fatal("expected no handler calls when item not in cart")
	}
}
