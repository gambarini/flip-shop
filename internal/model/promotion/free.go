package promotion

import (
	"github.com/gambarini/flip-shop/internal/model/item"
	"github.com/gambarini/flip-shop/utils"
)

type (

	// FreeItemPromotion
	// Describes a promotion where purchasing one item
	// gives another one free
	FreeItemPromotion struct {
		PurchasedItemSku item.Sku
		FreeItemSku      item.Sku
		FreeItemPrice    int64
	}
)

func (fIP FreeItemPromotion) Apply(getPurchasedHandler GetPurchasedItemHandler, addPromoHandler AddPromoItemToCartHandler, AddDiscountHandler AddDiscountToCartHandler) (err error) {

	itemPurchased, ok := getPurchasedHandler(fIP.PurchasedItemSku)

	if !ok {
		return nil
	}

	err = addPromoHandler(fIP.FreeItemSku, itemPurchased.Qty)

	if err != nil {
		return err
	}

	discount := utils.SaturatingMulInt64Int(fIP.FreeItemPrice, itemPurchased.Qty)

	err = AddDiscountHandler(fIP.FreeItemSku, discount)

	if err != nil {
		return err
	}

	return nil
}
