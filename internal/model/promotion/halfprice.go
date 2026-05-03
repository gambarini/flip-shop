package promotion

import (
	"errors"

	"github.com/gambarini/flip-shop/internal/model/item"
	"github.com/gambarini/flip-shop/utils"
)

type (
	// ItemQtyHalfPricePromotion
	// Describes a promotion where every Nth unit of an item
	// is charged at half price (e.g. buy 2 pay for 1.5)
	ItemQtyHalfPricePromotion struct {
		Name             string
		PurchasedItemSku item.Sku
		PurchasedQty     int
	}
)

func (p ItemQtyHalfPricePromotion) Validate() error {
	if p.PurchasedQty <= 0 {
		return errors.New("ItemQtyHalfPricePromotion: PurchasedQty must be > 0")
	}
	return nil
}

func (p ItemQtyHalfPricePromotion) Apply(ctx PromotionContext) error {
	purchased, ok := ctx.GetPurchased(p.PurchasedItemSku)
	if !ok || purchased.Qty < p.PurchasedQty {
		return nil
	}
	// One unit per group of PurchasedQty is discounted by 50%
	discountedUnits := purchased.Qty / p.PurchasedQty
	discount := utils.SaturatingMulInt64Int(purchased.Price/2, discountedUnits)
	if discount == 0 {
		return nil
	}
	applied, err := ctx.AddDiscount(p.PurchasedItemSku, discount)
	if err != nil {
		return err
	}
	if ctx.AddApplied != nil && applied > 0 {
		ctx.AddApplied(p.Name, applied)
	}
	return nil
}
