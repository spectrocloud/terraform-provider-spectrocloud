package common

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestSafeUint32 exercises every branch of SafeUint32: the negative-clamp
// path, the "no overflow" happy path, the MaxUint32-clamp path, and the
// exact boundary. Written as a table because they're the same shape and a
// regression here would break every resource that reports a numeric field
// (retention days, port numbers, etc.) up to the Palette API.
func TestSafeUint32(t *testing.T) {
	tests := []struct {
		name  string
		input int
		want  uint32
	}{
		{"zero", 0, 0},
		{"one", 1, 1},
		{"positive typical", 12345, 12345},
		{"negative one clamps", -1, 0},
		{"very negative clamps", math.MinInt, 0},
		{"MaxInt32 fits", int(math.MaxInt32), uint32(math.MaxInt32)},
		{"above MaxInt32 still fits", int(math.MaxInt32) + 1, uint32(math.MaxInt32) + 1},
		{"MaxUint32 exact", int(math.MaxUint32), uint32(math.MaxUint32)},
		{"above MaxUint32 clamps", int(math.MaxUint32) + 1, uint32(math.MaxUint32)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, SafeUint32(tt.input))
		})
	}
}

// TestSafeUintToInt covers both branches: MaxInt clamp and pass-through.
// The clamp case is deliberately expressed as "MaxUint > MaxInt" so the
// test's own assumption is checked at compile time.
func TestSafeUintToInt(t *testing.T) {
	tests := []struct {
		name  string
		input uint
		want  int
	}{
		{"zero", 0, 0},
		{"one", 1, 1},
		{"typical", 42, 42},
		{"MaxInt exact fits", uint(math.MaxInt), math.MaxInt},
		{"above MaxInt clamps", uint(math.MaxInt) + 1, math.MaxInt},
		{"MaxUint clamps", math.MaxUint, math.MaxInt},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, SafeUintToInt(tt.input))
		})
	}
}

// TestSafeIntToUint covers the "≤0 → 0" branch and the pass-through branch,
// including both zero and one as boundary cases either side of the branch.
func TestSafeIntToUint(t *testing.T) {
	tests := []struct {
		name  string
		input int
		want  uint
	}{
		{"zero yields zero", 0, 0},
		{"negative yields zero", -1, 0},
		{"very negative yields zero", math.MinInt, 0},
		{"one", 1, 1},
		{"typical", 42, 42},
		{"MaxInt pass-through", math.MaxInt, uint(math.MaxInt)},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, SafeIntToUint(tt.input))
		})
	}
}
