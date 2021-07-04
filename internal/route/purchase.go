package route

import (
	"encoding/json"
	"fmt"
	"github.com/gambarini/flip-shop/internal/model/cart"
	"github.com/gambarini/flip-shop/internal/model/item"
	"github.com/gambarini/flip-shop/internal/repo"
	"github.com/gambarini/flip-shop/utils"
	"io/ioutil"
	"log"
	"net/http"
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
			log.Printf("Error finding cart, %s", err)
			response.WriteHeader(http.StatusNotFound)
			return
		}

		bodyJSON, err := ioutil.ReadAll(request.Body)

		var rPayload PurchaseItemPayload

		err = json.Unmarshal(bodyJSON, &rPayload)

		if err != nil {
			log.Printf("Error reading payload, %s", err)
			response.WriteHeader(http.StatusInternalServerError)
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

		switch  {
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
			srv.ResponseErrorServerErr(response, fmt.Errorf("error storing Cart, %s", err))
			return
		}

		cartJSON, err := json.Marshal(currcart)

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
