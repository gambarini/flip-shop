package main

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"strconv"

	"github.com/gambarini/flip-shop/internal/model/item"
	"github.com/gambarini/flip-shop/internal/model/promotion"
	"github.com/gambarini/flip-shop/internal/repo"
	"github.com/gambarini/flip-shop/internal/route"
	"github.com/gambarini/flip-shop/utils"
	"github.com/gambarini/flip-shop/utils/memdb"
)

const (
	ItemGoogleHomeSku   = "120P90"
	ItemMacBookProSku   = "43N23P"
	ItemAlexaSpeakerSku = "A304SD"
	RaspberyPiSku       = "234234"
)

type invItem struct {
	Sku   string `json:"sku"`
	Name  string `json:"name"`
	Price int64  `json:"price"`
	Qty   int    `json:"qty"`
}

type config struct {
	port       int
	version    string
	configFile string
}

func loadConfig() config {
	port := 8001
	if p := os.Getenv("FLIPSHOP_PORT"); p == "" {
		if p = os.Getenv("PORT"); p != "" {
			if v, err := strconv.Atoi(p); err == nil {
				port = v
			}
		}
	} else {
		if v, err := strconv.Atoi(p); err == nil {
			port = v
		}
	}
	version := os.Getenv("FLIPSHOP_VERSION")
	if version == "" {
		version = "dev"
	}
	return config{
		port:       port,
		version:    version,
		configFile: os.Getenv("FLIPSHOP_CONFIG_FILE"),
	}
}

func newDatabase(seedItems []invItem) (*memdb.MemoryKVDatabase, error) {
	db := memdb.NewMemoryKVDatabase()

	seed := func(tx utils.Tx) error {
		if seedItems == nil {
			tx.Write(repo.ItemStoreName, ItemGoogleHomeSku, item.Item{Sku: ItemGoogleHomeSku, Name: "Google Home", QtyAvailable: 10, Price: 4999, QtyReserved: 0})
			tx.Write(repo.ItemStoreName, ItemMacBookProSku, item.Item{Sku: ItemMacBookProSku, Name: "MacBook Pro", QtyAvailable: 5, Price: 539999, QtyReserved: 0})
			tx.Write(repo.ItemStoreName, ItemAlexaSpeakerSku, item.Item{Sku: ItemAlexaSpeakerSku, Name: "Alexa Speaker", QtyAvailable: 10, Price: 10950, QtyReserved: 0})
			tx.Write(repo.ItemStoreName, RaspberyPiSku, item.Item{Sku: RaspberyPiSku, Name: "Raspberry Pi B", QtyAvailable: 2, Price: 3000, QtyReserved: 0})
			return nil
		}
		for _, it := range seedItems {
			if it.Sku == "" || it.Price < 0 || it.Qty < 0 {
				continue
			}
			tx.Write(repo.ItemStoreName, it.Sku, item.Item{Sku: item.Sku(it.Sku), Name: it.Name, QtyAvailable: it.Qty, Price: it.Price, QtyReserved: 0})
		}
		return nil
	}

	return db, db.WithTx(seed)
}

func defaultPromotions() []promotion.Promotion {
	return []promotion.Promotion{
		promotion.FreeItemPromotion{
			Name:             "Buy a MacBook Pro, get a Raspberry Pi B free",
			PurchasedItemSku: ItemMacBookProSku,
			FreeItemSku:      RaspberyPiSku,
		},
		promotion.ItemQtyPriceFreePromotion{
			Name:             "Buy 3 Google Homes, get 1 free",
			PurchasedItemSku: ItemGoogleHomeSku,
			PurchasedQty:     3,
		},
		promotion.ItemQtyPriceDiscountPercentagePromotion{
			Name:               "Buy 3 Alexa Speakers, get 10% off",
			PurchasedItemSku:   ItemAlexaSpeakerSku,
			PurchasedQty:       3,
			PercentageDiscount: 0.1,
		},
	}
}

func run() error {
	cfg := loadConfig()

	yamlCfg, err := loadYAMLConfig(cfg.configFile)
	if err != nil {
		return fmt.Errorf("config file: %w", err)
	}

	// Resolve items: YAML > JSON env > defaults
	var seedItems []invItem
	if yamlCfg.ItemsProvided() {
		if os.Getenv("FLIPSHOP_INVENTORY_JSON") != "" {
			fmt.Fprintln(os.Stderr, "warning: YAML config provides items; FLIPSHOP_INVENTORY_JSON is ignored")
		}
		for _, yi := range yamlCfg.Items {
			seedItems = append(seedItems, invItem(yi))
		}
	} else if invJSON := os.Getenv("FLIPSHOP_INVENTORY_JSON"); invJSON != "" {
		var parsed []invItem
		if err := json.Unmarshal([]byte(invJSON), &parsed); err == nil {
			seedItems = parsed
		}
	}

	db, err := newDatabase(seedItems)
	if err != nil {
		return fmt.Errorf("database initialization: %w", err)
	}

	// Resolve promotions: YAML > defaults
	var promotions []promotion.Promotion
	if yamlCfg.PromotionsProvided() {
		promotions, err = toPromotions(yamlCfg.Promotions)
		if err != nil {
			return fmt.Errorf("promotions config: %w", err)
		}
	} else {
		promotions = defaultPromotions()
	}

	initializeFunc := func(srv *utils.AppServer) error {
		for _, p := range promotions {
			if v, ok := p.(interface{ Validate() error }); ok {
				if err := v.Validate(); err != nil {
					return fmt.Errorf("invalid promotion configuration: %w", err)
				}
			}
		}
		itemRepo := repo.NewItemRepository(db)
		cartRepo := repo.NewCartRepository(db)
		return route.SetRoutes(srv, itemRepo, cartRepo, promotions)
	}

	cleanupFunc := func(srv *utils.AppServer) error {
		return nil
	}

	srv := utils.NewServerWithInitialization(cfg.port, initializeFunc, cleanupFunc)
	srv.Version = cfg.version

	return srv.Start(context.Background())
}

func main() {
	if err := run(); err != nil {
		fmt.Fprintf(os.Stderr, "fatal: %v\n", err)
		os.Exit(1)
	}
}
