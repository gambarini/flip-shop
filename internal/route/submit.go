package route

import (
	"encoding/json"
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

		cart, err := cartRepo.FindCartByID(cartID)

		if err != nil {
			log.Printf("Error finding cart, %s", err)
			response.WriteHeader(http.StatusInternalServerError)
			return
		}

		err = cartRepo.WithTx(func(tx utils.Tx) error {

			for _, p := range promotions {
				err = p.Apply(func(sku item.Sku) (promotion.PurchasedItem, bool) {
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

				}, func(sku item.Sku, qty int) error {

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

					return nil
				}, func(sku item.Sku, discount int64) error {

					if err := cart.DiscountPurchase(sku, discount); err != nil {
						return err
					}

					return nil
				})
			}

			if err != nil {
				return err
			}

			err = cart.SubmitCart()

			if err != nil {
				return err
			}

			if err := cartRepo.Store(tx, cart); err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			log.Printf("Error storing cart, %s", err)
			response.WriteHeader(http.StatusInternalServerError)
			return
		}

		cartJSON, err := json.Marshal(cart)

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
