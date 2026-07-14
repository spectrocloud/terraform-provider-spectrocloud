package convert

import (
	"math"
	"testing"

	"github.com/stretchr/testify/assert"
)

// TestExpandInt64FromInterface covers each branch of the switch, plus the
// clamp path for the uint64 case. Testing the unexported helper (rather
// than only the public ToHapiVolume/expandDataVolumeMetadataToVM) lets us
// pin the numeric-overflow contract in isolation — that's the branch
// most likely to regress under a "just use int64(g)" refactor.
func TestExpandInt64FromInterface(t *testing.T) {
	tests := []struct {
		name string
		in   interface{}
		want int64
	}{
		{"nil default returns zero", nil, 0},
		{"unknown type returns zero", "not a number", 0},
		{"int typical", int(42), 42},
		{"int negative", int(-42), -42},
		{"int64 pass-through", int64(math.MaxInt64), math.MaxInt64},
		{"int32 widens", int32(math.MaxInt32), int64(math.MaxInt32)},
		{"uint64 under MaxInt64", uint64(100), 100},
		{"uint64 at MaxInt64 exact", uint64(math.MaxInt64), math.MaxInt64},
		{"uint64 above MaxInt64 clamps", uint64(math.MaxInt64) + 1, math.MaxInt64},
		{"uint64 MaxUint64 clamps", uint64(math.MaxUint64), math.MaxInt64},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, expandInt64FromInterface(tt.in))
		})
	}
}
