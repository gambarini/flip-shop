package promotion

import (
	"errors"

	"github.com/gambarini/flip-shop/utils"
)

type (
	// CartSpendThresholdPromotion
	// Applies a percentage discount to every item when the pre-discount
	// cart total meets or exceeds the configured threshold.
	CartSpendThresholdPromotion struct {
		Name               string
		ThresholdAmount    int64
		PercentageDiscount float32
	}
)

func (p CartSpendThresholdPromotion) Validate() error {
	if p.ThresholdAmount <= 0 {
		return errors.New("CartSpendThresholdPromotion: ThresholdAmount must be > 0")
	}
	if p.PercentageDiscount <= 0 || p.PercentageDiscount > 1 {
		return errors.New("CartSpendThresholdPromotion: PercentageDiscount must be in (0, 1]")
	}
	return nil
}

func (p CartSpendThresholdPromotion) Apply(ctx PromotionContext) error {
	if ctx.GetCartTotal() < p.ThresholdAmount {
		return nil
	}

	per := p.PercentageDiscount
	if per < 0 {
		per = 0
	}
	if per > 1 {
		per = 1
	}

	var totalDiscount int64
	for _, pu := range ctx.GetAllPurchased() {
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
