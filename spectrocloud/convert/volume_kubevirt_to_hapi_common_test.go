package convert

import (
	"math"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/spectrocloud/terraform-provider-spectrocloud/spectrocloud/kubevirt/schema/datavolume"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
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

func TestExpandDataVolumeMetadataToVM(t *testing.T) {
	t.Run("empty metadata returns empty object meta", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, datavolume.DataVolumeFields(), map[string]interface{}{
			"cluster_context": "project",
		})
		got := expandDataVolumeMetadataToVM(d)
		require.NotNil(t, got)
		assert.Empty(t, got.Name)
	})

	t.Run("populated metadata", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, datavolume.DataVolumeFields(), map[string]interface{}{
			"cluster_context": "project",
			"metadata": []interface{}{
				map[string]interface{}{
					"name":      "boot-vol",
					"namespace": "default",
					"labels": map[string]interface{}{
						"app": "demo",
					},
					"annotations": map[string]interface{}{
						"note": "test",
					},
				},
			},
		})
		got := expandDataVolumeMetadataToVM(d)
		require.NotNil(t, got)
		assert.Equal(t, "boot-vol", got.Name)
		assert.Equal(t, "default", got.Namespace)
		assert.Equal(t, map[string]string{"app": "demo"}, got.Labels)
		assert.Equal(t, map[string]string{"note": "test"}, got.Annotations)
	})
}

func TestToHapiVolume(t *testing.T) {
	t.Run("minimal input succeeds", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, datavolume.DataVolumeFields(), map[string]interface{}{
			"cluster_context": "project",
			"metadata": []interface{}{
				map[string]interface{}{
					"name":      "boot-vol",
					"namespace": "default",
				},
			},
		})
		opts := &models.V1VMAddVolumeOptions{}
		got, err := ToHapiVolume(d, opts)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Same(t, opts, got.AddVolumeOptions)
		assert.True(t, got.Persist)
		require.NotNil(t, got.DataVolumeTemplate)
		require.NotNil(t, got.DataVolumeTemplate.Metadata)
		assert.Equal(t, "boot-vol", got.DataVolumeTemplate.Metadata.Name)
		require.NotNil(t, got.DataVolumeTemplate.Spec)
	})

	t.Run("populated spec", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, datavolume.DataVolumeFields(), map[string]interface{}{
			"cluster_context": "project",
			"metadata": []interface{}{
				map[string]interface{}{
					"name":      "boot-vol",
					"namespace": "default",
				},
			},
			"spec": []interface{}{
				map[string]interface{}{
					"content_type": "kubevirt",
					"source": []interface{}{
						map[string]interface{}{
							"blank": []interface{}{
								map[string]interface{}{},
							},
						},
					},
				},
			},
		})
		got, err := ToHapiVolume(d, nil)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Nil(t, got.AddVolumeOptions)
		require.NotNil(t, got.DataVolumeTemplate.Spec)
		assert.Equal(t, "kubevirt", got.DataVolumeTemplate.Spec.ContentType)
	})
}
