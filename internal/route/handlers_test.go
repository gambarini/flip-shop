package route

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gambarini/flip-shop/internal/model/item"
	"github.com/gambarini/flip-shop/internal/model/promotion"
	"github.com/gambarini/flip-shop/internal/repo"
	"github.com/gambarini/flip-shop/utils"
	"github.com/gambarini/flip-shop/utils/memdb"
)

// Test constants mirroring main.go inventory
const (
	ItemGoogleHomeSku   = "120P90"
	ItemMacBookProSku   = "43N23P"
	ItemAlexaSpeakerSku = "A304SD"
	RaspberryPiSku      = "234234"
)

type testEnv struct {
	srv      *utils.AppServer
	itemRepo repo.IItemRepository
	cartRepo repo.ICartRepository
}

func setupTestEnv(t *testing.T) testEnv {
	t.Helper()
	kv := memdb.NewMemoryKVDatabase()
	// Seed items
	if err := kv.WithTx(func(tx utils.Tx) error {
		tx.Write(repo.ItemStoreName, ItemGoogleHomeSku, item.Item{Sku: ItemGoogleHomeSku, Name: "Google Home", QtyAvailable: 10, Price: 4999})
		tx.Write(repo.ItemStoreName, ItemMacBookProSku, item.Item{Sku: ItemMacBookProSku, Name: "MacBook Pro", QtyAvailable: 5, Price: 539999})
		tx.Write(repo.ItemStoreName, ItemAlexaSpeakerSku, item.Item{Sku: ItemAlexaSpeakerSku, Name: "Alexa Speaker", QtyAvailable: 10, Price: 10950})
		tx.Write(repo.ItemStoreName, RaspberryPiSku, item.Item{Sku: RaspberryPiSku, Name: "Raspberry Pi B", QtyAvailable: 2, Price: 3000})
		return nil
	}); err != nil {
		t.Fatalf("seed failed: %v", err)
	}

	itemRepo := repo.NewItemRepository(kv)
	cartRepo := repo.NewCartRepository(kv)

	promos := []promotion.Promotion{
		promotion.FreeItemPromotion{PurchasedItemSku: ItemMacBookProSku, FreeItemSku: RaspberryPiSku, FreeItemPrice: 3000},
		promotion.ItemQtyPriceFreePromotion{PurchasedItemSku: ItemGoogleHomeSku, PurchasedQty: 3},
		promotion.ItemQtyPriceDiscountPercentagePromotion{PurchasedItemSku: ItemAlexaSpeakerSku, PurchasedQty: 3, PercentageDiscount: 0.1},
	}

	srv := utils.NewServer(0) // we won't start the server; we only use its router
	if err := SetRoutes(srv, itemRepo, cartRepo, promos); err != nil {
		t.Fatalf("set routes: %v", err)
	}

	return testEnv{srv: srv, itemRepo: itemRepo, cartRepo: cartRepo}
}

func doJSON(t *testing.T, srv *utils.AppServer, method, path string, body interface{}) *httptest.ResponseRecorder {
	t.Helper()
	var buf bytes.Buffer
	if body != nil {
		if err := json.NewEncoder(&buf).Encode(body); err != nil {
			t.Fatalf("encode body: %v", err)
		}
	}
	req := httptest.NewRequest(method, path, &buf)
	req.Header.Set("Content-Type", "application/json")
	rr := httptest.NewRecorder()
	srv.Handler.ServeHTTP(rr, req)
	return rr
}

func TestPostCart_HappyPath(t *testing.T) {
	env := setupTestEnv(t)
	rr := doJSON(t, env.srv, http.MethodPost, "/cart", nil)
	if rr.Code != http.StatusCreated {
		t.Fatalf("expected 201, got %d body=%s", rr.Code, rr.Body.String())
	}
	var resp struct {
		CartID     string `json:"CartID"`
		CartStatus string `json:"CartStatus"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if resp.CartID == "" || resp.CartStatus != "Available" {
		t.Fatalf("unexpected cart response: %+v", resp)
	}
}

func createCart(t *testing.T, srv *utils.AppServer) string {
	rr := doJSON(t, srv, http.MethodPost, "/cart", nil)
	if rr.Code != http.StatusCreated {
		t.Fatalf("create cart failed: %d", rr.Code)
	}
	var resp struct {
		CartID string `json:"CartID"`
	}
	_ = json.Unmarshal(rr.Body.Bytes(), &resp)
	return resp.CartID
}

func TestPurchase_HappyAndValidation(t *testing.T) {
	env := setupTestEnv(t)
	cid := createCart(t, env.srv)

	// happy path purchase 1 Google Home
	rr := doJSON(t, env.srv, http.MethodPut, "/cart/"+cid+"/purchase", map[string]interface{}{"sku": ItemGoogleHomeSku, "qty": 1})
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	// invalid qty
	rr = doJSON(t, env.srv, http.MethodPut, "/cart/"+cid+"/purchase", map[string]interface{}{"sku": ItemGoogleHomeSku, "qty": 0})
	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 for invalid qty, got %d", rr.Code)
	}
	// unavailable item (request huge qty)
	rr = doJSON(t, env.srv, http.MethodPut, "/cart/"+cid+"/purchase", map[string]interface{}{"sku": ItemGoogleHomeSku, "qty": 1000})
	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 for unavailable, got %d", rr.Code)
	}
}

func TestRemove_Happy_AndErrors(t *testing.T) {
	env := setupTestEnv(t)
	cid := createCart(t, env.srv)
	// purchase 2
	_ = doJSON(t, env.srv, http.MethodPut, "/cart/"+cid+"/purchase", map[string]interface{}{"sku": ItemGoogleHomeSku, "qty": 2})
	// happy remove 1
	rr := doJSON(t, env.srv, http.MethodDelete, "/cart/"+cid+"/purchase", map[string]interface{}{"sku": ItemGoogleHomeSku, "qty": 1})
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d", rr.Code)
	}
	// invalid qty
	rr = doJSON(t, env.srv, http.MethodDelete, "/cart/"+cid+"/purchase", map[string]interface{}{"sku": ItemGoogleHomeSku, "qty": 0})
	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 for invalid qty, got %d", rr.Code)
	}
	// 404 unknown cart
	rr = doJSON(t, env.srv, http.MethodDelete, "/cart/non-existent-id/purchase", map[string]interface{}{"sku": ItemGoogleHomeSku, "qty": 1})
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404 for missing cart, got %d", rr.Code)
	}
}

func TestSubmit_WithPromotions(t *testing.T) {
	env := setupTestEnv(t)
	cid := createCart(t, env.srv)
	// Buy 1 MacBook Pro -> free Raspberry Pi fully discounted; total should equal MacBook price
	rr := doJSON(t, env.srv, http.MethodPut, "/cart/"+cid+"/purchase", map[string]interface{}{"sku": ItemMacBookProSku, "qty": 1})
	if rr.Code != http.StatusOK {
		t.Fatalf("purchase failed: %d", rr.Code)
	}
	rr = doJSON(t, env.srv, http.MethodPut, "/cart/"+cid+"/status/submitted", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("submit failed: %d body=%s", rr.Code, rr.Body.String())
	}
	var resp struct {
		CartStatus string `json:"CartStatus"`
		Total      int64  `json:"Total"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid submit json: %v", err)
	}
	if resp.CartStatus != "Submitted" {
		t.Fatalf("expected Submitted, got %s", resp.CartStatus)
	}
	if resp.Total != 539999 { // equals MacBook price, Raspberry Pi free
		t.Fatalf("expected total 539999, got %d", resp.Total)
	}
}

// failingPromotion always returns an error on Apply
type failingPromotion struct{}

func (f failingPromotion) Apply(_ promotion.GetPurchasedItemHandler, _ promotion.AddPromoItemToCartHandler, _ promotion.AddDiscountToCartHandler) error {
	return fmt.Errorf("promo failed")
}

// countingPromotion increments counter when Apply is called
type countingPromotion struct{ calls *int }

func (c countingPromotion) Apply(_ promotion.GetPurchasedItemHandler, _ promotion.AddPromoItemToCartHandler, _ promotion.AddDiscountToCartHandler) error {
	*(c.calls)++
	return nil
}

func TestSubmit_PromotionErrorShortCircuits(t *testing.T) {
	// Setup env with custom promotions: first fails, second counts
	kv := memdb.NewMemoryKVDatabase()
	if err := kv.WithTx(func(tx utils.Tx) error {
		tx.Write(repo.ItemStoreName, ItemGoogleHomeSku, item.Item{Sku: ItemGoogleHomeSku, Name: "Google Home", QtyAvailable: 10, Price: 4999})
		return nil
	}); err != nil {
		t.Fatalf("seed failed: %v", err)
	}
	itemRepo := repo.NewItemRepository(kv)
	cartRepo := repo.NewCartRepository(kv)
	calls := 0
	promos := []promotion.Promotion{failingPromotion{}, countingPromotion{calls: &calls}}
	srv := utils.NewServer(0)
	if err := SetRoutes(srv, itemRepo, cartRepo, promos); err != nil {
		t.Fatalf("set routes: %v", err)
	}
	cid := createCart(t, srv)
	// make any purchase so submit will process promotions
	rr := doJSON(t, srv, http.MethodPut, "/cart/"+cid+"/purchase", map[string]interface{}{"sku": ItemGoogleHomeSku, "qty": 1})
	if rr.Code != http.StatusOK {
		t.Fatalf("purchase failed: %d", rr.Code)
	}

	rr = doJSON(t, srv, http.MethodPut, "/cart/"+cid+"/status/submitted", nil)
	if rr.Code != http.StatusInternalServerError { // generic 500 on unexpected promotion error
		t.Fatalf("expected 500 on failing promotion, got %d", rr.Code)
	}
	if calls != 0 {
		t.Fatalf("expected next promotion not to run, but Apply was called %d times", calls)
	}
}

func TestGetCart_OK(t *testing.T) {
	env := setupTestEnv(t)
	cid := createCart(t, env.srv)
	rr := doJSON(t, env.srv, http.MethodGet, "/cart/"+cid, nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	var resp struct {
		CartID string `json:"CartID"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &resp); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if resp.CartID != cid {
		t.Fatalf("expected same cartID, got %s", resp.CartID)
	}
}

func TestGetCart_NotFound(t *testing.T) {
	env := setupTestEnv(t)
	// random but fixed UUID that does not exist
	rr := doJSON(t, env.srv, http.MethodGet, "/cart/123e4567-e89b-12d3-a456-426614174000", nil)
	if rr.Code != http.StatusNotFound {
		t.Fatalf("expected 404, got %d", rr.Code)
	}
}

func TestGetCart_InvalidID(t *testing.T) {
	env := setupTestEnv(t)
	rr := doJSON(t, env.srv, http.MethodGet, "/cart/not-a-uuid", nil)
	if rr.Code != http.StatusUnprocessableEntity {
		t.Fatalf("expected 422 for invalid id, got %d", rr.Code)
	}
}
