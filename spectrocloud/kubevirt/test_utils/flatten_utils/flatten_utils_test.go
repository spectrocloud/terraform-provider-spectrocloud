package flatten_utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGetBaseOutputForDataVolume(t *testing.T) {
	got := GetBaseOutputForDataVolume()
	assert.NotNil(t, got)
	// The fixture is a map — invoking it once here covers every line of
	// the constant literal.
	m, ok := got.(map[string]interface{})
	assert.True(t, ok)
	assert.NotEmpty(t, m)
}
