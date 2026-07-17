package convert

import (
	"encoding/base64"
	"testing"

	"github.com/go-openapi/strfmt"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

// TestToKubevirtVM covers the guard clause (nil input → error) and the
// happy path (fields copied verbatim). These are the two paths — no need
// for a wider table.
func TestToKubevirtVM(t *testing.T) {
	t.Run("nil input errors", func(t *testing.T) {
		got, err := ToKubevirtVM(nil)
		require.Error(t, err)
		assert.Nil(t, got)
	})

	t.Run("copies all fields", func(t *testing.T) {
		in := &models.V1ClusterVirtualMachine{
			APIVersion: "kubevirt.io/v1",
			Kind:       "VirtualMachine",
			Metadata:   &models.V1VMObjectMeta{Name: "vm1"},
			Spec:       &models.V1ClusterVirtualMachineSpec{},
			Status:     &models.V1ClusterVirtualMachineStatus{Ready: true},
		}
		got, err := ToKubevirtVM(in)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, in.APIVersion, got.APIVersion)
		assert.Equal(t, in.Kind, got.Kind)
		assert.Same(t, in.Metadata, got.Metadata, "Metadata should be aliased, not deep-copied")
		assert.Same(t, in.Spec, got.Spec)
		assert.Same(t, in.Status, got.Status)
	})
}

// TestToKubevirtVMOwnerReferences round-trips an OwnerReference through
// ToKubevirtVMOwnerReferences → ToHapiVmOwnerReferences to verify the pair
// is symmetric. Fields APIVersion / Kind / Name / UID must survive; the
// BlockOwnerDeletion and Controller pointer semantics differ (HAPI: bare
// bool; k8s: *bool) so we assert on value not pointer identity.
func TestToKubevirtVMOwnerReferences(t *testing.T) {
	empty, err := ToKubevirtVMOwnerReferences(nil)
	require.NoError(t, err)
	assert.Empty(t, empty)

	in := []*models.V1VMOwnerReference{
		{
			APIVersion:         types.Ptr("v1"),
			BlockOwnerDeletion: true,
			Controller:         false,
			Kind:               types.Ptr("Deployment"),
			Name:               types.Ptr("web"),
			UID:                types.Ptr("uid-1"),
		},
		{
			APIVersion:         types.Ptr("v1"),
			BlockOwnerDeletion: false,
			Controller:         true,
			Kind:               types.Ptr("ReplicaSet"),
			Name:               types.Ptr("web-abc"),
			UID:                types.Ptr("uid-2"),
		},
	}
	got, err := ToKubevirtVMOwnerReferences(in)
	require.NoError(t, err)
	require.Len(t, got, 2)

	for i, ref := range got {
		assert.Equal(t, *in[i].APIVersion, ref.APIVersion)
		assert.Equal(t, *in[i].Kind, ref.Kind)
		assert.Equal(t, *in[i].Name, ref.Name)
		assert.Equal(t, string(ref.UID), *in[i].UID)
		require.NotNil(t, ref.BlockOwnerDeletion)
		require.NotNil(t, ref.Controller)
		assert.Equal(t, in[i].BlockOwnerDeletion, *ref.BlockOwnerDeletion)
		assert.Equal(t, in[i].Controller, *ref.Controller)
	}

	// Round-trip must be lossless for the fields it claims to copy.
	back := ToHapiVmOwnerReferences(got)
	require.Len(t, back, len(in))
	for i := range in {
		assert.Equal(t, *in[i].APIVersion, *back[i].APIVersion)
		assert.Equal(t, in[i].BlockOwnerDeletion, back[i].BlockOwnerDeletion)
		assert.Equal(t, in[i].Controller, back[i].Controller)
		assert.Equal(t, *in[i].Kind, *back[i].Kind)
		assert.Equal(t, *in[i].Name, *back[i].Name)
		assert.Equal(t, *in[i].UID, *back[i].UID)
	}
}

// TestToKubevirtVMManagedFields covers the happy path (FieldsV1 base64
// decodes cleanly) and the error path (an invalid base64 payload
// propagates the decode error). The former also transitively covers
// ToKvVmFieldsV1 and StrfmtBase64ToByte.
func TestToKubevirtVMManagedFields(t *testing.T) {
	empty, err := ToKubevirtVMManagedFields(nil)
	require.NoError(t, err)
	assert.Empty(t, empty)

	raw := []byte(`{"foo":"bar"}`)
	encoded := base64.StdEncoding.EncodeToString(raw)
	in := []*models.V1VMManagedFieldsEntry{
		{
			APIVersion: "v1",
			FieldsType: "FieldsV1",
			FieldsV1:   &models.V1VMFieldsV1{Raw: []strfmt.Base64{strfmt.Base64(encoded)}},
			Manager:    "kubectl",
			Operation:  "Apply",
		},
	}
	got, err := ToKubevirtVMManagedFields(in)
	require.NoError(t, err)
	require.Len(t, got, 1)
	assert.Equal(t, "v1", got[0].APIVersion)
	assert.Equal(t, "FieldsV1", got[0].FieldsType)
	assert.Equal(t, "kubectl", got[0].Manager)
	assert.Equal(t, metav1.ManagedFieldsOperationApply, got[0].Operation)
	require.NotNil(t, got[0].FieldsV1)
	assert.Equal(t, raw, got[0].FieldsV1.Raw)

	// A garbage base64 blob must bubble up as an error, not corrupt state
	// or panic during downstream consumption.
	bad := []*models.V1VMManagedFieldsEntry{
		{
			FieldsV1: &models.V1VMFieldsV1{Raw: []strfmt.Base64{strfmt.Base64("@@@not-valid-base64")}},
		},
	}
	_, err = ToKubevirtVMManagedFields(bad)
	require.Error(t, err)
}

// TestStrfmtBase64ToByte checks empty input (returns nil / empty, no
// panic), single value, and concatenation across multiple entries — plus
// the decode-error path.
func TestStrfmtBase64ToByte(t *testing.T) {
	got, err := StrfmtBase64ToByte(nil)
	require.NoError(t, err)
	assert.Empty(t, got)

	a := "hello "
	b := "world"
	encoded := []strfmt.Base64{
		strfmt.Base64(base64.StdEncoding.EncodeToString([]byte(a))),
		strfmt.Base64(base64.StdEncoding.EncodeToString([]byte(b))),
	}
	got, err = StrfmtBase64ToByte(encoded)
	require.NoError(t, err)
	assert.Equal(t, []byte(a+b), got, "consecutive entries should be concatenated in order")

	_, err = StrfmtBase64ToByte([]strfmt.Base64{strfmt.Base64("@@@not-valid")})
	require.Error(t, err)
}

// TestToKvVmManagedFieldsOperationType covers the three switch cases
// (Apply, Update, default). The default is important — an unknown
// operation must not panic, and the current implementation falls back to
// Apply, so pin that behavior.
func TestToKvVmManagedFieldsOperationType(t *testing.T) {
	assert.Equal(t, metav1.ManagedFieldsOperationApply, ToKvVmManagedFieldsOperationType("Apply"))
	assert.Equal(t, metav1.ManagedFieldsOperationUpdate, ToKvVmManagedFieldsOperationType("Update"))
	// Unknown → default to Apply. If this ever changes, tests should be
	// updated deliberately, not silently regressed.
	assert.Equal(t, metav1.ManagedFieldsOperationApply, ToKvVmManagedFieldsOperationType(""))
	assert.Equal(t, metav1.ManagedFieldsOperationApply, ToKvVmManagedFieldsOperationType("SomethingElse"))
}
