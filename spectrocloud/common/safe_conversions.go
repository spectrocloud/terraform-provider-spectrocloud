package common

import "math"

const maxUint32 = 0xFFFFFFFF

// SafeUint32 converts int to uint32 with bounds checking to prevent overflow
func SafeUint32(value int) uint32 {
	if value < 0 {
		return 0
	}
	// On 32-bit systems, int max is smaller than uint32 max, so no overflow possible
	// On 64-bit systems, we need to check against uint32 max
	if ^uint(0)>>32 == 0 {
		// 32-bit system: int and uint32 have same size, no overflow possible
		if value >= 0 {
			return uint32(value) // #nosec G115 -- 32-bit: int and uint32 same range
		}
		return 0
	}
	// 64-bit system: check against uint32 max
	// Avoid comparing against an untyped constant as `int` on 32-bit targets.
	if uint64(value) > uint64(math.MaxUint32) {
		return uint32(math.MaxUint32)
	}
	return uint32(value) // #nosec G115 -- value is non-negative and bounded by MaxUint32 above
}

// SafeUintToInt converts uint to int with bounds checking to prevent overflow (G115).
// Values larger than math.MaxInt are clamped to math.MaxInt.
func SafeUintToInt(value uint) int {
	if value > math.MaxInt {
		return math.MaxInt
	}
	return int(value)
}

// SafeIntToUint converts int to uint with bounds checking to prevent overflow (G115).
// Negative values return 0.
func SafeIntToUint(value int) uint {
	if value <= 0 {
		return 0
	}
	return uint(value)
}
