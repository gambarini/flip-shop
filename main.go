package main

import (
	"encoding/json"
	"log"
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

var (
	memDb               *memdb.MemoryKVDatabase
	availablePromotions []promotion.Promotion
)

func init() {
	// Prepare the memory database with the items before the application starts
	memDb = memdb.NewMemoryKVDatabase()

	// Optionally seed inventory from environment variable FLIPSHOP_INVENTORY_JSON
	// Expected format: [{"sku":"120P90","name":"Google Home","price":4999,"qty":10}, ...]
	type invItem struct {
		Sku   string `json:"sku"`
		Name  string `json:"name"`
		Price int64  `json:"price"`
		Qty   int    `json:"qty"`
	}

	seed := func(tx utils.Tx) error {
		// Default inventory
		tx.Write(repo.ItemStoreName, ItemGoogleHomeSku, item.Item{Sku: ItemGoogleHomeSku, Name: "Google Home", QtyAvailable: 10, Price: 4999, QtyReserved: 0})
		tx.Write(repo.ItemStoreName, ItemMacBookProSku, item.Item{Sku: ItemMacBookProSku, Name: "MacBook Pro", QtyAvailable: 5, Price: 539999, QtyReserved: 0})
		tx.Write(repo.ItemStoreName, ItemAlexaSpeakerSku, item.Item{Sku: ItemAlexaSpeakerSku, Name: "Alexa Speaker", QtyAvailable: 10, Price: 10950, QtyReserved: 0})
		tx.Write(repo.ItemStoreName, RaspberyPiSku, item.Item{Sku: RaspberyPiSku, Name: "Raspberry Pi B", QtyAvailable: 2, Price: 3000, QtyReserved: 0})
		return nil
	}

	if invJSON := os.Getenv("FLIPSHOP_INVENTORY_JSON"); invJSON != "" {
		_ = memDb.WithTx(func(tx utils.Tx) error {
			var items []invItem
			if err := json.Unmarshal([]byte(invJSON), &items); err != nil {
				// fallback to defaults on parse error
				return seed(tx)
			}
			// clear existing by reinitializing store snapshot (overwrite writes)
			for _, it := range items {
				if it.Sku == "" || it.Price < 0 || it.Qty < 0 {
					continue
				}
				tx.Write(repo.ItemStoreName, it.Sku, item.Item{Sku: item.Sku(it.Sku), Name: it.Name, QtyAvailable: it.Qty, Price: it.Price, QtyReserved: 0})
			}
			return nil
		})
	} else {
		if err := memDb.WithTx(seed); err != nil {
			log.Fatalf("Error initializing, %s", err)
		}
	}
}

func main() {

	initializeFunc := func(srv *utils.AppServer) (err error) {

		// Here we setup the expected promotions
		availablePromotions = append(availablePromotions, promotion.FreeItemPromotion{
			PurchasedItemSku: ItemMacBookProSku,
			FreeItemSku:      RaspberyPiSku,
			FreeItemPrice:    3000,
		})

		availablePromotions = append(availablePromotions, promotion.ItemQtyPriceFreePromotion{
			PurchasedItemSku: ItemGoogleHomeSku,
			PurchasedQty:     3,
		})

		availablePromotions = append(availablePromotions, promotion.ItemQtyPriceDiscountPercentagePromotion{
			PurchasedItemSku:   ItemAlexaSpeakerSku,
			PurchasedQty:       3,
			PercentageDiscount: 0.1,
		})

		itemRepo := repo.NewItemRepository(memDb)
		cartRepo := repo.NewCartRepository(memDb)

		err = route.SetRoutes(srv, itemRepo, cartRepo, availablePromotions)

		if err != nil {
			return err
		}

		return nil
	}

	cleanupFunc := func(srv *utils.AppServer) (err error) {
		return nil
	}

	// Configure port and version from environment variables, preserving defaults
	port := 8001
	if p := os.Getenv("FLIPSHOP_PORT"); p == "" {
		// also support PORT for common platforms
		p = os.Getenv("PORT")
		if p != "" {
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

	srv := utils.NewServerWithInitialization(port, initializeFunc, cleanupFunc)
	srv.Version = version

	srv.Start()
}
