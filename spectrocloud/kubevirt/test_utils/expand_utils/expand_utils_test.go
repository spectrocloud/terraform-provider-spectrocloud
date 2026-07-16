package expand_utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// expand_utils_test.go — smoke coverage for the fixture builders. These
// funcs are large hardcoded map literals used across other test suites;
// calling them once here covers every line.

func TestGetBaseInputForDataVolume(t *testing.T) {
	got := GetBaseInputForDataVolume()
	assert.NotNil(t, got)
	// The fixture is []interface{} — a schema.TypeList shape. We only
	// assert non-emptiness to keep the fixture free to evolve.
	s, ok := got.([]interface{})
	assert.True(t, ok)
	assert.NotEmpty(t, s)
}

func TestGetBaseInputForVirtualMachine(t *testing.T) {
	got := GetBaseInputForVirtualMachine()
	assert.NotNil(t, got)
	// VM fixture is a top-level map, not a slice.
	m, ok := got.(map[string]interface{})
	assert.True(t, ok)
	assert.NotEmpty(t, m)
}
