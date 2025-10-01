package promotion

import (
	"github.com/gambarini/flip-shop/internal/model/item"
	"reflect"
	"testing"
)

func TestItemQtyPriceDiscountPromotion_Apply(t *testing.T) {
	type promoConfig struct {
		purchasedItemSku item.Sku
		purchasedQty     int
	}
	type args struct {
		qty int
	}
	type want struct {
		discount int64
	}
	tests := []struct {
		name        string
		promoConfig promoConfig
		args        args
		want        want
		wantErr     bool
	}{
		{"2 items 1 free",
			promoConfig{"TEST", 2},
			args{qty: 2},
			want{1000},
			false},
		{"4 items 2 free",
			promoConfig{"TEST", 2},
			args{qty: 4},
			want{2000},
			false},
		{"5 items 2 free",
			promoConfig{"TEST", 2},
			args{qty: 5},
			want{2000},
			false},
		{"3 items 0 free",
			promoConfig{"TEST", 4},
			args{qty: 3},
			want{0},
			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fIP := ItemQtyPriceFreePromotion{
				PurchasedItemSku: tt.promoConfig.purchasedItemSku,
				PurchasedQty:     tt.promoConfig.purchasedQty,
			}

			var expcDiscount int64

			if err := fIP.Apply(func(sku item.Sku) (PurchasedItem, bool) {
				return PurchasedItem{
					Sku:      "TEST",
					Name:     "TEST",
					Price:    1000,
					Qty:      tt.args.qty,
					Discount: 0,
				}, true
			},
				func(i item.Sku, qty int) error { return nil },
				func(discountItemSku item.Sku, discount int64) error {
					expcDiscount = discount
					return nil
				}); (err != nil) != tt.wantErr {
				t.Errorf("Apply() error = %v, wantErr %v", err, tt.wantErr)
			}

			if !reflect.DeepEqual(expcDiscount, tt.want.discount) {
				t.Errorf("discount = %v, want %v", expcDiscount, tt.want.discount)
			}
		})
	}
}
