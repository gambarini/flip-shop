package promotion

import (
	"github.com/gambarini/flip-shop/internal/model/item"
	"reflect"
	"testing"
)

func TestItemQtyPriceDiscountPercentagePromotion_Apply(t *testing.T) {
	type promoConfig struct {
		purchasedItemSku   item.Sku
		purchasedQty       int
		percentageDiscount float32
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
		{"2 item .5 disc",
			promoConfig{"TEST", 1, 0.5},
			args{2},
			want{discount: 1000},
			false},
		{"4 item .5 disc",
			promoConfig{"TEST", 2, 0.5},
			args{4},
			want{discount: 2000},
			false},
		{"4 item no disc",
			promoConfig{"TEST", 4, 0.5},
			args{4},
			want{discount: 0},
			false},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fIP := ItemQtyPriceDiscountPercentagePromotion{
				PurchasedItemSku:   tt.promoConfig.purchasedItemSku,
				PurchasedQty:       tt.promoConfig.purchasedQty,
				PercentageDiscount: tt.promoConfig.percentageDiscount,
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
