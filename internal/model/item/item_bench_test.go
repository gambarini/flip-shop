package item

import "testing"

func BenchmarkItemReserveRelease(b *testing.B) {
	it := Item{Sku: Sku("X"), Name: "Bench", Price: 1000, QtyAvailable: 1_000_000}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = it.ReserveItem(1)
		_ = it.ReleaseItem(1)
	}
}
