package promotion

import (
	"errors"

	"github.com/gambarini/flip-shop/internal/model/item"
	"github.com/gambarini/flip-shop/utils"
)

type (
	// ItemQtyPriceDiscountPercentagePromotion
	// Describes a promotion where purchasing a qty of
	// an item gives a percentage discount on these items
	ItemQtyPriceDiscountPercentagePromotion struct {
		Name               string
		PurchasedItemSku   item.Sku
		PurchasedQty       int
		PercentageDiscount float32
	}
)

func (iQD ItemQtyPriceDiscountPercentagePromotion) Validate() error {
	if iQD.PurchasedQty <= 0 {
		return errors.New("ItemQtyPriceDiscountPercentagePromotion: PurchasedQty must be > 0")
	}
	if iQD.PercentageDiscount <= 0 || iQD.PercentageDiscount > 1 {
		return errors.New("ItemQtyPriceDiscountPercentagePromotion: PercentageDiscount must be in (0, 1]")
	}
	return nil
}

func (iQD ItemQtyPriceDiscountPercentagePromotion) Apply(ctx PromotionContext) (err error) {

	itemPurchased, ok := ctx.GetPurchased(iQD.PurchasedItemSku)

	if !ok {
		return nil
	}

	var discount int64

	if itemPurchased.Qty >= iQD.PurchasedQty {
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

	applied, err := ctx.AddDiscount(iQD.PurchasedItemSku, discount)
	if err != nil {
		return err
	}
	if ctx.AddApplied != nil && applied > 0 {
		ctx.AddApplied(iQD.Name, applied)
	}
	return nil
}
