package promotion

import (
	"github.com/gambarini/flip-shop/internal/model/item"
)

type (
	// ItemQtyPriceDiscountPercentagePromotion
	// Describes a promotion where purchasing a qty of
	// an item gives a percentage discount on these items
	ItemQtyPriceDiscountPercentagePromotion struct {
		PurchasedItemSku   item.Sku
		PurchasedQty       int
		PercentageDiscount float32
	}
)

func (fIP ItemQtyPriceDiscountPercentagePromotion) Apply(getPurchasedHandler GetPurchasedItemHandler, addPromoHandler AddPromoItemToCartHandler, AddDiscountHandler AddDiscountToCartHandler) (err error) {

	itemPurchased, ok := getPurchasedHandler(fIP.PurchasedItemSku)

	if !ok {
		return nil
	}

	var discount int64

	if itemPurchased.Qty > fIP.PurchasedQty {
		total := (itemPurchased.Price * int64(itemPurchased.Qty)) / 100

		discount = int64((float32(total) * fIP.PercentageDiscount) * 100)

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
