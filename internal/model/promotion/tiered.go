package promotion

import (
	"errors"
	"sort"

	"github.com/gambarini/flip-shop/internal/model/item"
	"github.com/gambarini/flip-shop/utils"
)

type (
	// DiscountTier pairs a minimum quantity with a percentage discount.
	DiscountTier struct {
		MinQty             int
		PercentageDiscount float32
	}

	// ItemQtyTieredDiscountPromotion
	// Applies the highest matching tier's percentage discount to the line total.
	// Tiers are evaluated highest-MinQty-first; first match wins.
	ItemQtyTieredDiscountPromotion struct {
		Name             string
		PurchasedItemSku item.Sku
		Tiers            []DiscountTier
	}
)

func (p ItemQtyTieredDiscountPromotion) Validate() error {
	if len(p.Tiers) == 0 {
		return errors.New("ItemQtyTieredDiscountPromotion: Tiers must not be empty")
	}
	for _, t := range p.Tiers {
		if t.MinQty <= 0 {
			return errors.New("ItemQtyTieredDiscountPromotion: tier MinQty must be > 0")
		}
		if t.PercentageDiscount <= 0 || t.PercentageDiscount > 1 {
			return errors.New("ItemQtyTieredDiscountPromotion: tier PercentageDiscount must be in (0, 1]")
		}
	}
	return nil
}

func (p ItemQtyTieredDiscountPromotion) Apply(ctx PromotionContext) error {
	purchased, ok := ctx.GetPurchased(p.PurchasedItemSku)
	if !ok {
		return nil
	}

	// Sort descending so the first matching tier is the highest qualifying one
	tiers := make([]DiscountTier, len(p.Tiers))
	copy(tiers, p.Tiers)
	sort.Slice(tiers, func(i, j int) bool { return tiers[i].MinQty > tiers[j].MinQty })

	var pct float32
	for _, t := range tiers {
		if purchased.Qty >= t.MinQty {
			pct = t.PercentageDiscount
			break
		}
	}
	if pct == 0 {
		return nil
	}

	total := utils.SaturatingMulInt64Int(purchased.Price, purchased.Qty)
	total = total / 100
	discount := int64((float32(total) * pct) * 100)
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
