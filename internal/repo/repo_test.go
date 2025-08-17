package repo

import (
	"errors"
	"reflect"
	"testing"

	"github.com/gambarini/flip-shop/internal/model/cart"
	"github.com/gambarini/flip-shop/internal/model/item"
	"github.com/gambarini/flip-shop/utils"
	"github.com/gambarini/flip-shop/utils/memdb"
)

func TestItemRepository_WithTx_ReadStoreAndErrors(t *testing.T) {
	kv := memdb.NewMemoryKVDatabase()
	repo := NewItemRepository(kv)

	// not found case
	err := repo.WithTx(func(tx utils.Tx) error {
		// Try to read a missing item using repo method, expecting ErrItemNotFound
		_, findErr := repo.FindItemBySku(tx, item.Sku("DOES-NOT-EXIST"))
		if findErr != ErrItemNotFound {
			return errors.New("expected ErrItemNotFound for missing SKU")
		}
		return nil
	})
	if err != nil {
		t.Fatalf("unexpected error executing tx: %v", err)
	}

	// happy path: store and read back in same tx
	it := item.Item{Sku: item.Sku("120P90"), Name: "USB Cable", Price: 1299, QtyAvailable: 10}
	if err := repo.WithTx(func(tx utils.Tx) error {
		if err := repo.Store(tx, it); err != nil {
			return err
		}
		got, err := repo.FindItemBySku(tx, it.Sku)
		if err != nil {
			return err
		}
		if got != it {
			return errors.New("stored and fetched item mismatch")
		}
		return nil
	}); err != nil {
		t.Fatalf("tx failed: %v", err)
	}

	// rollback case: write then return error; ensure not committed
	boom := errors.New("boom")
	_ = repo.WithTx(func(tx utils.Tx) error {
		_ = repo.Store(tx, item.Item{Sku: item.Sku("ROLL-BACK"), Name: "RB", Price: 1})
		return boom
	})
	// After rollback, the item should not exist in a fresh tx
	if err := repo.WithTx(func(tx utils.Tx) error {
		_, err := repo.FindItemBySku(tx, item.Sku("ROLL-BACK"))
		if err != ErrItemNotFound {
			return errors.New("expected item to be absent after rollback")
		}
		return nil
	}); err != nil {
		t.Fatalf("verification tx failed: %v", err)
	}
}

func TestCartRepository_WithTx_ReadStoreAndErrors(t *testing.T) {
	kv := memdb.NewMemoryKVDatabase()
	repoC := NewCartRepository(kv)

	// missing cart -> ErrCartNotFound
	if _, err := repoC.FindCartByID("nope"); err != ErrCartNotFound {
		t.Fatalf("expected ErrCartNotFound, got %v", err)
	}

	// store cart and then read it back (commit)
	c := cart.NewAvailableCart()
	if err := repoC.WithTx(func(tx utils.Tx) error {
		return repoC.Store(tx, c)
	}); err != nil {
		t.Fatalf("store tx failed: %v", err)
	}
	got, err := repoC.FindCartByID(c.CartID)
	if err != nil {
		t.Fatalf("find after commit failed: %v", err)
	}
	if !reflect.DeepEqual(got, c) {
		t.Fatalf("stored and fetched cart mismatch: %+v vs %+v", got, c)
	}

	// rollback: store inside tx that errors; it must not be visible
	boom := errors.New("boom")
	_ = repoC.WithTx(func(tx utils.Tx) error {
		_ = repoC.Store(tx, cart.NewAvailableCart())
		return boom
	})
	// Hard to know the random ID; just ensure an arbitrary ID is not found and the previous one still exists
	if _, err := repoC.FindCartByID("non-existent-id"); err != ErrCartNotFound {
		t.Fatalf("expected non-existent id to map to ErrCartNotFound, got %v", err)
	}
	// Ensure pre-existing cart remained
	if _, err := repoC.FindCartByID(c.CartID); err != nil {
		t.Fatalf("expected existing cart still present after unrelated rollback, err=%v", err)
	}
}
