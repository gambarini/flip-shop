package promotion

import (
	"github.com/gambarini/flip-shop/internal/model/item"
	"github.com/gambarini/flip-shop/utils"
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

	if iQF.PurchasedQty > 0 && itemPurchased.Qty >= iQF.PurchasedQty {
		freeCount := itemPurchased.Qty / iQF.PurchasedQty
		discount = utils.SaturatingMulInt64Int(itemPurchased.Price, freeCount)
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
