package promotion

import (
	"errors"

	"github.com/gambarini/flip-shop/internal/model/item"
	"github.com/gambarini/flip-shop/utils"
)

type (
	// BundleDiscountPromotion
	// Applies a percentage discount to every item in the bundle
	// when all specified SKUs are present in the cart.
	BundleDiscountPromotion struct {
		Name               string
		ItemSkus           []item.Sku
		PercentageDiscount float32
	}
)

func (p BundleDiscountPromotion) Validate() error {
	if len(p.ItemSkus) < 2 {
		return errors.New("BundleDiscountPromotion: ItemSkus must contain at least 2 SKUs")
	}
	if p.PercentageDiscount <= 0 || p.PercentageDiscount > 1 {
		return errors.New("BundleDiscountPromotion: PercentageDiscount must be in (0, 1]")
	}
	return nil
}

func (p BundleDiscountPromotion) Apply(ctx PromotionContext) error {
	// Verify every bundle SKU is in the cart before applying any discount
	purchased := make([]PurchasedItem, 0, len(p.ItemSkus))
	for _, sku := range p.ItemSkus {
		pu, ok := ctx.GetPurchased(sku)
		if !ok {
			return nil
		}
		purchased = append(purchased, pu)
	}

	per := p.PercentageDiscount
	if per < 0 {
		per = 0
	}
	if per > 1 {
		per = 1
	}

	var totalDiscount int64
	for _, pu := range purchased {
		total := utils.SaturatingMulInt64Int(pu.Price, pu.Qty)
		total = total / 100
		discount := int64((float32(total) * per) * 100)
		if discount == 0 {
			continue
		}
		applied, err := ctx.AddDiscount(pu.Sku, discount)
		if err != nil {
			return err
		}
		totalDiscount += applied
	}
	if ctx.AddApplied != nil && totalDiscount > 0 {
		ctx.AddApplied(p.Name, totalDiscount)
	}
	return nil
}
