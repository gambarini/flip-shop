package promotion

import (
	"errors"
)

type (
	// CheapestItemFreePromotion
	// When the total cart quantity meets MinItemCount,
	// applies a full unit discount on the cheapest item by unit price.
	CheapestItemFreePromotion struct {
		Name         string
		MinItemCount int
	}
)

func (p CheapestItemFreePromotion) Validate() error {
	if p.MinItemCount < 2 {
		return errors.New("CheapestItemFreePromotion: MinItemCount must be >= 2")
	}
	return nil
}

func (p CheapestItemFreePromotion) Apply(ctx PromotionContext) error {
	items := ctx.GetAllPurchased()

	var totalQty int
	for _, pu := range items {
		totalQty += pu.Qty
	}
	if totalQty < p.MinItemCount {
		return nil
	}

	var cheapest *PurchasedItem
	for i := range items {
		if cheapest == nil || items[i].Price < cheapest.Price {
			cheapest = &items[i]
		}
	}
	if cheapest == nil || cheapest.Price == 0 {
		return nil
	}

	applied, err := ctx.AddDiscount(cheapest.Sku, cheapest.Price)
	if err != nil {
		return err
	}
	if ctx.AddApplied != nil && applied > 0 {
		ctx.AddApplied(p.Name, applied)
	}
	return nil
}
