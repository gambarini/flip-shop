package promotion

import (
	"errors"

	"github.com/gambarini/flip-shop/internal/model/item"
)

type (
	// FreeItemPromotion
	// Describes a promotion where purchasing one item
	// gives another one free
	FreeItemPromotion struct {
		Name             string
		PurchasedItemSku item.Sku
		FreeItemSku      item.Sku
	}
)

func (fIP FreeItemPromotion) Validate() error {
	if fIP.PurchasedItemSku == "" {
		return errors.New("FreeItemPromotion: PurchasedItemSku must not be empty")
	}
	if fIP.FreeItemSku == "" {
		return errors.New("FreeItemPromotion: FreeItemSku must not be empty")
	}
	return nil
}

func (fIP FreeItemPromotion) Apply(ctx PromotionContext) (err error) {

	itemPurchased, ok := ctx.GetPurchased(fIP.PurchasedItemSku)

	if !ok {
		return nil
	}

	charged, err := ctx.AddPromo(fIP.FreeItemSku, itemPurchased.Qty)

	if err != nil {
		if errors.Is(err, item.ErrItemNotAvailableReservation) {
			return nil
		}
		return err
	}

	applied, err := ctx.AddDiscount(fIP.FreeItemSku, charged)
	if err != nil {
		return err
	}
	if ctx.AddApplied != nil && applied > 0 {
		ctx.AddApplied(fIP.Name, applied)
	}
	return nil
}
