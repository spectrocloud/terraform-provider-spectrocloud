package k8s

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
)

// persistent_volume_claim_test.go — Batch 8.
// Round-trip tests for the 8 PVC helper funcs. The flatten helpers
// consume typed v1 structs; the expand helpers consume the schema-map
// shape. We verify both directions plus the schema builder itself.

func TestPersistentVolumeClaimSpecSchema(t *testing.T) {
	s := PersistentVolumeClaimSpecSchema()
	require.NotNil(t, s)
	assert.Equal(t, schema.TypeList, s.Type)
	assert.Equal(t, 1, s.MaxItems)

	res, ok := s.Elem.(*schema.Resource)
	require.True(t, ok)
	assert.Contains(t, res.Schema, "access_modes")
	assert.Contains(t, res.Schema, "resources")
	assert.Contains(t, res.Schema, "selector")
	assert.Contains(t, res.Schema, "volume_name")
	assert.Contains(t, res.Schema, "storage_class_name")
	assert.Contains(t, res.Schema, "volume_mode")
}

func TestPersistentVolumeClaimSpecFieldsBuilder(t *testing.T) {
	f := persistentVolumeClaimSpecFields()
	assert.Contains(t, f, "access_modes")
	assert.Contains(t, f, "resources")
	assert.Contains(t, f, "selector")
	assert.Contains(t, f, "volume_name")
	assert.Contains(t, f, "storage_class_name")
	assert.Contains(t, f, "volume_mode")
}

func TestExpandFlattenPersistentVolumeAccessModes(t *testing.T) {
	// Round-trip.
	in := []interface{}{"ReadWriteOnce", "ReadOnlyMany"}
	got := expandPersistentVolumeAccessModes(in)
	require.Len(t, got, 2)
	assert.Equal(t, v1.PersistentVolumeAccessMode("ReadWriteOnce"), got[0])

	// Flatten returns a schema.Set (unordered) — check membership.
	set := flattenPersistentVolumeAccessModes(got)
	require.NotNil(t, set)
	assert.ElementsMatch(t, []interface{}{"ReadWriteOnce", "ReadOnlyMany"}, set.List())
}

func TestExpandResourceRequirements(t *testing.T) {
	t.Run("empty", func(t *testing.T) {
		got, err := expandResourceRequirements(nil)
		require.NoError(t, err)
		require.NotNil(t, got)
	})

	t.Run("populated", func(t *testing.T) {
		got, err := expandResourceRequirements([]interface{}{
			map[string]interface{}{
				"limits":   map[string]interface{}{"storage": "10Gi"},
				"requests": map[string]interface{}{"storage": "5Gi"},
			},
		})
		require.NoError(t, err)
		assert.NotNil(t, got.Limits)
		assert.NotNil(t, got.Requests)
	})

	t.Run("invalid quantity in limits", func(t *testing.T) {
		_, err := expandResourceRequirements([]interface{}{
			map[string]interface{}{"limits": map[string]interface{}{"storage": "not-a-quantity"}},
		})
		assert.Error(t, err)
	})
}

func TestFlattenResourceRequirements(t *testing.T) {
	// Empty in → empty map, still returns []interface{} of len 1.
	got := flattenResourceRequirements(v1.ResourceRequirements{})
	require.Len(t, got, 1)

	got = flattenResourceRequirements(v1.ResourceRequirements{
		Limits: v1.ResourceList{
			v1.ResourceStorage: resource.MustParse("10Gi"),
		},
		Requests: v1.ResourceList{
			v1.ResourceStorage: resource.MustParse("5Gi"),
		},
	})
	require.Len(t, got, 1)
	m := got[0].(map[string]interface{})
	assert.Contains(t, m, "limits")
	assert.Contains(t, m, "requests")
}

func TestExpandPersistentVolumeClaimSpec(t *testing.T) {
	t.Run("empty input", func(t *testing.T) {
		got, err := ExpandPersistentVolumeClaimSpec(nil)
		require.NoError(t, err)
		require.NotNil(t, got)
	})

	t.Run("fully populated", func(t *testing.T) {
		got, err := ExpandPersistentVolumeClaimSpec([]interface{}{
			map[string]interface{}{
				"access_modes": schema.NewSet(schema.HashString, []interface{}{"ReadWriteOnce"}),
				"resources": []interface{}{
					map[string]interface{}{"requests": map[string]interface{}{"storage": "5Gi"}},
				},
				"selector":           []interface{}{}, // absent
				"volume_name":        "pv-1",
				"storage_class_name": "standard",
				"volume_mode":        "Filesystem",
			},
		})
		require.NoError(t, err)
		require.NotNil(t, got)
		require.Len(t, got.AccessModes, 1)
		assert.Equal(t, "pv-1", got.VolumeName)
		require.NotNil(t, got.StorageClassName)
		assert.Equal(t, "standard", *got.StorageClassName)
		require.NotNil(t, got.VolumeMode)
		assert.Equal(t, v1.PersistentVolumeFilesystem, *got.VolumeMode)
	})

	t.Run("Block volume mode", func(t *testing.T) {
		got, err := ExpandPersistentVolumeClaimSpec([]interface{}{
			map[string]interface{}{
				"access_modes": schema.NewSet(schema.HashString, []interface{}{"ReadWriteOnce"}),
				"resources":    []interface{}{},
				"volume_mode":  "Block",
			},
		})
		require.NoError(t, err)
		require.NotNil(t, got.VolumeMode)
		assert.Equal(t, v1.PersistentVolumeBlock, *got.VolumeMode)
	})

	t.Run("invalid volume mode rejected", func(t *testing.T) {
		_, err := ExpandPersistentVolumeClaimSpec([]interface{}{
			map[string]interface{}{
				"access_modes": schema.NewSet(schema.HashString, []interface{}{"ReadWriteOnce"}),
				"resources":    []interface{}{},
				"volume_mode":  "SomethingElse",
			},
		})
		assert.Error(t, err)
	})
}

func TestFlattenPersistentVolumeClaimSpec(t *testing.T) {
	sc := "standard"
	vm := v1.PersistentVolumeFilesystem

	got := FlattenPersistentVolumeClaimSpec(v1.PersistentVolumeClaimSpec{
		AccessModes: []v1.PersistentVolumeAccessMode{v1.ReadWriteOnce},
		Resources: v1.VolumeResourceRequirements{
			Requests: v1.ResourceList{v1.ResourceStorage: resource.MustParse("5Gi")},
		},
		Selector:         &metav1.LabelSelector{MatchLabels: map[string]string{"app": "web"}},
		VolumeName:       "pv-1",
		StorageClassName: &sc,
		VolumeMode:       &vm,
	})
	require.Len(t, got, 1)
	m := got[0].(map[string]interface{})
	assert.Contains(t, m, "access_modes")
	assert.Contains(t, m, "resources")
	assert.Contains(t, m, "selector")
	assert.Equal(t, "pv-1", m["volume_name"])
	assert.Equal(t, "standard", m["storage_class_name"])
	// volume_mode stored as v1.PersistentVolumeMode, not string — the
	// schema layer normalizes to string; here we get the raw typed value.
	assert.EqualValues(t, "Filesystem", m["volume_mode"])
}
