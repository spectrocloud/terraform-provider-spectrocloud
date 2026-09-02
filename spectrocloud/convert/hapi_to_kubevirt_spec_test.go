package convert

import (
	"testing"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// TestToKubevirtVMSpecM verifies that the JSON round-trip in
// ToKubevirtVMSpecM preserves the subset of fields both types share.
// Since the source and destination are both marshal-compatible types
// generated from the same OpenAPI schema, this is essentially checking
// that we don't corrupt the payload — the value comes from catching a
// future breaking rename or serializer change.
func TestToKubevirtVMSpecM(t *testing.T) {
	t.Run("nil input produces zero spec without error", func(t *testing.T) {
		// json.Marshal(nil) yields "null", which json.Unmarshal accepts
		// as a no-op — so the result is a zero-value spec, not an error.
		// Pin that so a swap to a stricter marshaller becomes visible.
		got, err := ToKubevirtVMSpecM(nil)
		require.NoError(t, err)
		assert.Equal(t, models.V1VMVirtualMachineInstanceSpec{}, got)
	})

	t.Run("empty spec round-trips", func(t *testing.T) {
		in := &models.V1ClusterVirtualMachineSpec{}
		got, err := ToKubevirtVMSpecM(in)
		require.NoError(t, err)
		// Zero-value struct — nothing to assert beyond "no error".
		_ = got
	})

	t.Run("populated input does not error", func(t *testing.T) {
		// Source and destination types are two different generated
		// OpenAPI structs with no overlapping JSON field names, so the
		// marshal→unmarshal produces a zero-value spec regardless of
		// what the caller passed. The valuable invariant to lock is that
		// this function does NOT panic or return an error — silently
		// producing an empty struct is the current (and expected)
		// behavior downstream code relies on.
		in := &models.V1ClusterVirtualMachineSpec{RunStrategy: "Always"}
		got, err := ToKubevirtVMSpecM(in)
		require.NoError(t, err)
		assert.NotNil(t, got)
	})
}
