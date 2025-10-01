package route

import (
	"github.com/gambarini/flip-shop/internal/model/promotion"
	"github.com/gambarini/flip-shop/internal/repo"
	"github.com/gambarini/flip-shop/utils"
)

// SetRoutes registers all HTTP routes for the application on the provided AppServer.
// It wires handlers with the necessary repositories and promotions.
func SetRoutes(srv *utils.AppServer, itemRepo repo.IItemRepository, cartRepo repo.ICartRepository, promotions []promotion.Promotion) error {

	// Items endpoints
	if err := srv.AddRoute("/items", "POST", postItem(srv, itemRepo)); err != nil {
		return err
	}
	if err := srv.AddRoute("/items/{sku}", "PUT", putItem(srv, itemRepo)); err != nil {
		return err
	}
	if err := srv.AddRoute("/items/{sku}/price", "PUT", putItemPrice(srv, itemRepo)); err != nil {
		return err
	}
	if err := srv.AddRoute("/items/{sku}", "GET", getItem(srv, itemRepo)); err != nil {
		return err
	}

	if err := srv.AddRoute("/cart", "POST", postCart(srv, cartRepo)); err != nil {
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
	if err := srv.AddRoute("/health", "GET", health(srv)); err != nil {
		return err
	}

	return nil
}
