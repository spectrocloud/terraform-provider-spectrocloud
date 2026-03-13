package datavolume

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cdiv1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"

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

	// Check access modes
	assert.Len(t, result.AccessModes, 2)
	assert.Contains(t, result.AccessModes, v1.ReadWriteOnce)
	assert.Contains(t, result.AccessModes, v1.ReadOnlyMany)

	// Check resources
	requestStorage := result.Resources.Requests[v1.ResourceStorage]
	limitStorage := result.Resources.Limits[v1.ResourceStorage]
	expectedRequest, _ := resource.ParseQuantity("10Gi")
	expectedLimit, _ := resource.ParseQuantity("20Gi")
	assert.True(t, requestStorage.Equal(expectedRequest))
	assert.True(t, limitStorage.Equal(expectedLimit))

	// Check selector
	assert.NotNil(t, result.Selector)
	assert.Equal(t, "ssd", result.Selector.MatchLabels["type"])
	assert.Equal(t, "premium", result.Selector.MatchLabels["tier"])
	assert.Len(t, result.Selector.MatchExpressions, 1)
	assert.Equal(t, "zone", result.Selector.MatchExpressions[0].Key)
	assert.Equal(t, metav1.LabelSelectorOpIn, result.Selector.MatchExpressions[0].Operator)
	assert.Contains(t, result.Selector.MatchExpressions[0].Values, "us-west-1a")
	assert.Contains(t, result.Selector.MatchExpressions[0].Values, "us-west-1b")

	// Check other fields
	assert.Equal(t, "test-volume", result.VolumeName)
	assert.Equal(t, "premium-ssd", *result.StorageClassName)
	assert.Equal(t, v1.PersistentVolumeFilesystem, *result.VolumeMode)
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
	assert.Equal(t, v1.ReadWriteOnce, result.AccessModes[0])

	requestStorage := result.Resources.Requests[v1.ResourceStorage]
	expectedRequest, _ := resource.ParseQuantity("5Gi")
	assert.True(t, requestStorage.Equal(expectedRequest))

	// Optional fields should be nil/empty
	assert.Nil(t, result.Selector)
	assert.Empty(t, result.VolumeName)
	assert.Nil(t, result.StorageClassName)
	assert.Nil(t, result.VolumeMode)
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
				assert.Nil(t, result.VolumeMode)
			} else {
				assert.Equal(t, *tt.expected, *result.VolumeMode)
			}
		})
	}
}

func TestFlattenDataVolumeStorage_Full(t *testing.T) {
	requestStorage, _ := resource.ParseQuantity("10Gi")
	limitStorage, _ := resource.ParseQuantity("20Gi")

	input := cdiv1.StorageSpec{
		AccessModes: []v1.PersistentVolumeAccessMode{
			v1.ReadWriteOnce,
			v1.ReadOnlyMany,
		},
		Resources: v1.VolumeResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceStorage: requestStorage,
			},
			Limits: v1.ResourceList{
				v1.ResourceStorage: limitStorage,
			},
		},
		Selector: &metav1.LabelSelector{
			MatchLabels: map[string]string{
				"type": "ssd",
				"tier": "premium",
			},
			MatchExpressions: []metav1.LabelSelectorRequirement{
				{
					Key:      "zone",
					Operator: metav1.LabelSelectorOpIn,
					Values:   []string{"us-west-1a", "us-west-1b"},
				},
			},
		},
		VolumeName:       "test-volume",
		StorageClassName: types.Ptr("premium-ssd"),
		VolumeMode:       types.Ptr(v1.PersistentVolumeFilesystem),
	}

	result := flattenDataVolumeStorage(input)
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
	assert.Equal(t, metav1.LabelSelectorOpIn, expr["operator"])

	// Check other fields
	assert.Equal(t, "test-volume", flattened["volume_name"])
	assert.Equal(t, "premium-ssd", flattened["storage_class_name"])
	assert.Equal(t, "Filesystem", flattened["volume_mode"])
}

func TestFlattenDataVolumeStorage_Minimal(t *testing.T) {
	requestStorage, _ := resource.ParseQuantity("5Gi")

	input := cdiv1.StorageSpec{
		AccessModes: []v1.PersistentVolumeAccessMode{
			v1.ReadWriteOnce,
		},
		Resources: v1.VolumeResourceRequirements{
			Requests: v1.ResourceList{
				v1.ResourceStorage: requestStorage,
			},
		},
	}

	result := flattenDataVolumeStorage(input)
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
	assert.Equal(t, v1.ReadWriteOnce, result.Storage.AccessModes[0])
	assert.Equal(t, "fast-ssd", *result.Storage.StorageClassName)

	// Check that PVC is nil when using storage
	assert.Nil(t, result.PVC)

	// Check other fields
	assert.NotNil(t, result.Source)
	assert.Equal(t, cdiv1.DataVolumeContentType("kubevirt"), result.ContentType)
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
	assert.NotNil(t, result.PVC)
	assert.Len(t, result.PVC.AccessModes, 1)
	assert.Equal(t, v1.ReadWriteOnce, result.PVC.AccessModes[0])
	assert.Equal(t, "standard", *result.PVC.StorageClassName)

	// Check that Storage is nil when using PVC
	assert.Nil(t, result.Storage)

	// Check other fields
	assert.NotNil(t, result.Source)
	assert.Equal(t, cdiv1.DataVolumeContentType("archive"), result.ContentType)
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
	assert.NotNil(t, result.PVC)
	assert.NotNil(t, result.Storage)

	// Check PVC
	assert.Len(t, result.PVC.AccessModes, 1)
	assert.Equal(t, v1.ReadWriteOnce, result.PVC.AccessModes[0])

	// Check Storage
	assert.Len(t, result.Storage.AccessModes, 1)
	assert.Equal(t, v1.ReadWriteMany, result.Storage.AccessModes[0])
}

func TestFlattenDataVolumeSpec_WithStorage(t *testing.T) {
	requestStorage, _ := resource.ParseQuantity("10Gi")

	input := cdiv1.DataVolumeSpec{
		Source: &cdiv1.DataVolumeSource{
			Blank: &cdiv1.DataVolumeBlankImage{},
		},
		Storage: &cdiv1.StorageSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.ReadWriteOnce,
			},
			Resources: v1.VolumeResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceStorage: requestStorage,
				},
			},
			StorageClassName: types.Ptr("fast-ssd"),
		},
		ContentType: cdiv1.DataVolumeContentType("kubevirt"),
	}

	result := FlattenDataVolumeSpec(input)
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
	requestStorage, _ := resource.ParseQuantity("10Gi")

	input := cdiv1.DataVolumeSpec{
		Source: &cdiv1.DataVolumeSource{
			Blank: &cdiv1.DataVolumeBlankImage{},
		},
		PVC: &v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.ReadWriteOnce,
			},
			Resources: v1.VolumeResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceStorage: requestStorage,
				},
			},
			StorageClassName: types.Ptr("standard"),
		},
		ContentType: cdiv1.DataVolumeContentType("archive"),
	}

	result := FlattenDataVolumeSpec(input)
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
	requestStorage, _ := resource.ParseQuantity("10Gi")

	input := cdiv1.DataVolumeSpec{
		Source: &cdiv1.DataVolumeSource{
			Blank: &cdiv1.DataVolumeBlankImage{},
		},
		PVC: &v1.PersistentVolumeClaimSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.ReadWriteOnce,
			},
			Resources: v1.VolumeResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceStorage: requestStorage,
				},
			},
		},
		Storage: &cdiv1.StorageSpec{
			AccessModes: []v1.PersistentVolumeAccessMode{
				v1.ReadWriteMany,
			},
			Resources: v1.VolumeResourceRequirements{
				Requests: v1.ResourceList{
					v1.ResourceStorage: requestStorage,
				},
			},
		},
	}

	result := FlattenDataVolumeSpec(input)
	require.Len(t, result, 1)

	flattened := result[0].(map[string]interface{})

	// Both should be present if both are in the spec
	assert.Contains(t, flattened, "pvc")
	assert.Contains(t, flattened, "storage")
}
