package promotion

import (
	"github.com/gambarini/flip-shop/internal/model/item"
)

type (
	// ItemQtyPriceDiscountPromotion
	// Describes a promotion where purchasing a qty of
	// an item gives some of these items for free
	ItemQtyPriceDiscountPromotion struct {
		PurchasedItemSku item.Sku
		PurchasedQty     int
	}
)

func (fIP ItemQtyPriceDiscountPromotion) Apply(getPurchasedHandler GetPurchasedItemHandler, addPromoHandler AddPromoItemToCartHandler, AddDiscountHandler AddDiscountToCartHandler) (err error) {

	itemPurchased, ok := getPurchasedHandler(fIP.PurchasedItemSku)

	if !ok {
		return nil
	}

	var discount int64

	for i := 1; i <= itemPurchased.Qty; i++ {

		if i%fIP.PurchasedQty == 0 {

			discount += itemPurchased.Price
		}
	}

	if discount == 0 {
		return
	}

	err = AddDiscountHandler(fIP.PurchasedItemSku, discount)

	if err != nil {
		return err
	}

	return nil
}



