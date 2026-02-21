package route

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"sort"

	"github.com/gambarini/flip-shop/internal/model/item"
	"github.com/gambarini/flip-shop/internal/repo"
	"github.com/gambarini/flip-shop/utils"
)

type (
	// AddItemPayload represents the request body to add a new item to inventory
	AddItemPayload struct {
		Sku   string `json:"sku"`
		Name  string `json:"name"`
		Price int64  `json:"price"`
		Qty   int    `json:"qty"`
	}
	// UpdateItemQtyPayload represents the request body to add quantity to an existing item
	UpdateItemQtyPayload struct {
		Qty int `json:"qty"`
	}
	// UpdateItemPricePayload represents the request body to adjust price of an existing item
	UpdateItemPricePayload struct {
		Price int64 `json:"price"`
	}
)

// listItems returns all items in inventory (unordered by default); for stable outputs we sort by SKU.
func listItems(srv *utils.AppServer, itemRepo repo.IItemRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		items, err := itemRepo.ListItems()
		if err != nil {
			srv.ResponseErrorServerErr(w, fmt.Errorf("error listing items: %w", err))
			return
		}
		// Sort by SKU for deterministic responses in tests/clients
		sort.Slice(items, func(i, j int) bool { return string(items[i].Sku) < string(items[j].Sku) })
		srv.RespondJSON(w, http.StatusOK, items)
	}
}

// getItem returns an item by sku
func getItem(srv *utils.AppServer, itemRepo repo.IItemRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sku := srv.Vars(r)["sku"]
		if sku == "" {
			srv.ResponseErrorEntityUnproc(w, fmt.Errorf("sku must be provided"))
			return
		}

		var found item.Item
		err := itemRepo.WithTx(func(tx utils.Tx) error {
			it, err := itemRepo.FindItemBySku(tx, item.Sku(sku))
			if err != nil {
				return err
			}
			found = it
			return nil
		})

		switch {
		case errors.Is(err, repo.ErrItemNotFound):
			srv.ResponseErrorNotfound(w, err)
			return
		case err != nil:
			srv.ResponseErrorServerErr(w, fmt.Errorf("error finding item: %w", err))
			return
		}

		srv.RespondJSON(w, http.StatusOK, found)
	}
}

// postItem creates a new item in inventory. If the SKU already exists, returns 422.
func postItem(srv *utils.AppServer, itemRepo repo.IItemRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		var payload AddItemPayload
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&payload); err != nil {
			srv.ResponseErrorEntityUnproc(w, fmt.Errorf("invalid JSON payload: %w", err))
			return
		}

		// basic validations
		if payload.Sku == "" {
			srv.ResponseErrorEntityUnproc(w, fmt.Errorf("sku must be provided"))
			return
		}
		if payload.Price < 0 {
			srv.ResponseErrorEntityUnproc(w, fmt.Errorf("price must be >= 0"))
			return
		}
		if payload.Qty < 0 {
			srv.ResponseErrorEntityUnproc(w, fmt.Errorf("qty must be >= 0"))
			return
		}

		var it item.Item
		if err := itemRepo.WithTx(func(tx utils.Tx) error {
			// Check if item already exists
			_, err := itemRepo.FindItemBySku(tx, item.Sku(payload.Sku))
			if err == nil {
				return fmt.Errorf("item with sku %s already exists", payload.Sku)
			}
			if !errors.Is(err, repo.ErrItemNotFound) {
				return err
			}
			// Create new item using constructor
			it = item.NewItem(item.Sku(payload.Sku), payload.Name, payload.Price, payload.Qty)
			return itemRepo.Store(tx, it)
		}); err != nil {
			// Map already exists to 422
			if err.Error() == fmt.Sprintf("item with sku %s already exists", payload.Sku) {
				srv.ResponseErrorEntityUnproc(w, err)
				return
			}
			srv.ResponseErrorServerErr(w, fmt.Errorf("error storing item: %w", err))
			return
		}

		srv.RespondJSON(w, http.StatusCreated, it)
	}
}

// putItem adds quantity to an existing item identified by path SKU.
func putItem(srv *utils.AppServer, itemRepo repo.IItemRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sku := srv.Vars(r)["sku"]
		if sku == "" {
			srv.ResponseErrorEntityUnproc(w, fmt.Errorf("sku must be provided"))
			return
		}

		var payload UpdateItemQtyPayload
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&payload); err != nil {
			srv.ResponseErrorEntityUnproc(w, fmt.Errorf("invalid JSON payload: %w", err))
			return
		}
		if payload.Qty < 0 {
			srv.ResponseErrorEntityUnproc(w, fmt.Errorf("qty must be >= 0"))
			return
		}

		var it item.Item
		if err := itemRepo.WithTx(func(tx utils.Tx) error {
			found, err := itemRepo.FindItemBySku(tx, item.Sku(sku))
			if err != nil {
				return err
			}
			found.Restock(payload.Qty)
			it = found
			return itemRepo.Store(tx, it)
		}); err != nil {
			switch {
			case errors.Is(err, repo.ErrItemNotFound):
				srv.ResponseErrorNotfound(w, err)
				return
			default:
				srv.ResponseErrorServerErr(w, fmt.Errorf("error updating item qty: %w", err))
				return
			}
		}

		srv.RespondJSON(w, http.StatusOK, it)
	}
}

// putItemPrice adjusts the price of an existing item identified by path SKU.
func putItemPrice(srv *utils.AppServer, itemRepo repo.IItemRepository) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		sku := srv.Vars(r)["sku"]
		if sku == "" {
			srv.ResponseErrorEntityUnproc(w, fmt.Errorf("sku must be provided"))
			return
		}

		var payload UpdateItemPricePayload
		dec := json.NewDecoder(r.Body)
		dec.DisallowUnknownFields()
		if err := dec.Decode(&payload); err != nil {
			srv.ResponseErrorEntityUnproc(w, fmt.Errorf("invalid JSON payload: %w", err))
			return
		}
		if payload.Price < 0 {
			srv.ResponseErrorEntityUnproc(w, fmt.Errorf("price must be >= 0"))
			return
		}

		var it item.Item
		if err := itemRepo.WithTx(func(tx utils.Tx) error {
			found, err := itemRepo.FindItemBySku(tx, item.Sku(sku))
			if err != nil {
				return err
			}
			found.AdjustPrice(payload.Price)
			it = found
			return itemRepo.Store(tx, it)
		}); err != nil {
			switch {
			case errors.Is(err, repo.ErrItemNotFound):
				srv.ResponseErrorNotfound(w, err)
				return
			default:
				srv.ResponseErrorServerErr(w, fmt.Errorf("error updating item price: %w", err))
				return
			}
		}

		srv.RespondJSON(w, http.StatusOK, it)
	}
}
