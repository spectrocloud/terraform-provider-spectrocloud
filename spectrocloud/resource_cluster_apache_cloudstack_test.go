package spectrocloud

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestResourceMachinePoolApacheCloudStackHash(t *testing.T) {
	testCases := []struct {
		name     string
		input    interface{}
		expected int
	}{
		{
			name: "Complete CloudStack machine pool with all fields",
			input: map[string]interface{}{
				"name":  "cloudstack-pool-1",
				"count": 3,
				"additional_labels": map[string]interface{}{
					"env":  "production",
					"team": "platform",
				},
				"additional_annotations": map[string]interface{}{
					"custom.io/annotation1":   "value1",
					"company.com/annotation2": "value2",
				},
				"offering": "medium-instance",
				"template": []interface{}{
					map[string]interface{}{
						"id":   "template-123",
						"name": "ubuntu-20.04",
					},
				},
				"network": []interface{}{
					map[string]interface{}{
						"network_name": "network-1",
					},
				},
				"control_plane":           false,
				"control_plane_as_worker": false,
				"update_strategy":         "RollingUpdateScaleOut",
				"node_repave_interval":    0,
			},
			expected: 0, // Will be calculated in test
		},
		{
			name: "Minimal CloudStack machine pool",
			input: map[string]interface{}{
				"name":                    "cloudstack-pool-2",
				"count":                   2,
				"offering":                "small-instance",
				"control_plane":           true,
				"control_plane_as_worker": false,
			},
			expected: 0, // Will be calculated in test
		},
		{
			name: "CloudStack machine pool with annotations only",
			input: map[string]interface{}{
				"name":  "cloudstack-pool-3",
				"count": 1,
				"additional_annotations": map[string]interface{}{
					"annotation1": "value1",
				},
				"offering":                "medium-instance",
				"control_plane":           false,
				"control_plane_as_worker": false,
			},
			expected: 0, // Will be calculated in test
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			actual := resourceMachinePoolApacheCloudStackHash(tc.input)
			// For first run, just ensure hash is generated
			assert.NotEqual(t, 0, actual, "Hash should not be zero")
		})
	}
}

func TestResourceMachinePoolApacheCloudStackHashAnnotationChangeDetection(t *testing.T) {
	// Base machine pool without annotations
	baseMachinePool := map[string]interface{}{
		"name":  "test-pool",
		"count": 3,
		"additional_labels": map[string]interface{}{
			"label1": "value1",
		},
		"offering":                "medium-instance",
		"control_plane":           false,
		"control_plane_as_worker": false,
	}

	// Machine pool with annotations
	withAnnotations := map[string]interface{}{
		"name":  "test-pool",
		"count": 3,
		"additional_labels": map[string]interface{}{
			"label1": "value1",
		},
		"additional_annotations": map[string]interface{}{
			"annotation1": "value1",
		},
		"offering":                "medium-instance",
		"control_plane":           false,
		"control_plane_as_worker": false,
	}

	// Machine pool with different annotations
	differentAnnotations := map[string]interface{}{
		"name":  "test-pool",
		"count": 3,
		"additional_labels": map[string]interface{}{
			"label1": "value1",
		},
		"additional_annotations": map[string]interface{}{
			"annotation1": "value2",
		},
		"offering":                "medium-instance",
		"control_plane":           false,
		"control_plane_as_worker": false,
	}

	// Machine pool with additional annotations
	moreAnnotations := map[string]interface{}{
		"name":  "test-pool",
		"count": 3,
		"additional_labels": map[string]interface{}{
			"label1": "value1",
		},
		"additional_annotations": map[string]interface{}{
			"annotation1": "value1",
			"annotation2": "value2",
		},
		"offering":                "medium-instance",
		"control_plane":           false,
		"control_plane_as_worker": false,
	}

	baseHash := resourceMachinePoolApacheCloudStackHash(baseMachinePool)
	withAnnotationsHash := resourceMachinePoolApacheCloudStackHash(withAnnotations)
	differentAnnotationsHash := resourceMachinePoolApacheCloudStackHash(differentAnnotations)
	moreAnnotationsHash := resourceMachinePoolApacheCloudStackHash(moreAnnotations)

	// Hash should be different when annotations are added
	assert.NotEqual(t, baseHash, withAnnotationsHash, "Adding annotations should change hash")

	// Hash should be different when annotation values change
	assert.NotEqual(t, withAnnotationsHash, differentAnnotationsHash, "Changing annotation values should change hash")

	// Hash should be different when adding more annotations
	assert.NotEqual(t, withAnnotationsHash, moreAnnotationsHash, "Adding more annotations should change hash")

	// Hash should be consistent for same input
	sameHash := resourceMachinePoolApacheCloudStackHash(withAnnotations)
	assert.Equal(t, withAnnotationsHash, sameHash, "Same input should produce same hash")
}

func TestResourceMachinePoolApacheCloudStackHashAllFields(t *testing.T) {
	testCases := []struct {
		name        string
		baseInput   map[string]interface{}
		modifyField func(map[string]interface{})
		description string
	}{
		{
			name: "Offering change affects hash",
			baseInput: map[string]interface{}{
				"name":                    "pool-1",
				"count":                   2,
				"offering":                "small-instance",
				"control_plane":           false,
				"control_plane_as_worker": false,
			},
			modifyField: func(m map[string]interface{}) {
				m["offering"] = "medium-instance"
			},
			description: "Changing offering should change hash",
		},
		{
			name: "Template change affects hash",
			baseInput: map[string]interface{}{
				"name":     "pool-1",
				"count":    2,
				"offering": "small-instance",
				"template": []interface{}{
					map[string]interface{}{
						"id":   "template-123",
						"name": "ubuntu-20.04",
					},
				},
				"control_plane":           false,
				"control_plane_as_worker": false,
			},
			modifyField: func(m map[string]interface{}) {
				m["template"] = []interface{}{
					map[string]interface{}{
						"id":   "template-456",
						"name": "ubuntu-22.04",
					},
				}
			},
			description: "Changing template should change hash",
		},
		{
			name: "Network change affects hash",
			baseInput: map[string]interface{}{
				"name":     "pool-1",
				"count":    2,
				"offering": "small-instance",
				"network": []interface{}{
					map[string]interface{}{
						"network_name": "network-1",
					},
				},
				"control_plane":           false,
				"control_plane_as_worker": false,
			},
			modifyField: func(m map[string]interface{}) {
				m["network"] = []interface{}{
					map[string]interface{}{
						"network_name": "network-2",
					},
				}
			},
			description: "Changing network should change hash",
		},
		{
			name: "Annotation change affects hash",
			baseInput: map[string]interface{}{
				"name":     "pool-1",
				"count":    2,
				"offering": "small-instance",
				"additional_annotations": map[string]interface{}{
					"annotation1": "value1",
				},
				"control_plane":           false,
				"control_plane_as_worker": false,
			},
			modifyField: func(m map[string]interface{}) {
				m["additional_annotations"] = map[string]interface{}{
					"annotation1": "value2",
				}
			},
			description: "Changing annotations should change hash",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Get hash of base input
			baseHash := resourceMachinePoolApacheCloudStackHash(tc.baseInput)

			// Create modified copy
			modified := copyMap(tc.baseInput)
			tc.modifyField(modified)

			// Get hash of modified input
			modifiedHash := resourceMachinePoolApacheCloudStackHash(modified)

			// Hashes should be different
			if baseHash == modifiedHash {
				t.Errorf("%s: Base hash %d equals modified hash %d, but they should differ.\nBase: %+v\nModified: %+v",
					tc.description, baseHash, modifiedHash, tc.baseInput, modified)
			}
		})
	}
}

func TestResourceMachinePoolApacheCloudStackHashKubernetesStyleAnnotations(t *testing.T) {
	machinePool := map[string]interface{}{
		"name":  "test-pool",
		"count": 2,
		"additional_annotations": map[string]interface{}{
			"custom.io/annotation":        "value1",
			"company.com/another-annot":   "value2",
			"kubernetes.io/some-metadata": "metadata-value",
		},
		"offering":                "medium-instance",
		"control_plane":           false,
		"control_plane_as_worker": false,
	}

	// Should generate a valid hash with Kubernetes-style annotations
	hash := resourceMachinePoolApacheCloudStackHash(machinePool)
	assert.NotEqual(t, 0, hash, "Hash should be generated for Kubernetes-style annotations")

	// Hash should be consistent
	sameHash := resourceMachinePoolApacheCloudStackHash(machinePool)
	assert.Equal(t, hash, sameHash, "Same input should produce same hash")
}
