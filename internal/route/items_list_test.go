package route

import (
	"encoding/json"
	"net/http"
	"testing"
)

func TestListItems_ReturnsSeeded(t *testing.T) {
	env := setupTestEnv(t)
	rr := doJSON(t, env.srv, http.MethodGet, "/items", nil)
	if rr.Code != http.StatusOK {
		t.Fatalf("expected 200, got %d body=%s", rr.Code, rr.Body.String())
	}
	var items []struct{
		Sku string `json:"Sku"`
	}
	if err := json.Unmarshal(rr.Body.Bytes(), &items); err != nil {
		t.Fatalf("invalid json: %v", err)
	}
	if len(items) < 4 {
		t.Fatalf("expected at least 4 items, got %d", len(items))
	}
	// Quick presence checks
	found := map[string]bool{}
	for _, it := range items {
		found[it.Sku] = true
	}
	if !found[ItemGoogleHomeSku] || !found[ItemMacBookProSku] || !found[ItemAlexaSpeakerSku] || !found[RaspberryPiSku] {
		t.Fatalf("expected seeded SKUs present, got: %+v", found)
	}
}
