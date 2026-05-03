package route

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/gambarini/flip-shop/internal/model/cart"
	"github.com/gambarini/flip-shop/internal/model/item"
	"github.com/gambarini/flip-shop/internal/model/promotion"
	"github.com/gambarini/flip-shop/internal/repo"
	"github.com/gambarini/flip-shop/utils"
)

func submit(srv *utils.AppServer, cartRepo repo.ICartRepository, itemRepo repo.IItemRepository, promotions []promotion.Promotion) http.HandlerFunc {

	return func(response http.ResponseWriter, request *http.Request) {

		cartID := srv.Vars(request)["cartID"]

		var submitCart cart.Cart

		err := cartRepo.WithTx(func(tx utils.Tx) error {

			var innerErr error
			submitCart, innerErr = cartRepo.FindCartByIDInTx(tx, cartID)
			if innerErr != nil {
				return innerErr
			}

			var appliedPromos []cart.AppliedPromotion
			ctx := promotion.PromotionContext{
				GetPurchased:    GetPurchasedItemForPromotion(submitCart),
				AddPromo:        AddPurchaseToCartForPromotion(tx, itemRepo, submitCart),
				AddDiscount:     AddDiscountToPurchaseForPromotion(submitCart),
				GetAllPurchased: GetAllPurchasedItemsForPromotion(submitCart),
				GetCartTotal:    GetCartTotalForPromotion(submitCart),
				AddApplied: func(name string, discount int64) {
					appliedPromos = append(appliedPromos, cart.AppliedPromotion{Name: name, Discount: discount})
				},
			}

			for _, p := range promotions {
				if err := p.Apply(ctx); err != nil {
					return err
				}
			}
			submitCart.AppliedPromotions = appliedPromos

			for _, pu := range submitCart.Purchases {

				i, err := itemRepo.FindItemBySku(tx, pu.Sku)

				if err != nil {
					return err
				}

				if err = i.RemoveItem(pu.Qty); err != nil {
					return err
				}

				if err = itemRepo.Store(tx, i); err != nil {
					return err
				}

			}

			err := submitCart.SubmitCart()

			if err != nil {
				return err
			}

			if err := cartRepo.Store(tx, submitCart); err != nil {
				return err
			}

			return nil
		})

		switch {
		case errors.Is(err, repo.ErrCartNotFound):
			srv.ResponseErrorNotfound(response, err)
			return
		case errors.Is(err, repo.ErrItemNotFound):
			srv.ResponseErrorEntityUnproc(response, err)
			return
		case errors.Is(err, item.ErrInvalidReleaseQuantity):
			srv.ResponseErrorEntityUnproc(response, err)
			return
		case errors.Is(err, item.ErrItemNotAvailableReservation):
			srv.ResponseErrorEntityUnproc(response, err)
			return
		case errors.Is(err, item.ErrInvalidRemoveQuantity):
			srv.ResponseErrorEntityUnproc(response, err)
			return
		case errors.Is(err, cart.ErrItemQtyAddedInvalid):
			srv.ResponseErrorEntityUnproc(response, err)
			return
		case errors.Is(err, cart.ErrItemNotInCart):
			srv.ResponseErrorEntityUnproc(response, err)
			return
		case err != nil:
			srv.ResponseErrorServerErr(response, fmt.Errorf("error storing Cart: %w", err))
			return
		}

		srv.RespondJSON(response, http.StatusOK, submitCart)

	}
}

func AddDiscountToPurchaseForPromotion(cart cart.Cart) func(sku item.Sku, discount int64) (int64, error) {
	return func(sku item.Sku, discount int64) (int64, error) {
		return cart.DiscountPurchase(sku, discount)
	}
}

// AddPurchaseToCartForPromotion reserves the promotional items before adding them to the cart to ensure
// inventory invariants are maintained. Returns the total price charged for the added items so the caller
// can apply an exact matching discount. If reservation fails (insufficient availability), the promotion
// application aborts and no cart state is mutated, as the call happens within the transaction boundary.
func AddPurchaseToCartForPromotion(tx utils.Tx, itemRepo repo.IItemRepository, cart cart.Cart) func(sku item.Sku, qty int) (int64, error) {
	return func(sku item.Sku, qty int) (int64, error) {

		i, err := itemRepo.FindItemBySku(tx, sku)

		if err != nil {
			return 0, err
		}

		if err = i.ReserveItem(qty); err != nil {
			return 0, err
		}

		if err := cart.PurchaseItem(i, qty); err != nil {
			return 0, err
		}

		if err = itemRepo.Store(tx, i); err != nil {
			return 0, err
		}

		return utils.SaturatingMulInt64Int(i.Price, qty), nil
	}
}

func GetPurchasedItemForPromotion(cart cart.Cart) func(sku item.Sku) (promotion.PurchasedItem, bool) {
	return func(sku item.Sku) (promotion.PurchasedItem, bool) {
		pu, ok := cart.Purchases[sku]

		if !ok {
			return promotion.PurchasedItem{}, false
		}

		return promotion.PurchasedItem{
			Sku:      pu.Sku,
			Name:     pu.Name,
			Price:    pu.Price,
			Qty:      pu.Qty,
			Discount: pu.Discount,
		}, true

	}
}

func GetAllPurchasedItemsForPromotion(c cart.Cart) func() []promotion.PurchasedItem {
	return func() []promotion.PurchasedItem {
		items := make([]promotion.PurchasedItem, 0, len(c.Purchases))
		for _, pu := range c.Purchases {
			items = append(items, promotion.PurchasedItem{
				Sku:      pu.Sku,
				Name:     pu.Name,
				Price:    pu.Price,
				Qty:      pu.Qty,
				Discount: pu.Discount,
			})
		}
		return items
	}
}

// GetCartTotalForPromotion returns the pre-discount line total across all purchases.
func GetCartTotalForPromotion(c cart.Cart) func() int64 {
	return func() int64 {
		var total int64
		for _, pu := range c.Purchases {
			total = utils.SaturatingAddInt64(total, utils.SaturatingMulInt64Int(pu.Price, pu.Qty))
		}
		return total
	}
}
