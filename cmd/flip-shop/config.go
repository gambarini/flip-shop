package main

import (
	"errors"
	"fmt"
	"os"

	"github.com/gambarini/flip-shop/internal/model/item"
	"github.com/gambarini/flip-shop/internal/model/promotion"
	"gopkg.in/yaml.v3"
)

type YAMLItem struct {
	Sku   string `yaml:"sku"`
	Name  string `yaml:"name"`
	Price int64  `yaml:"price"`
	Qty   int    `yaml:"qty"`
}

type YAMLTier struct {
	MinQty             int     `yaml:"min_qty"`
	PercentageDiscount float32 `yaml:"percentage_discount"`
}

type YAMLPromotion struct {
	Type               string     `yaml:"type"`
	Name               string     `yaml:"name"`
	PurchasedItemSku   string     `yaml:"purchased_item_sku"`
	FreeItemSku        string     `yaml:"free_item_sku"`
	PurchasedQty       int        `yaml:"purchased_qty"`
	PercentageDiscount float32    `yaml:"percentage_discount"`
	FixedDiscount      int64      `yaml:"fixed_discount"`
	ItemSkus           []string   `yaml:"item_skus"`
	Tiers              []YAMLTier `yaml:"tiers"`
	ThresholdAmount    int64      `yaml:"threshold_amount"`
	MinItemCount       int        `yaml:"min_item_count"`
}

type YAMLConfig struct {
	Items      []YAMLItem      `yaml:"items"`
	Promotions []YAMLPromotion `yaml:"promotions"`
}

// yaml.v3 leaves a slice nil when the key is absent; sets it to []T{} when present-but-empty.
func (c *YAMLConfig) ItemsProvided() bool      { return c != nil && c.Items != nil }
func (c *YAMLConfig) PromotionsProvided() bool { return c != nil && c.Promotions != nil }

func loadYAMLConfig(path string) (*YAMLConfig, error) {
	if path == "" {
		return nil, nil
	}
	f, err := os.Open(path)
	if err != nil {
		if errors.Is(err, os.ErrNotExist) {
			fmt.Fprintf(os.Stderr, "warning: FLIPSHOP_CONFIG_FILE=%q not found, using defaults\n", path)
			return nil, nil
		}
		return nil, fmt.Errorf("open config file: %w", err)
	}
	defer f.Close()
	dec := yaml.NewDecoder(f)
	dec.KnownFields(true)
	var cfg YAMLConfig
	if err := dec.Decode(&cfg); err != nil {
		return nil, fmt.Errorf("parse config file %q: %w", path, err)
	}
	return &cfg, nil
}

func toPromotions(yps []YAMLPromotion) ([]promotion.Promotion, error) {
	out := make([]promotion.Promotion, 0, len(yps))
	for i, yp := range yps {
		switch yp.Type {
		case "free_item":
			out = append(out, promotion.FreeItemPromotion{
				Name:             yp.Name,
				PurchasedItemSku: item.Sku(yp.PurchasedItemSku),
				FreeItemSku:      item.Sku(yp.FreeItemSku),
			})
		case "qty_price_free":
			out = append(out, promotion.ItemQtyPriceFreePromotion{
				Name:             yp.Name,
				PurchasedItemSku: item.Sku(yp.PurchasedItemSku),
				PurchasedQty:     yp.PurchasedQty,
			})
		case "qty_discount_percentage":
			out = append(out, promotion.ItemQtyPriceDiscountPercentagePromotion{
				Name:               yp.Name,
				PurchasedItemSku:   item.Sku(yp.PurchasedItemSku),
				PurchasedQty:       yp.PurchasedQty,
				PercentageDiscount: yp.PercentageDiscount,
			})
		case "qty_discount_fixed":
			out = append(out, promotion.ItemQtyFixedDiscountPromotion{
				Name:             yp.Name,
				PurchasedItemSku: item.Sku(yp.PurchasedItemSku),
				PurchasedQty:     yp.PurchasedQty,
				FixedDiscount:    yp.FixedDiscount,
			})
		case "qty_half_price":
			out = append(out, promotion.ItemQtyHalfPricePromotion{
				Name:             yp.Name,
				PurchasedItemSku: item.Sku(yp.PurchasedItemSku),
				PurchasedQty:     yp.PurchasedQty,
			})
		case "qty_tiered_discount":
			tiers := make([]promotion.DiscountTier, 0, len(yp.Tiers))
			for _, t := range yp.Tiers {
				tiers = append(tiers, promotion.DiscountTier{
					MinQty:             t.MinQty,
					PercentageDiscount: t.PercentageDiscount,
				})
			}
			out = append(out, promotion.ItemQtyTieredDiscountPromotion{
				Name:             yp.Name,
				PurchasedItemSku: item.Sku(yp.PurchasedItemSku),
				Tiers:            tiers,
			})
		case "bundle_discount":
			skus := make([]item.Sku, 0, len(yp.ItemSkus))
			for _, s := range yp.ItemSkus {
				skus = append(skus, item.Sku(s))
			}
			out = append(out, promotion.BundleDiscountPromotion{
				Name:               yp.Name,
				ItemSkus:           skus,
				PercentageDiscount: yp.PercentageDiscount,
			})
		case "cart_spend_threshold":
			out = append(out, promotion.CartSpendThresholdPromotion{
				Name:               yp.Name,
				ThresholdAmount:    yp.ThresholdAmount,
				PercentageDiscount: yp.PercentageDiscount,
			})
		case "cheapest_item_free":
			out = append(out, promotion.CheapestItemFreePromotion{
				Name:         yp.Name,
				MinItemCount: yp.MinItemCount,
			})
		default:
			return nil, fmt.Errorf("promotion[%d]: unknown type %q", i, yp.Type)
		}
	}
	return out, nil
}
