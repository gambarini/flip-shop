package route

import (
	"encoding/json"
	"fmt"
	"net/http"

	"github.com/gambarini/flip-shop/internal/model/cart"
	"github.com/gambarini/flip-shop/internal/model/item"
	"github.com/gambarini/flip-shop/internal/repo"
	"github.com/gambarini/flip-shop/utils"
)

type (
	PurchaseItemPayload struct {
		Sku string `json:"sku"`
		Qty int    `json:"qty"`
	}
)

func purchase(srv *utils.AppServer, cartRepo repo.ICartRepository, itemRepo repo.IItemRepository) http.HandlerFunc {

	return func(response http.ResponseWriter, request *http.Request) {

		cartID := srv.Vars(request)["cartID"]

		currcart, err := cartRepo.FindCartByID(cartID)

		if err != nil {
			if err == repo.ErrCartNotFound {
				srv.ResponseErrorNotfound(response, err)
				return
			}
			srv.ResponseErrorServerErr(response, fmt.Errorf("error finding cart: %w", err))
			return
		}

		var rPayload PurchaseItemPayload
		dec := json.NewDecoder(request.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&rPayload); err != nil {
			srv.ResponseErrorEntityUnproc(response, fmt.Errorf("invalid JSON payload: %w", err))
			return
		}
		// Basic validation
		if rPayload.Sku == "" {
			srv.ResponseErrorEntityUnproc(response, fmt.Errorf("sku must be provided"))
			return
		}
		if rPayload.Qty <= 0 {
			srv.ResponseErrorEntityUnproc(response, cart.ErrItemQtyAddedInvalid)
			return
		}

		err = cartRepo.WithTx(func(tx utils.Tx) error {

			item, err := itemRepo.FindItemBySku(tx, item.Sku(rPayload.Sku))

			if err != nil {
				return err
			}

			err = item.ReserveItem(rPayload.Qty)

			if err != nil {
				return err
			}

			err = currcart.PurchaseItem(item, rPayload.Qty)

			if err != nil {
				return err
			}

			if err := cartRepo.Store(tx, currcart); err != nil {
				return err
			}

			if err := itemRepo.Store(tx, item); err != nil {
				return err
			}

			return nil
		})

		switch {
		case err == repo.ErrItemNotFound:
			srv.ResponseErrorEntityUnproc(response, err)
			return
		case err == item.ErrItemNotAvailableReservation:
			srv.ResponseErrorEntityUnproc(response, err)
			return
		case err == cart.ErrItemQtyAddedInvalid:
			srv.ResponseErrorEntityUnproc(response, err)
			return
		case err != nil:
			srv.ResponseErrorServerErr(response, fmt.Errorf("error storing Cart: %w", err))
			return
		}

		srv.RespondJSON(response, http.StatusOK, currcart)

	}
}
