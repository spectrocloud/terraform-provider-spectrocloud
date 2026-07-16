package convert

import (
	"testing"

	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

// TestToKubevirtVMStatusM checks the same JSON-marshal round-trip that
// ToKubevirtVMSpecM does, but for the Status type. Same rationale — pins
// the current implementation so an accidental schema drift is loud.
func TestToKubevirtVMStatusM(t *testing.T) {
	t.Run("nil input round-trips through null", func(t *testing.T) {
		got, err := ToKubevirtVMStatusM(nil)
		require.NoError(t, err)
		assert.Equal(t, models.V1ClusterVirtualMachineStatus{}, got)
	})

	t.Run("ready+created flags preserved", func(t *testing.T) {
		in := &models.V1ClusterVirtualMachineStatus{
			Ready:              true,
			Created:            true,
			SnapshotInProgress: "snap-1",
			RestoreInProgress:  "restore-1",
			PrintableStatus:    "Running",
		}
		got, err := ToKubevirtVMStatusM(in)
		require.NoError(t, err)
		assert.Equal(t, in.Ready, got.Ready)
		assert.Equal(t, in.Created, got.Created)
		assert.Equal(t, in.SnapshotInProgress, got.SnapshotInProgress)
		assert.Equal(t, in.RestoreInProgress, got.RestoreInProgress)
		assert.Equal(t, in.PrintableStatus, got.PrintableStatus)
	})
}

// TestToKubevirtVMStatus exercises the shallow-copy variant. Unlike the
// *M form it has no JSON marshalling, so field-by-field aliasing is
// the contract to lock in — a future refactor to deep-copy would need
// the caller side reviewed, and this test would flag it.
func TestToKubevirtVMStatus(t *testing.T) {
	t.Run("nil input returns zero struct", func(t *testing.T) {
		got := ToKubevirtVMStatus(nil)
		assert.Equal(t, models.V1ClusterVirtualMachineStatus{}, got)
	})

	t.Run("populated status is copied field-by-field", func(t *testing.T) {
		in := &models.V1ClusterVirtualMachineStatus{
			SnapshotInProgress:  "snap-1",
			RestoreInProgress:   "restore-1",
			Created:             true,
			Ready:               true,
			PrintableStatus:     "Running",
			Conditions:          []*models.V1VMVirtualMachineCondition{{Type: types.Ptr("Ready"), Status: types.Ptr("True")}},
			StateChangeRequests: []*models.V1VMVirtualMachineStateChangeRequest{{Action: types.Ptr("Start")}},
		}
		got := ToKubevirtVMStatus(in)
		assert.Equal(t, in.SnapshotInProgress, got.SnapshotInProgress)
		assert.Equal(t, in.RestoreInProgress, got.RestoreInProgress)
		assert.Equal(t, in.Created, got.Created)
		assert.Equal(t, in.Ready, got.Ready)
		assert.Equal(t, in.PrintableStatus, got.PrintableStatus)
		// Slices are shared by reference (shallow copy) — that's the current
		// contract; assert on element identity so a future deep-copy shows
		// up as a deliberate diff here.
		require.Len(t, got.Conditions, 1)
		assert.Same(t, in.Conditions[0], got.Conditions[0])
		require.Len(t, got.StateChangeRequests, 1)
		assert.Same(t, in.StateChangeRequests[0], got.StateChangeRequests[0])
	})
}

// TestToKvVmStatusConditions covers the trivial map — empty, one entry,
// many — and TestToKvVmStatusCondition covers the nil-guard + field copy.
func TestToKvVmStatusConditions(t *testing.T) {
	assert.Nil(t, ToKvVmStatusConditions(nil), "nil input yields nil result")
	assert.Nil(t, ToKvVmStatusConditions([]*models.V1VMVirtualMachineCondition{}), "empty input yields nil (append semantics)")

	in := []*models.V1VMVirtualMachineCondition{
		{Type: types.Ptr("Ready"), Status: types.Ptr("True"), Reason: "AllUp"},
		{Type: types.Ptr("Paused"), Status: types.Ptr("False")},
	}
	got := ToKvVmStatusConditions(in)
	require.Len(t, got, 2)
	require.NotNil(t, got[0].Type)
	require.NotNil(t, got[0].Status)
	assert.Equal(t, "Ready", *got[0].Type)
	assert.Equal(t, "True", *got[0].Status)
	assert.Equal(t, "AllUp", got[0].Reason)
	require.NotNil(t, got[1].Type)
	assert.Equal(t, "Paused", *got[1].Type)
}

func TestToKvVmStatusCondition(t *testing.T) {
	// nil in → zero out; asserts the nil-guard rather than the field copies.
	assert.Equal(t, models.V1VMVirtualMachineCondition{}, ToKvVmStatusCondition(nil))

	in := &models.V1VMVirtualMachineCondition{
		Type:    types.Ptr("Ready"),
		Status:  types.Ptr("True"),
		Reason:  "Started",
		Message: "vm is up",
	}
	got := ToKvVmStatusCondition(in)
	// Type/Status are *string — assert via the pointer contents.
	require.NotNil(t, got.Type)
	require.NotNil(t, got.Status)
	assert.Equal(t, *in.Type, *got.Type)
	assert.Equal(t, *in.Status, *got.Status)
	assert.Equal(t, in.Reason, got.Reason)
	assert.Equal(t, in.Message, got.Message)
}
