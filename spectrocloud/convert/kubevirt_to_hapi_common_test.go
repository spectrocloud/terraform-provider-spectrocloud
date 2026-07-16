package convert

import (
	"encoding/base64"
	"testing"
	"time"

	"github.com/go-openapi/strfmt"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

// TestToHapiVmOwnerReferences is the "back" leg of the round-trip already
// exercised in hapi_to_kubevirt_common_test.go; kept as its own test so
// coverage attributes correctly and so a break in one direction has a
// dedicated failure signal.
func TestToHapiVmOwnerReferences(t *testing.T) {
	block, ctrl := true, false

	in := []metav1.OwnerReference{
		{
			APIVersion:         "v1",
			Kind:               "Deployment",
			Name:               "web",
			UID:                types.UID("uid-1"),
			BlockOwnerDeletion: &block,
			Controller:         &ctrl,
		},
	}
	got := ToHapiVmOwnerReferences(in)
	require.Len(t, got, 1)
	assert.Equal(t, "v1", *got[0].APIVersion)
	assert.Equal(t, "Deployment", *got[0].Kind)
	assert.Equal(t, "web", *got[0].Name)
	assert.Equal(t, "uid-1", *got[0].UID)
	assert.True(t, got[0].BlockOwnerDeletion)
	assert.False(t, got[0].Controller)

	assert.Empty(t, ToHapiVmOwnerReferences(nil))
}

// TestToHapiVmManagedFields verifies (a) empty slice and (b) that the
// Fields payload survives ByteToStrfmtBase64 → StrfmtBase64ToByte
// round-trip.
func TestToHapiVmManagedFields(t *testing.T) {
	assert.Empty(t, ToHapiVmManagedFields(nil))

	raw := []byte(`{"spec":{"foo":"bar"}}`)
	in := []metav1.ManagedFieldsEntry{
		{
			APIVersion: "v1",
			FieldsType: "FieldsV1",
			FieldsV1:   &metav1.FieldsV1{Raw: raw},
			Manager:    "kubectl",
			Operation:  metav1.ManagedFieldsOperationApply,
		},
	}
	got := ToHapiVmManagedFields(in)
	require.Len(t, got, 1)
	assert.Equal(t, "v1", got[0].APIVersion)
	assert.Equal(t, "FieldsV1", got[0].FieldsType)
	assert.Equal(t, "kubectl", got[0].Manager)
	assert.Equal(t, string(metav1.ManagedFieldsOperationApply), got[0].Operation)
	require.NotNil(t, got[0].FieldsV1)
	require.Len(t, got[0].FieldsV1.Raw, 1)

	// Round-trip: encoded bytes decode back to the input.
	decoded, err := base64.StdEncoding.DecodeString(string(got[0].FieldsV1.Raw[0]))
	require.NoError(t, err)
	assert.Equal(t, raw, decoded)
}

func TestToHapiV1Time(t *testing.T) {
	// Any concrete time — the important invariant is v1.Time.Time round-trips
	// through the V1Time cast without shifting.
	// V1Time is a defined type over strfmt.DateTime which is time.Time, so we
	// unwrap by casting back through both layers.
	now := metav1.NewTime(time.Date(2025, 6, 1, 12, 0, 0, 0, time.UTC))
	got := ToHapiV1Time(now)
	assert.Equal(t, now.Time, time.Time(strfmt.DateTime(got)))
}

func TestToHapiVmFieldsV1(t *testing.T) {
	raw := []byte(`{"answer":42}`)
	got := ToHapiVmFieldsV1(&metav1.FieldsV1{Raw: raw})
	require.NotNil(t, got)
	require.Len(t, got.Raw, 1)

	decoded, err := base64.StdEncoding.DecodeString(string(got.Raw[0]))
	require.NoError(t, err)
	assert.Equal(t, raw, decoded)
}

func TestByteToStrfmtBase64(t *testing.T) {
	// Empty input still returns a slice with one entry — the empty
	// base64 string. That's intentional (aligns with the decode side,
	// which iterates a slice), so pin it.
	got := ByteToStrfmtBase64(nil)
	require.Len(t, got, 1)
	assert.Equal(t, strfmt.Base64(""), got[0])

	got = ByteToStrfmtBase64([]byte("hello"))
	require.Len(t, got, 1)
	assert.Equal(t, strfmt.Base64(base64.StdEncoding.EncodeToString([]byte("hello"))), got[0])
}

func TestToHapiVmQuantityDivisor(t *testing.T) {
	q, err := resource.ParseQuantity("1Gi")
	require.NoError(t, err)
	got := ToHapiVmQuantityDivisor(q)
	assert.Equal(t, models.V1VMQuantity("1Gi"), got)

	zero, err := resource.ParseQuantity("0")
	require.NoError(t, err)
	assert.Equal(t, models.V1VMQuantity("0"), ToHapiVmQuantityDivisor(zero))
}
