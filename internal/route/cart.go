package route

import (
	"net/http"

	"github.com/gambarini/flip-shop/internal/model/cart"
	"github.com/gambarini/flip-shop/internal/repo"
	"github.com/gambarini/flip-shop/utils"
)

func postCart(srv *utils.AppServer, cartRepo repo.ICartRepository) http.HandlerFunc {

	return func(response http.ResponseWriter, request *http.Request) {

		newCart := cart.NewAvailableCart()

		err := cartRepo.WithTx(func(tx utils.Tx) error {

			if err := cartRepo.Store(tx, newCart); err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			srv.Logger().Error("cart_store_error", utils.Fields{"error": err.Error()})
			response.Header().Set("Content-Type", "application/json")
			response.WriteHeader(http.StatusInternalServerError)
			return
		}

		srv.RespondJSON(response, http.StatusCreated, newCart)

	}
}
