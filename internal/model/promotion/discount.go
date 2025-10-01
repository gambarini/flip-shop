package promotion

import (
	"github.com/gambarini/flip-shop/internal/model/item"
	"github.com/gambarini/flip-shop/utils"
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

func (iQD ItemQtyPriceDiscountPercentagePromotion) Apply(getPurchasedHandler GetPurchasedItemHandler, addPromoHandler AddPromoItemToCartHandler, AddDiscountHandler AddDiscountToCartHandler) (err error) {

	itemPurchased, ok := getPurchasedHandler(iQD.PurchasedItemSku)

	if !ok {
		return nil
	}

	var discount int64

	if itemPurchased.Qty > iQD.PurchasedQty {
		total := utils.SaturatingMulInt64Int(itemPurchased.Price, itemPurchased.Qty)
		// convert to the whole currency then back to cents to apply percentage deterministically
		total = total / 100
		per := iQD.PercentageDiscount
		if per < 0 {
			per = 0
		}
		if per > 1 {
			per = 1
		}
		discount = int64((float32(total) * per) * 100)

	}

	if discount == 0 {
		return
	}

	err = AddDiscountHandler(iQD.PurchasedItemSku, discount)

	if err != nil {
		return err
	}

	return nil
}
