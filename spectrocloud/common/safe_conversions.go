package common

// SafeUint32 converts int to uint32 with bounds checking to prevent overflow
func SafeUint32(value int) uint32 {
	if value < 0 {
		return 0
	}
	// On 32-bit systems, int max is smaller than uint32 max, so no overflow possible
	// On 64-bit systems, we need to check against uint32 max
	if ^uint(0)>>32 == 0 {
		// 32-bit system: int and uint32 have same size, no overflow possible
		return uint32(value)
	}
	// 64-bit system: check against uint32 max
	if uint64(value) > 0xFFFFFFFF {
		return 0xFFFFFFFF
	}
	return uint32(value)
}
