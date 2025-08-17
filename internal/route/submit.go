package route

import (
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

		submitCart, err := cartRepo.FindCartByID(cartID)

		if err != nil {
			if err == repo.ErrCartNotFound {
				srv.ResponseErrorNotfound(response, err)
				return
			}
			srv.ResponseErrorServerErr(response, fmt.Errorf("error finding cart: %w", err))
			return
		}

		err = cartRepo.WithTx(func(tx utils.Tx) error {

			for _, p := range promotions {
				if err := p.Apply(
					GetPurchasedItemForPromotion(submitCart),
					AddPurchaseToCartForPromotion(tx, itemRepo, submitCart),
					AddDiscountToPurchaseForPromotion(submitCart)); err != nil {
					return err
				}
			}

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

			err = submitCart.SubmitCart()

			if err != nil {
				return err
			}

			if err := cartRepo.Store(tx, submitCart); err != nil {
				return err
			}

			return nil
		})

		switch {
		case err == repo.ErrItemNotFound:
			srv.ResponseErrorEntityUnproc(response, err)
			return
		case err == item.ErrInvalidReleaseQuantity:
			srv.ResponseErrorEntityUnproc(response, err)
			return
		case err == item.ErrItemNotAvailableReservation:
			srv.ResponseErrorEntityUnproc(response, err)
			return
		case err == item.ErrInvalidRemoveQuantity:
			srv.ResponseErrorEntityUnproc(response, err)
			return
		case err == cart.ErrItemQtyAddedInvalid:
			srv.ResponseErrorEntityUnproc(response, err)
			return
		case err == cart.ErrItemNotInCart:
			srv.ResponseErrorEntityUnproc(response, err)
			return
		case err != nil:
			srv.ResponseErrorServerErr(response, fmt.Errorf("error storing Cart: %w", err))
			return
		}

		srv.RespondJSON(response, http.StatusOK, submitCart)

	}
}

func AddDiscountToPurchaseForPromotion(cart cart.Cart) func(sku item.Sku, discount int64) error {
	return func(sku item.Sku, discount int64) error {

		if err := cart.DiscountPurchase(sku, discount); err != nil {
			return err
		}

		return nil
	}
}

// AddPurchaseToCartForPromotion reserves the promotional items before adding them to the cart to ensure
// inventory invariants are maintained. If reservation fails (insufficient availability), the promotion
// application aborts and no cart state is mutated, as the call happens within the transaction boundary.
func AddPurchaseToCartForPromotion(tx utils.Tx, itemRepo repo.IItemRepository, cart cart.Cart) func(sku item.Sku, qty int) error {
	return func(sku item.Sku, qty int) error {

		i, err := itemRepo.FindItemBySku(tx, sku)

		if err != nil {
			return err
		}

		if err = i.ReserveItem(qty); err != nil {
			return err
		}

		if err := cart.PurchaseItem(i, qty); err != nil {
			return err
		}

		if err = itemRepo.Store(tx, i); err != nil {
			return err
		}

		return nil
	}
}

func GetPurchasedItemForPromotion(cart cart.Cart) func(sku item.Sku) (promotion.PurchasedItem, bool) {
	return func(sku item.Sku) (promotion.PurchasedItem, bool) {
		pu, ok := cart.Purchases[sku]

		if !ok {
			return promotion.PurchasedItem{}, false
		}

		return promotion.PurchasedItem{
			Name:     pu.Name,
			Price:    pu.Price,
			Qty:      pu.Qty,
			Discount: pu.Discount,
		}, true

	}
}
