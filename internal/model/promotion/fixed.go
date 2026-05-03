package promotion

import (
	"errors"

	"github.com/gambarini/flip-shop/internal/model/item"
)

type (
	// ItemQtyFixedDiscountPromotion
	// Describes a promotion where purchasing a minimum qty of
	// an item applies a fixed cent amount off the line total
	ItemQtyFixedDiscountPromotion struct {
		Name             string
		PurchasedItemSku item.Sku
		PurchasedQty     int
		FixedDiscount    int64
	}
)

func (p ItemQtyFixedDiscountPromotion) Validate() error {
	if p.PurchasedQty <= 0 {
		return errors.New("ItemQtyFixedDiscountPromotion: PurchasedQty must be > 0")
	}
	if p.FixedDiscount <= 0 {
		return errors.New("ItemQtyFixedDiscountPromotion: FixedDiscount must be > 0")
	}
	return nil
}

func (p ItemQtyFixedDiscountPromotion) Apply(ctx PromotionContext) error {
	purchased, ok := ctx.GetPurchased(p.PurchasedItemSku)
	if !ok || purchased.Qty < p.PurchasedQty {
		return nil
	}
	applied, err := ctx.AddDiscount(p.PurchasedItemSku, p.FixedDiscount)
	if err != nil {
		return err
	}
	if ctx.AddApplied != nil && applied > 0 {
		ctx.AddApplied(p.Name, applied)
	}
	return nil
}
