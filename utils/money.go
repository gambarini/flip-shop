package utils

import (
	"math"
)

// Money representation: all monetary values are represented as integer cents (int64).
// The helpers below provide saturating arithmetic to guard against integer overflow.
// On overflow, results clamp to math.MaxInt64 (never negative), matching a defensive stance.
// Note: In typical usage with realistic prices/quantities, overflow should never occur.

// SaturatingAddInt64 returns a+b, clamping to MaxInt64 on overflow.
func SaturatingAddInt64(a, b int64) int64 {
	if b > 0 && a > math.MaxInt64-b {
		return math.MaxInt64
	}
	if b < 0 && a < math.MinInt64-b {
		return math.MinInt64
	}
	return a + b
}

// SaturatingSubInt64 returns a-b, clamping to 0 on underflow for money semantics.
func SaturatingSubInt64(a, b int64) int64 {
	// We treat negatives as zero to avoid negative totals/discounts.
	if b > a {
		return 0
	}
	return a - b
}

// SaturatingMulInt64 returns a*b with int64 operands, clamping to MaxInt64 on overflow.
func SaturatingMulInt64(a, b int64) int64 {
	if a == 0 || b == 0 {
		return 0
	}
	// Only handle positive domain for money semantics (prices/qty are non-negative)
	if a > 0 && b > 0 {
		if a > math.MaxInt64/b {
			return math.MaxInt64
		}
		return a * b
	}
	// Any negative paths clamp to 0 (should not occur for money values)
	prod := a * b
	if prod < 0 {
		return 0
	}
	return prod
}

// SaturatingMulInt64Int returns a*int64(b), clamping to MaxInt64 on overflow.
func SaturatingMulInt64Int(a int64, b int) int64 {
	if b < 0 {
		// not expected for quantities; treat as zero to be defensive
		return 0
	}
	return SaturatingMulInt64(a, int64(b))
}
