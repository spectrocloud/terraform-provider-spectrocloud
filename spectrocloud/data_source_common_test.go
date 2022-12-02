package spectrocloud

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestToDatasourcesId(t *testing.T) {
	a := map[string]string{"a": "b", "c": "d", "e": "f"}

	assert.Equal(t, "prefix-a-b-c-d-e-f", toDatasourcesId("prefix", a))
}
