package promotion

import (
	"github.com/gambarini/flip-shop/internal/model/item"
	"testing"
)

func TestFreeItemPromotion_Apply_MultiplesAccumulate(t *testing.T) {
	f := FreeItemPromotion{
		PurchasedItemSku: "BUYSKU",
		FreeItemSku:      "FREESKU",
	}

	var gotAddQty int
	var gotDiscount int64

	get := func(sku item.Sku) (PurchasedItem, bool) {
		return PurchasedItem{Sku: "BUYSKU", Price: 500, Qty: 3}, true
	}
	add := func(sku item.Sku, qty int) (int64, error) {
		if sku != "FREESKU" {
			t.Fatalf("unexpected free sku: %s", sku)
		}
		gotAddQty = qty
		return int64(qty) * 300, nil // simulate 300 per unit charged
	}
	addDisc := func(sku item.Sku, discount int64) (int64, error) {
		if sku != "FREESKU" {
			t.Fatalf("unexpected discount sku: %s", sku)
		}
		gotDiscount = discount
		return discount, nil
	}

	ctx := PromotionContext{GetPurchased: get, AddPromo: add, AddDiscount: addDisc}
	if err := f.Apply(ctx); err != nil {
		t.Fatalf("Apply() unexpected error: %v", err)
	}

	if gotAddQty != 3 {
		t.Fatalf("expected add free qty=3, got %d", gotAddQty)
	}
	if gotDiscount != 900 { // charged 300*3=900
		t.Fatalf("expected discount 900, got %d", gotDiscount)
	}
}

func TestFreeItemPromotion_Apply_FreeItemOutOfStock_Skips(t *testing.T) {
	f := FreeItemPromotion{PurchasedItemSku: "BUY", FreeItemSku: "FREE"}

	discountCalled := false
	err := f.Apply(PromotionContext{
		GetPurchased: func(sku item.Sku) (PurchasedItem, bool) {
			return PurchasedItem{Sku: "BUY", Price: 500, Qty: 1}, true
		},
		AddPromo: func(sku item.Sku, qty int) (int64, error) {
			return 0, item.ErrItemNotAvailableReservation
		},
		AddDiscount: func(sku item.Sku, discount int64) (int64, error) {
			discountCalled = true
			return discount, nil
		},
	})

	if err != nil {
		t.Fatalf("Apply() should not error when free item is out of stock, got: %v", err)
	}
	if discountCalled {
		t.Fatal("discount handler should not be called when free item is unavailable")
	}
}

func TestFreeItemPromotion_Apply_NoPurchased_Noops(t *testing.T) {
	f := FreeItemPromotion{PurchasedItemSku: "BUY", FreeItemSku: "FREE"}
	called := false
	ctx := PromotionContext{
		GetPurchased: func(sku item.Sku) (PurchasedItem, bool) { return PurchasedItem{}, false },
		AddPromo:     func(s item.Sku, i int) (int64, error) { called = true; return 0, nil },
		AddDiscount:  func(s item.Sku, d int64) (int64, error) { called = true; return d, nil },
	}
	if err := f.Apply(ctx); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Fatalf("expected no handler calls when item not purchased")
	}
}
