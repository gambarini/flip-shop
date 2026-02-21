package mcp

import (
	"context"
	"encoding/json"
	"io"
	"log"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/gambarini/flip-shop/internal/model/item"
	"github.com/gambarini/flip-shop/internal/model/promotion"
	"github.com/gambarini/flip-shop/internal/repo"
	"github.com/gambarini/flip-shop/internal/route"
	"github.com/gambarini/flip-shop/utils"
	"github.com/gambarini/flip-shop/utils/memdb"
)

// Test_Integration_Create_Add_Submit spins up an in-memory flip-shop HTTP server
// using the real routes and runs a full happy path via the MCP server.
func Test_Integration_Create_Add_Submit(t *testing.T) {
	// Build flip-shop HTTP handler with real routes, repos, and seeded inventory
	memDb := memdb.NewMemoryKVDatabase()
	seedInventory := func(tx utils.Tx) error {
		tx.Write(repo.ItemStoreName, "120P90", item.Item{Sku: "120P90", Name: "Google Home", QtyAvailable: 10, Price: 4999})
		tx.Write(repo.ItemStoreName, "43N23P", item.Item{Sku: "43N23P", Name: "MacBook Pro", QtyAvailable: 5, Price: 539999})
		tx.Write(repo.ItemStoreName, "A304SD", item.Item{Sku: "A304SD", Name: "Alexa Speaker", QtyAvailable: 10, Price: 10950})
		tx.Write(repo.ItemStoreName, "234234", item.Item{Sku: "234234", Name: "Raspberry Pi B", QtyAvailable: 2, Price: 3000})
		return nil
	}
	if err := memDb.WithTx(seedInventory); err != nil {
		t.Fatalf("seed error: %v", err)
	}

	availablePromotions := []promotion.Promotion{
		promotion.FreeItemPromotion{
			PurchasedItemSku: "43N23P",
			FreeItemSku:      "234234",
			FreeItemPrice:    3000,
		},
		promotion.ItemQtyPriceFreePromotion{
			PurchasedItemSku: "120P90",
			PurchasedQty:     3,
		},
		promotion.ItemQtyPriceDiscountPercentagePromotion{
			PurchasedItemSku:   "A304SD",
			PurchasedQty:       3,
			PercentageDiscount: 0.1,
		},
	}

	itemRepo := repo.NewItemRepository(memDb)
	cartRepo := repo.NewCartRepository(memDb)

	app := utils.NewServer(0) // handler only; not starting a real listener
	if err := route.SetRoutes(app, itemRepo, cartRepo, availablePromotions); err != nil {
		t.Fatalf("route setup error: %v", err)
	}
	// Host the handler in an httptest server
	ts := httptest.NewServer(app.Handler)
	defer ts.Close()

	// Create MCP server configured to call the in-memory flip-shop server
	cfg := Config{BaseURL: ts.URL, Timeout: 5 * time.Second}
	s := NewServer(log.New(io.Discard, "", 0), cfg)

	// 1) create cart
	res, err := s.invoke(context.Background(), "cart.create", map[string]any{})
	if err != nil {
		t.Fatalf("cart.create error: %v", err)
	}
	cr, ok := res.(cartResponse)
	if !ok || cr.Cart == nil {
		t.Fatalf("unexpected create response: %#v", res)
	}
	// extract cartID from the generic map
	b, _ := json.Marshal(cr.Cart)
	var cart struct{ CartID string }
	_ = json.Unmarshal(b, &cart)
	if cart.CartID == "" {
		t.Fatalf("missing cart id in response: %s", string(b))
	}

	// 2) add purchase
	res, err = s.invoke(context.Background(), "cart.purchase.add", map[string]any{
		"cartID": cart.CartID,
		"sku":    "120P90",
		"qty":    3,
	})
	if err != nil {
		t.Fatalf("cart.purchase.add error: %v", err)
	}
	_ = res.(cartResponse)

	// 3) submit
	res, err = s.invoke(context.Background(), "cart.submit", map[string]any{"cartID": cart.CartID})
	if err != nil {
		t.Fatalf("cart.submit error: %v", err)
	}
	cr2 := res.(cartResponse)
	// Basic sanity: response still includes same cart ID
	b2, _ := json.Marshal(cr2.Cart)
	var submitted struct{ CartID string }
	_ = json.Unmarshal(b2, &submitted)
	if submitted.CartID != cart.CartID {
		t.Fatalf("submit returned different cart id: got %s want %s", submitted.CartID, cart.CartID)
	}
}

