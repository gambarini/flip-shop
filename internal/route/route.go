package route

import (
	"fmt"
	"github.com/gambarini/flip-shop/internal/model/promotion"
	"github.com/gambarini/flip-shop/internal/repo"
	"github.com/gambarini/flip-shop/utils"
	"net/http"
)

func SetRoutes(srv *utils.AppServer, itemRepo repo.IItemRepository, cartRepo repo.ICartRepository, promotions []promotion.Promotion) error {

	if err := srv.AddRoute("/cart", "POST", postCart(cartRepo)); err != nil {
		return err
	}
	if err := srv.AddRoute("/cart/{cartID}/purchase", "PUT", purchase(srv, cartRepo, itemRepo)); err != nil {
		return err
	}
	if err := srv.AddRoute("/cart/{cartID}/purchase", "DELETE", remove(srv, cartRepo, itemRepo)); err != nil {
		return err
	}
	if err := srv.AddRoute("/cart/{cartID}/status/submitted", "PUT", submit(srv, cartRepo, itemRepo, promotions)); err != nil {
		return err
	}

	return nil
}

func ResponseErrorEntityUnproc(response http.ResponseWriter, err error) {
	response.WriteHeader(http.StatusUnprocessableEntity)
	_,_ = response.Write([]byte(fmt.Sprintf("{\"error\":\"%s\"}", err)))
}

func ResponseErrorServerErr(response http.ResponseWriter, err error) {
	response.WriteHeader(http.StatusInternalServerError)
}

func ResponseErrorNotfound(response http.ResponseWriter, err error) {
	response.WriteHeader(http.StatusNotFound)
}
