package datavolume

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/spectrocloud/terraform-provider-spectrocloud/types"
)

func TestDataVolumeSpecFields(t *testing.T) {
	fields := dataVolumeSpecFields()
	assert.NotNil(t, fields)
	assert.Contains(t, fields, "source")
	assert.Contains(t, fields, "pvc")
	assert.Contains(t, fields, "storage") // New field
	assert.Contains(t, fields, "content_type")
}

func TestDataVolumeStorageSchema(t *testing.T) {
	storageSchema := dataVolumeStorageSchema()
	assert.NotNil(t, storageSchema)
	assert.Equal(t, schema.TypeList, storageSchema.Type)
	assert.True(t, storageSchema.Optional)
	assert.Equal(t, 1, storageSchema.MaxItems)

	// Check storage schema fields
	resource := storageSchema.Elem.(*schema.Resource)
	assert.Contains(t, resource.Schema, "access_modes")
	assert.Contains(t, resource.Schema, "resources")
	assert.Contains(t, resource.Schema, "selector")
	assert.Contains(t, resource.Schema, "volume_name")
	assert.Contains(t, resource.Schema, "storage_class_name")
	assert.Contains(t, resource.Schema, "volume_mode")
}

func TestExpandDataVolumeStorage_Full(t *testing.T) {
	input := []interface{}{
		map[string]interface{}{
			"access_modes": schema.NewSet(schema.HashString, []interface{}{
				"ReadWriteOnce",
				"ReadOnlyMany",
			}),
			"resources": []interface{}{
				map[string]interface{}{
					"requests": map[string]interface{}{
						"storage": "10Gi",
					},
					"limits": map[string]interface{}{
						"storage": "20Gi",
					},
				},
			},
			"selector": []interface{}{
				map[string]interface{}{
					"match_labels": map[string]interface{}{
						"type": "ssd",
						"tier": "premium",
					},
					"match_expressions": []interface{}{
						map[string]interface{}{
							"key":      "zone",
							"operator": "In",
							"values": schema.NewSet(schema.HashString, []interface{}{
								"us-west-1a",
								"us-west-1b",
							}),
						},
					},
				},
			},
			"volume_name":        "test-volume",
			"storage_class_name": "premium-ssd",
			"volume_mode":        "Filesystem",
		},
	}

	result, err := expandDataVolumeStorage(input)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Check access modes ([]string in HAPI model, not []PersistentVolumeAccessMode)
	assert.Len(t, result.AccessModes, 2)
	assert.Contains(t, result.AccessModes, string(v1.ReadWriteOnce))
	assert.Contains(t, result.AccessModes, string(v1.ReadOnlyMany))

	// Check resources (map[string]models.V1VMQuantity, not core/v1.ResourceList)
	require.NotNil(t, result.Resources)
	requestStorage := result.Resources.Requests["storage"]
	limitStorage := result.Resources.Limits["storage"]
	assert.Equal(t, "10Gi", string(requestStorage))
	assert.Equal(t, "20Gi", string(limitStorage))

	// Check selector
	assert.NotNil(t, result.Selector)
	assert.Equal(t, "ssd", result.Selector.MatchLabels["type"])
	assert.Equal(t, "premium", result.Selector.MatchLabels["tier"])
	assert.Len(t, result.Selector.MatchExpressions, 1)
	require.NotNil(t, result.Selector.MatchExpressions[0].Key)
	require.NotNil(t, result.Selector.MatchExpressions[0].Operator)
	assert.Equal(t, "zone", *result.Selector.MatchExpressions[0].Key)
	assert.Equal(t, string(metav1.LabelSelectorOpIn), *result.Selector.MatchExpressions[0].Operator)
	assert.Contains(t, result.Selector.MatchExpressions[0].Values, "us-west-1a")
	assert.Contains(t, result.Selector.MatchExpressions[0].Values, "us-west-1b")

	// Check other fields
	assert.Equal(t, "test-volume", result.VolumeName)
	assert.Equal(t, "premium-ssd", result.StorageClassName)
	assert.Equal(t, string(v1.PersistentVolumeFilesystem), result.VolumeMode)
}

func TestExpandDataVolumeStorage_Minimal(t *testing.T) {
	input := []interface{}{
		map[string]interface{}{
			"access_modes": schema.NewSet(schema.HashString, []interface{}{
				"ReadWriteOnce",
			}),
			"resources": []interface{}{
				map[string]interface{}{
					"requests": map[string]interface{}{
						"storage": "5Gi",
					},
				},
			},
		},
	}

	result, err := expandDataVolumeStorage(input)
	require.NoError(t, err)
	require.NotNil(t, result)

	// Check minimal configuration
	assert.Len(t, result.AccessModes, 1)
	assert.Equal(t, string(v1.ReadWriteOnce), result.AccessModes[0])

	require.NotNil(t, result.Resources)
	requestStorage := result.Resources.Requests["storage"]
	assert.Equal(t, "5Gi", string(requestStorage))

	// Optional fields should be nil/empty
	assert.Nil(t, result.Selector)
	assert.Empty(t, result.VolumeName)
	assert.Empty(t, result.StorageClassName)
	assert.Empty(t, result.VolumeMode)
}

func TestExpandDataVolumeStorage_Empty(t *testing.T) {
	tests := []struct {
		name  string
		input []interface{}
	}{
		{"nil input", nil},
		{"empty slice", []interface{}{}},
		{"nil element", []interface{}{nil}},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := expandDataVolumeStorage(tt.input)
			assert.NoError(t, err)
			assert.Nil(t, result)
		})
	}
}

func TestExpandDataVolumeStorage_VolumeMode(t *testing.T) {
	tests := []struct {
		input    string
		expected *v1.PersistentVolumeMode
	}{
		{"Block", types.Ptr(v1.PersistentVolumeBlock)},
		{"Filesystem", types.Ptr(v1.PersistentVolumeFilesystem)},
		{"", nil},
	}

	for _, tt := range tests {
		t.Run(tt.input, func(t *testing.T) {
			input := []interface{}{
				map[string]interface{}{
					"volume_mode": tt.input,
				},
			}

			result, err := expandDataVolumeStorage(input)
			require.NoError(t, err)
			require.NotNil(t, result)

			if tt.expected == nil {
				assert.Empty(t, result.VolumeMode)
			} else {
				assert.Equal(t, string(*tt.expected), result.VolumeMode)
			}
		})
	}
}

func TestFlattenDataVolumeStorage_Full(t *testing.T) {
	input := &models.V1VMStorageSpec{
		AccessModes: []string{string(v1.ReadWriteOnce), string(v1.ReadOnlyMany)},
		Resources: &models.V1VMCoreResourceRequirements{
			Requests: map[string]models.V1VMQuantity{"storage": "10Gi"},
			Limits:   map[string]models.V1VMQuantity{"storage": "20Gi"},
		},
		Selector: &models.V1VMLabelSelector{
			MatchLabels: map[string]string{
				"type": "ssd",
				"tier": "premium",
			},
			MatchExpressions: []*models.V1VMLabelSelectorRequirement{
				{
					Key:      types.Ptr("zone"),
					Operator: types.Ptr(string(metav1.LabelSelectorOpIn)),
					Values:   []string{"us-west-1a", "us-west-1b"},
				},
			},
		},
		VolumeName:       "test-volume",
		StorageClassName: "premium-ssd",
		VolumeMode:       string(v1.PersistentVolumeFilesystem),
	}

	result := flattenDataVolumeStorageFromVM(input)
	require.Len(t, result, 1)

	flattened := result[0].(map[string]interface{})

	// Check access modes
	accessModes := flattened["access_modes"].(*schema.Set)
	assert.Equal(t, 2, accessModes.Len())
	assert.True(t, accessModes.Contains("ReadWriteOnce"))
	assert.True(t, accessModes.Contains("ReadOnlyMany"))

	// Check resources
	resources := flattened["resources"].([]interface{})
	require.Len(t, resources, 1)
	resourceMap := resources[0].(map[string]interface{})
	requests := resourceMap["requests"].(map[string]interface{})
	limits := resourceMap["limits"].(map[string]interface{})
	assert.Equal(t, "10Gi", requests["storage"])
	assert.Equal(t, "20Gi", limits["storage"])

	// Check selector
	selector := flattened["selector"].([]interface{})
	require.Len(t, selector, 1)
	selectorMap := selector[0].(map[string]interface{})
	matchLabels := selectorMap["match_labels"].(map[string]interface{})
	assert.Equal(t, "ssd", matchLabels["type"])
	assert.Equal(t, "premium", matchLabels["tier"])

	matchExpressions := selectorMap["match_expressions"].([]interface{})
	require.Len(t, matchExpressions, 1)
	expr := matchExpressions[0].(map[string]interface{})
	assert.Equal(t, "zone", expr["key"])
	assert.Equal(t, string(metav1.LabelSelectorOpIn), expr["operator"])

	// Check other fields
	assert.Equal(t, "test-volume", flattened["volume_name"])
	assert.Equal(t, "premium-ssd", flattened["storage_class_name"])
	assert.Equal(t, "Filesystem", flattened["volume_mode"])
}

func TestFlattenDataVolumeStorage_Minimal(t *testing.T) {
	input := &models.V1VMStorageSpec{
		AccessModes: []string{string(v1.ReadWriteOnce)},
		Resources: &models.V1VMCoreResourceRequirements{
			Requests: map[string]models.V1VMQuantity{"storage": "5Gi"},
		},
	}

	result := flattenDataVolumeStorageFromVM(input)
	require.Len(t, result, 1)

	flattened := result[0].(map[string]interface{})

	// Check access modes
	accessModes := flattened["access_modes"].(*schema.Set)
	assert.Equal(t, 1, accessModes.Len())
	assert.True(t, accessModes.Contains("ReadWriteOnce"))

	// Check resources
	resources := flattened["resources"].([]interface{})
	require.Len(t, resources, 1)
	resourceMap := resources[0].(map[string]interface{})
	requests := resourceMap["requests"].(map[string]interface{})
	assert.Equal(t, "5Gi", requests["storage"])

	// Optional fields should not be present or be empty
	assert.NotContains(t, flattened, "selector")
	assert.NotContains(t, flattened, "volume_name")
	assert.NotContains(t, flattened, "storage_class_name")
	assert.NotContains(t, flattened, "volume_mode")
}

func TestExpandDataVolumeSpec_WithStorage(t *testing.T) {
	input := []interface{}{
		map[string]interface{}{
			"source": []interface{}{
				map[string]interface{}{
					"blank": []interface{}{
						map[string]interface{}{},
					},
				},
			},
			"storage": []interface{}{
				map[string]interface{}{
					"access_modes": schema.NewSet(schema.HashString, []interface{}{
						"ReadWriteOnce",
					}),
					"resources": []interface{}{
						map[string]interface{}{
							"requests": map[string]interface{}{
								"storage": "10Gi",
							},
						},
					},
					"storage_class_name": "fast-ssd",
				},
			},
			"content_type": "kubevirt",
		},
	}

	result, err := ExpandDataVolumeSpec(input)
	require.NoError(t, err)

	// Check that storage field is populated
	assert.NotNil(t, result.Storage)
	assert.Len(t, result.Storage.AccessModes, 1)
	assert.Equal(t, string(v1.ReadWriteOnce), result.Storage.AccessModes[0])
	assert.Equal(t, "fast-ssd", result.Storage.StorageClassName)

	// Check that PVC is nil when using storage
	assert.Nil(t, result.Pvc)

	// Check other fields
	assert.NotNil(t, result.Source)
	assert.Equal(t, "kubevirt", result.ContentType)
}

func TestExpandDataVolumeSpec_WithPVC(t *testing.T) {
	input := []interface{}{
		map[string]interface{}{
			"source": []interface{}{
				map[string]interface{}{
					"blank": []interface{}{
						map[string]interface{}{},
					},
				},
			},
			"pvc": []interface{}{
				map[string]interface{}{
					"access_modes": schema.NewSet(schema.HashString, []interface{}{
						"ReadWriteOnce",
					}),
					"resources": []interface{}{
						map[string]interface{}{
							"requests": map[string]interface{}{
								"storage": "10Gi",
							},
						},
					},
					"storage_class_name": "standard",
				},
			},
			"content_type": "archive",
		},
	}

	result, err := ExpandDataVolumeSpec(input)
	require.NoError(t, err)

	// Check that PVC field is populated
	assert.NotNil(t, result.Pvc)
	assert.Len(t, result.Pvc.AccessModes, 1)
	assert.Equal(t, string(v1.ReadWriteOnce), result.Pvc.AccessModes[0])
	assert.Equal(t, "standard", result.Pvc.StorageClassName)

	// Check that Storage is nil when using PVC
	assert.Nil(t, result.Storage)

	// Check other fields
	assert.NotNil(t, result.Source)
	assert.Equal(t, "archive", result.ContentType)
}

func TestExpandDataVolumeSpec_WithBothPVCAndStorage(t *testing.T) {
	input := []interface{}{
		map[string]interface{}{
			"source": []interface{}{
				map[string]interface{}{
					"blank": []interface{}{
						map[string]interface{}{},
					},
				},
			},
			"pvc": []interface{}{
				map[string]interface{}{
					"access_modes": schema.NewSet(schema.HashString, []interface{}{
						"ReadWriteOnce",
					}),
					"resources": []interface{}{
						map[string]interface{}{
							"requests": map[string]interface{}{
								"storage": "10Gi",
							},
						},
					},
				},
			},
			"storage": []interface{}{
				map[string]interface{}{
					"access_modes": schema.NewSet(schema.HashString, []interface{}{
						"ReadWriteMany",
					}),
					"resources": []interface{}{
						map[string]interface{}{
							"requests": map[string]interface{}{
								"storage": "20Gi",
							},
						},
					},
				},
			},
		},
	}

	result, err := ExpandDataVolumeSpec(input)
	require.NoError(t, err)

	// Both should be populated if both are provided
	assert.NotNil(t, result.Pvc)
	assert.NotNil(t, result.Storage)

	// Check PVC
	assert.Len(t, result.Pvc.AccessModes, 1)
	assert.Equal(t, string(v1.ReadWriteOnce), result.Pvc.AccessModes[0])

	// Check Storage
	assert.Len(t, result.Storage.AccessModes, 1)
	assert.Equal(t, string(v1.ReadWriteMany), result.Storage.AccessModes[0])
}

func TestFlattenDataVolumeSpec_WithStorage(t *testing.T) {
	input := &models.V1VMDataVolumeSpec{
		Source: &models.V1VMDataVolumeSource{},
		Storage: &models.V1VMStorageSpec{
			AccessModes: []string{string(v1.ReadWriteOnce)},
			Resources: &models.V1VMCoreResourceRequirements{
				Requests: map[string]models.V1VMQuantity{"storage": "10Gi"},
			},
			StorageClassName: "fast-ssd",
		},
		ContentType: "kubevirt",
	}

	result := FlattenDataVolumeSpecFromVM(input)
	require.Len(t, result, 1)

	flattened := result[0].(map[string]interface{})

	// Check that storage field is present
	assert.Contains(t, flattened, "storage")
	storage := flattened["storage"].([]interface{})
	require.Len(t, storage, 1)

	storageMap := storage[0].(map[string]interface{})
	assert.Equal(t, "fast-ssd", storageMap["storage_class_name"])

	// Check that PVC is not present when Storage is used
	assert.NotContains(t, flattened, "pvc")

	// Check other fields
	assert.Contains(t, flattened, "source")
	assert.Equal(t, "kubevirt", flattened["content_type"])
}

func TestFlattenDataVolumeSpec_WithPVC(t *testing.T) {
	input := &models.V1VMDataVolumeSpec{
		Source: &models.V1VMDataVolumeSource{},
		Pvc: &models.V1VMPersistentVolumeClaimSpec{
			AccessModes: []string{string(v1.ReadWriteOnce)},
			Resources: &models.V1VMCoreResourceRequirements{
				Requests: map[string]models.V1VMQuantity{"storage": "10Gi"},
			},
			StorageClassName: "standard",
		},
		ContentType: "archive",
	}

	result := FlattenDataVolumeSpecFromVM(input)
	require.Len(t, result, 1)

	flattened := result[0].(map[string]interface{})

	// Check that PVC field is present
	assert.Contains(t, flattened, "pvc")
	pvc := flattened["pvc"].([]interface{})
	require.Len(t, pvc, 1)

	// Check that storage is not present when PVC is used
	assert.NotContains(t, flattened, "storage")

	// Check other fields
	assert.Contains(t, flattened, "source")
	assert.Equal(t, "archive", flattened["content_type"])
}

func TestFlattenDataVolumeSpec_WithBothPVCAndStorage(t *testing.T) {
	input := &models.V1VMDataVolumeSpec{
		Source: &models.V1VMDataVolumeSource{},
		Pvc: &models.V1VMPersistentVolumeClaimSpec{
			AccessModes: []string{string(v1.ReadWriteOnce)},
			Resources: &models.V1VMCoreResourceRequirements{
				Requests: map[string]models.V1VMQuantity{"storage": "10Gi"},
			},
		},
		Storage: &models.V1VMStorageSpec{
			AccessModes: []string{string(v1.ReadWriteMany)},
			Resources: &models.V1VMCoreResourceRequirements{
				Requests: map[string]models.V1VMQuantity{"storage": "10Gi"},
			},
		},
	}

	result := FlattenDataVolumeSpecFromVM(input)
	require.Len(t, result, 1)

	flattened := result[0].(map[string]interface{})

	// Both should be present if both are in the spec
	assert.Contains(t, flattened, "pvc")
	assert.Contains(t, flattened, "storage")
}

func TestFlattenDataVolumeSpec_WithBlankSource(t *testing.T) {
	input := &models.V1VMDataVolumeSpec{
		Source: &models.V1VMDataVolumeSource{
			Blank: map[string]interface{}{},
		},
		Pvc: &models.V1VMPersistentVolumeClaimSpec{
			AccessModes: []string{string(v1.ReadWriteOnce)},
			Resources: &models.V1VMCoreResourceRequirements{
				Requests: map[string]models.V1VMQuantity{"storage": "5Gi"},
			},
		},
	}

	result := FlattenDataVolumeSpecFromVM(input)
	require.Len(t, result, 1)
	flattened := result[0].(map[string]interface{})
	source := flattened["source"].([]interface{})
	require.Len(t, source, 1)
	srcMap := source[0].(map[string]interface{})
	blank := srcMap["blank"].([]interface{})
	require.Len(t, blank, 1)
	assert.Equal(t, map[string]interface{}{}, blank[0].(map[string]interface{}))
}
