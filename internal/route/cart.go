package route

import (
	"encoding/json"
	"github.com/gambarini/flip-shop/internal/model/cart"
	"github.com/gambarini/flip-shop/internal/repo"
	"github.com/gambarini/flip-shop/utils"
	"log"
	"net/http"
)

func postCart(cartRepo repo.ICartRepository) http.HandlerFunc {

	return func(response http.ResponseWriter, request *http.Request) {

		newCart := cart.NewAvailableCart()

		err := cartRepo.WithTx(func(tx utils.Tx) error {

			if err := cartRepo.Store(tx, newCart); err != nil {
				return err
			}

			return nil
		})

		if err != nil {
			log.Printf("Error storing cart, %s", err)
			response.WriteHeader(http.StatusInternalServerError)
			return
		}

		cartJSON, err := json.Marshal(newCart)

		if err != nil {
			log.Printf("Error serializing response, %s", err)
			response.WriteHeader(http.StatusInternalServerError)
			return
		}

		response.Header().Set("Content-Type", "application/json")
		response.WriteHeader(http.StatusCreated)

		_, err = response.Write(cartJSON)

		if err != nil {
			log.Printf("Error serializing response, %s", err)
			response.WriteHeader(http.StatusInternalServerError)
			return
		}


	}
}
