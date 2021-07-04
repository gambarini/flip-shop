package promotion

import (
	"github.com/gambarini/flip-shop/internal/model/item"
)

type (
	// ItemQtyPriceFreePromotion
	// Describes a promotion where purchasing a qty of
	// an item gives some of these items for free
	ItemQtyPriceFreePromotion struct {
		PurchasedItemSku item.Sku
		PurchasedQty     int
	}
)

func (iQF ItemQtyPriceFreePromotion) Apply(getPurchasedHandler GetPurchasedItemHandler, addPromoHandler AddPromoItemToCartHandler, AddDiscountHandler AddDiscountToCartHandler) (err error) {

	itemPurchased, ok := getPurchasedHandler(iQF.PurchasedItemSku)

	if !ok {
		return nil
	}

	var discount int64

	for i := 1; i <= itemPurchased.Qty; i++ {

		if i%iQF.PurchasedQty == 0 {

			discount += itemPurchased.Price
		}
	}

	if discount == 0 {
		return
	}

	err = AddDiscountHandler(iQF.PurchasedItemSku, discount)

	if err != nil {
		return err
	}

	return nil
}



