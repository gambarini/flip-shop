package cart

import (
	"github.com/gambarini/flip-shop/internal/model/item"
	"reflect"
	"testing"
)

func TestCart_PurchaseItem(t *testing.T) {
	newCart := func(s Status) Cart {
		return Cart{
			CartID:     "CartID",
			Purchases:  map[item.Sku]Purchase{},
			CartStatus: s,
			Total:      0,
		}
	}

	newCartWithItem := func(s Status) Cart {
		return Cart{
			CartID:     "CartID",
			Purchases:  map[item.Sku]Purchase{item.Sku("TEST"): {
				Sku:      item.Sku("TEST"),
				Name:     "Test",
				Price:    1000,
				Qty:      3,
				Discount: 0,
			}},
			CartStatus: s,
			Total:      0,
		}
	}

	newItem := func() item.Item {
		return item.Item{
			Sku:   item.Sku("TEST"),
			Name:  "Test",
			Price: 1000,
		}
	}

	type args struct {
		i    item.Item
		qty  int
		cart Cart
	}
	type want struct {
		pLen int
		iQty int
	}
	tests := []struct {
		name    string
		args    args
		want    want
		wantErr bool
	}{
		{"Purchase 1", args{newItem(), 1, newCart(CartStatusAvailable)}, want{1, 1}, false},
		{"Purchase 5", args{newItem(), 5, newCart(CartStatusAvailable)}, want{1, 5}, false},
		{"Remove 2", args{newItem(), -2, newCartWithItem(CartStatusAvailable)}, want{1, 1}, false},
		{"Remove 3", args{newItem(), -3, newCartWithItem(CartStatusAvailable)}, want{0, 0}, false},
		{"Add 1", args{newItem(), 1, newCartWithItem(CartStatusAvailable)}, want{1, 4}, false},
		{"Remove 4 with error", args{newItem(), -4, newCartWithItem(CartStatusAvailable)}, want{1, 3}, true},
		{"Status with error", args{newItem(), 1, newCartWithItem(CartStatusSubmitted)}, want{1, 3}, true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := tt.args.cart

			if err := c.PurchaseItem(tt.args.i, tt.args.qty); (err != nil) != tt.wantErr {
				t.Errorf("PurchaseItem() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(len(c.Purchases), tt.want.pLen) {
				t.Errorf("total purchases = %v, want %v", len(c.Purchases), tt.want.pLen)
			}

			if !reflect.DeepEqual(c.Purchases[tt.args.i.Sku].Qty, tt.want.iQty) {
				t.Errorf("item quantity = %v, want %v", c.Purchases[tt.args.i.Sku].Qty, tt.want.iQty)
			}
		})
	}
}





