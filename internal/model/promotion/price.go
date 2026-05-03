package promotion

import (
	"errors"

	"github.com/gambarini/flip-shop/internal/model/item"
	"github.com/gambarini/flip-shop/utils"
)

type (
	// ItemQtyPriceFreePromotion
	// Describes a promotion where purchasing a qty of
	// an item gives some of these items for free
	ItemQtyPriceFreePromotion struct {
		Name             string
		PurchasedItemSku item.Sku
		PurchasedQty     int
	}
)

func (iQF ItemQtyPriceFreePromotion) Validate() error {
	if iQF.PurchasedQty <= 0 {
		return errors.New("ItemQtyPriceFreePromotion: PurchasedQty must be > 0")
	}
	return nil
}

func (iQF ItemQtyPriceFreePromotion) Apply(ctx PromotionContext) (err error) {

	itemPurchased, ok := ctx.GetPurchased(iQF.PurchasedItemSku)

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

	applied, err := ctx.AddDiscount(iQF.PurchasedItemSku, discount)
	if err != nil {
		return err
	}
	if ctx.AddApplied != nil && applied > 0 {
		ctx.AddApplied(iQF.Name, applied)
	}
	return nil
}
