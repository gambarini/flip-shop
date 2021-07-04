package promotion

import (
	"github.com/gambarini/flip-shop/internal/model/item"
	"testing"
)

func TestFreeItemPromotion_Apply(t *testing.T) {
	type fields struct {
		purchasedItemSku item.Sku
		FreeItemSku      item.Sku
		FreeItemPrice    int64
	}
	type args struct {
		getPurchasedHandler GetPurchasedItemHandler
		addPromoHandler     AddPromoItemToCartHandler
		AddDiscountHandler  AddDiscountToCartHandler
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		//TODO: add tests for free items,
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			fIP := FreeItemPromotion{
				PurchasedItemSku: tt.fields.purchasedItemSku,
				FreeItemSku:      tt.fields.FreeItemSku,
				FreeItemPrice:    tt.fields.FreeItemPrice,
			}
			if err := fIP.Apply(tt.args.getPurchasedHandler, tt.args.addPromoHandler, tt.args.AddDiscountHandler); (err != nil) != tt.wantErr {
				t.Errorf("Apply() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
