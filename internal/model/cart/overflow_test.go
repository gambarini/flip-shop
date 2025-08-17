package cart

import (
	"math"
	"testing"

	"github.com/gambarini/flip-shop/internal/model/item"
)

// Test that SubmitCart guards against integer overflow and saturates at MaxInt64
func TestSubmitCart_OverflowGuard(t *testing.T) {
	c := Cart{
		CartID:     "C",
		CartStatus: CartStatusAvailable,
		Purchases:  map[item.Sku]Purchase{},
	}
	// Very large price and quantity to force overflow in naive multiplication
	p := Purchase{Sku: item.Sku("BIG"), Name: "Big", Price: math.MaxInt64 / 2, Qty: 3, Discount: 0}
	c.Purchases[p.Sku] = p

	if err := c.SubmitCart(); err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if c.Total != math.MaxInt64 {
		t.Fatalf("expected saturated total %d, got %d", math.MaxInt64, c.Total)
	}
}
