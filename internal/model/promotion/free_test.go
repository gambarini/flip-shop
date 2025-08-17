package promotion

import (
	"github.com/gambarini/flip-shop/internal/model/item"
	"testing"
)

func TestFreeItemPromotion_Apply_MultiplesAccumulate(t *testing.T) {
	f := FreeItemPromotion{
		PurchasedItemSku: "BUYSKU",
		FreeItemSku:      "FREESKU",
		FreeItemPrice:    300,
	}

	var gotAddQty int
	var gotDiscount int64

	get := func(sku item.Sku) (PurchasedItem, bool) {
		return PurchasedItem{Sku: "BUYSKU", Price: 500, Qty: 3}, true
	}
	add := func(sku item.Sku, qty int) error {
		if sku != "FREESKU" {
			t.Fatalf("unexpected free sku: %s", sku)
		}
		gotAddQty = qty
		return nil
	}
	addDisc := func(sku item.Sku, discount int64) error {
		if sku != "FREESKU" {
			t.Fatalf("unexpected discount sku: %s", sku)
		}
		gotDiscount = discount
		return nil
	}

	if err := f.Apply(get, add, addDisc); err != nil {
		t.Fatalf("Apply() unexpected error: %v", err)
	}

	if gotAddQty != 3 {
		t.Fatalf("expected add free qty=3, got %d", gotAddQty)
	}
	if gotDiscount != 900 { // FreeItemPrice * Qty
		t.Fatalf("expected discount 900, got %d", gotDiscount)
	}
}

func TestFreeItemPromotion_Apply_NoPurchased_Noops(t *testing.T) {
	f := FreeItemPromotion{PurchasedItemSku: "BUY", FreeItemSku: "FREE", FreeItemPrice: 100}
	called := false
	if err := f.Apply(func(sku item.Sku) (PurchasedItem, bool) { return PurchasedItem{}, false },
		func(s item.Sku, i int) error { called = true; return nil },
		func(s item.Sku, d int64) error { called = true; return nil },
	); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if called {
		t.Fatalf("expected no handler calls when item not purchased")
	}
}
