package main

import (
	"github.com/gambarini/flip-shop/internal/model/item"
	"github.com/gambarini/flip-shop/internal/model/promotion"
	"github.com/gambarini/flip-shop/internal/repo"
	"github.com/gambarini/flip-shop/internal/route"
	"github.com/gambarini/flip-shop/utils"
	"github.com/gambarini/flip-shop/utils/memdb"
	"log"
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

func init() () {

	// Here we prepare the memory database with the items
	// before the application starts
	memDb = memdb.NewDMemoryKVDatabase()

	err := memDb.WithTx(func(tx utils.Tx) error {

		tx.Write("Items", ItemGoogleHomeSku, item.Item{
			Sku:          ItemGoogleHomeSku,
			Name:         "Google Home",
			QtyAvailable: 10,
			Price:        4999,
			QtyReserved:  0,
		})

		tx.Write("Items", ItemMacBookProSku, item.Item{
			Sku:          ItemMacBookProSku,
			Name:         "MacBook Pro",
			QtyAvailable: 5,
			Price:        539999,
			QtyReserved:  0,
		})

		tx.Write("Items", ItemAlexaSpeakerSku, item.Item{
			Sku:          ItemAlexaSpeakerSku,
			Name:         "Alexa Speaker",
			QtyAvailable: 10,
			Price:        10950,
			QtyReserved:  0,
		})

		tx.Write("Items", RaspberyPiSku, item.Item{
			Sku:          RaspberyPiSku,
			Name:         "Raspberry Pi B",
			QtyAvailable: 2,
			Price:        3000,
			QtyReserved:  0,
		})

		return nil
	})

	if err != nil {
		log.Fatalf("Error initializing, %s", err)
	}

}

func main() {

	initializeFunc := func(srv *utils.AppServer) (err error) {

		// Here we setup the expected promotions
		availablePromotions = append(availablePromotions, promotion.FreeItemPromotion{
			PurchasedItemSku: ItemMacBookProSku,
			FreeItemSku: RaspberyPiSku,
			FreeItemPrice: 3000,
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

	srv := utils.NewServerWithInitialization(8001, initializeFunc, cleanupFunc)

	srv.Start()
}
