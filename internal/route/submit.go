package route

import (
	"encoding/json"
	"fmt"
	"github.com/gambarini/flip-shop/internal/model/cart"
	"github.com/gambarini/flip-shop/internal/model/item"
	"github.com/gambarini/flip-shop/internal/model/promotion"
	"github.com/gambarini/flip-shop/internal/repo"
	"github.com/gambarini/flip-shop/utils"
	"log"
	"net/http"
)

func submit(srv *utils.AppServer, cartRepo repo.ICartRepository, itemRepo repo.IItemRepository, promotions []promotion.Promotion) http.HandlerFunc {

	return func(response http.ResponseWriter, request *http.Request) {

		cartID := srv.Vars(request)["cartID"]

		submitCart, err := cartRepo.FindCartByID(cartID)

		if err != nil {
			log.Printf("Error finding Cart, %s", err)
			response.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = cartRepo.WithTx(func(tx utils.Tx) error {

			for _, p := range promotions {
				err = p.Apply(
					GetPurchasedItemForPromotion(submitCart),
					AddPurchaseToCartForPromotion(tx, itemRepo, submitCart),
					AddDiscountToPurchaseForPromotion(submitCart))
			}

			if err != nil {
				return err
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



		switch  {
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
			srv.ResponseErrorServerErr(response, fmt.Errorf("error storing Cart, %s", err))
			return
		}

		cartJSON, err := json.Marshal(submitCart)

		if err != nil {
			log.Printf("Error serializing response, %s", err)
			response.WriteHeader(http.StatusInternalServerError)
			return
		}

		response.Header().Set("Content-Type", "application/json")
		response.WriteHeader(http.StatusOK)

		_, err = response.Write(cartJSON)

		if err != nil {
			log.Printf("Error serializing response, %s", err)
			response.WriteHeader(http.StatusInternalServerError)
			return
		}

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
