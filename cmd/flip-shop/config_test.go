package main

import (
	"os"
	"path/filepath"
	"testing"

	"github.com/gambarini/flip-shop/internal/model/promotion"
)

func TestLoadYAMLConfig(t *testing.T) {
	t.Run("empty path returns nil nil", func(t *testing.T) {
		cfg, err := loadYAMLConfig("")
		if err != nil || cfg != nil {
			t.Fatalf("expected nil,nil got cfg=%v err=%v", cfg, err)
		}
	})

	t.Run("nonexistent path returns nil nil", func(t *testing.T) {
		cfg, err := loadYAMLConfig("/nonexistent/path/config.yaml")
		if err != nil || cfg != nil {
			t.Fatalf("expected nil,nil for missing file, got cfg=%v err=%v", cfg, err)
		}
	})

	t.Run("valid full config", func(t *testing.T) {
		path := writeTempYAML(t, `
items:
  - sku: "SKU1"
    name: "Item One"
    price: 100
    qty: 5
promotions:
  - type: "free_item"
    purchased_item_sku: "SKU1"
    free_item_sku: "SKU2"
`)
		cfg, err := loadYAMLConfig(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !cfg.ItemsProvided() || len(cfg.Items) != 1 {
			t.Fatalf("expected 1 item, got %+v", cfg.Items)
		}
		if !cfg.PromotionsProvided() || len(cfg.Promotions) != 1 {
			t.Fatalf("expected 1 promotion, got %+v", cfg.Promotions)
		}
	})

	t.Run("malformed YAML returns error", func(t *testing.T) {
		path := writeTempYAML(t, "items: [invalid yaml :::")
		_, err := loadYAMLConfig(path)
		if err == nil {
			t.Fatal("expected error for malformed YAML")
		}
	})

	t.Run("unknown key returns error", func(t *testing.T) {
		path := writeTempYAML(t, "unknown_key: value\n")
		_, err := loadYAMLConfig(path)
		if err == nil {
			t.Fatal("expected error for unknown YAML key")
		}
	})

	t.Run("items key absent means not provided", func(t *testing.T) {
		path := writeTempYAML(t, "promotions: []\n")
		cfg, err := loadYAMLConfig(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if cfg.ItemsProvided() {
			t.Fatal("expected items not provided when key absent")
		}
		if !cfg.PromotionsProvided() {
			t.Fatal("expected promotions provided for explicit empty list")
		}
	})

	t.Run("explicit empty items list is provided", func(t *testing.T) {
		path := writeTempYAML(t, "items: []\n")
		cfg, err := loadYAMLConfig(path)
		if err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
		if !cfg.ItemsProvided() {
			t.Fatal("expected items provided for explicit empty list")
		}
		if len(cfg.Items) != 0 {
			t.Fatalf("expected 0 items, got %d", len(cfg.Items))
		}
	})
}

func TestToPromotions(t *testing.T) {
	tests := []struct {
		name    string
		input   []YAMLPromotion
		wantLen int
		wantErr bool
		wantType interface{}
	}{
		{
			name:     "free_item",
			input:    []YAMLPromotion{{Type: "free_item", PurchasedItemSku: "A", FreeItemSku: "B"}},
			wantLen:  1,
			wantType: promotion.FreeItemPromotion{},
		},
		{
			name:     "qty_price_free",
			input:    []YAMLPromotion{{Type: "qty_price_free", PurchasedItemSku: "A", PurchasedQty: 2}},
			wantLen:  1,
			wantType: promotion.ItemQtyPriceFreePromotion{},
		},
		{
			name:     "qty_discount_percentage",
			input:    []YAMLPromotion{{Type: "qty_discount_percentage", PurchasedItemSku: "A", PurchasedQty: 2, PercentageDiscount: 0.5}},
			wantLen:  1,
			wantType: promotion.ItemQtyPriceDiscountPercentagePromotion{},
		},
		{
			name:    "unknown type returns error",
			input:   []YAMLPromotion{{Type: "invalid_type"}},
			wantErr: true,
		},
		{
			name:    "empty slice is valid",
			input:   []YAMLPromotion{},
			wantLen: 0,
		},
		{
			name: "mixed valid types",
			input: []YAMLPromotion{
				{Type: "free_item", PurchasedItemSku: "A", FreeItemSku: "B"},
				{Type: "qty_price_free", PurchasedItemSku: "A", PurchasedQty: 3},
			},
			wantLen: 2,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := toPromotions(tt.input)
			if (err != nil) != tt.wantErr {
				t.Fatalf("toPromotions() error = %v, wantErr %v", err, tt.wantErr)
			}
			if len(got) != tt.wantLen {
				t.Fatalf("len = %d, want %d", len(got), tt.wantLen)
			}
		})
	}
}

func writeTempYAML(t *testing.T, content string) string {
	t.Helper()
	path := filepath.Join(t.TempDir(), "config.yaml")
	if err := os.WriteFile(path, []byte(content), 0644); err != nil {
		t.Fatal(err)
	}
	return path
}
