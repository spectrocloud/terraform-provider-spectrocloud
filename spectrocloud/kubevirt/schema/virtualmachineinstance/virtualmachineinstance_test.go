package virtualmachineinstance

import (
	"github.com/stretchr/testify/assert"
	"k8s.io/apimachinery/pkg/api/resource"
	kubevirtapiv1 "kubevirt.io/api/core/v1"
	"testing"
)

func TestExpandProbe(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected *kubevirtapiv1.Probe
	}{
		{
			name:     "Empty input",
			input:    []interface{}{},
			expected: nil,
		},
		{
			name: "Nil input",
			input: []interface{}{
				nil,
			},
			expected: nil,
		},
		{
			name: "Valid input",
			input: []interface{}{
				map[string]interface{}{
					// Add key-value pairs for your Probe fields here
				},
			},
			expected: &kubevirtapiv1.Probe{
				// Fill in the expected Probe fields
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandProbe(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestFlattenProbe(t *testing.T) {
	tests := []struct {
		name     string
		input    kubevirtapiv1.Probe
		expected []interface{}
	}{
		{
			name:     "Empty input",
			input:    kubevirtapiv1.Probe{},
			expected: []interface{}{map[string]interface{}{}},
		},
		{
			name:  "Valid input",
			input: kubevirtapiv1.Probe{
				// Fill in the Probe fields
			},
			expected: []interface{}{
				map[string]interface{}{
					// Add key-value pairs for your Probe fields here
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flattenProbe(tt.input)
			assert.ElementsMatch(t, tt.expected, result)
		})
	}
}

func TestFlattenContainerDisk(t *testing.T) {
	tests := []struct {
		name     string
		input    kubevirtapiv1.ContainerDiskSource
		expected []interface{}
	}{
		{
			name: "Empty input",
			input: kubevirtapiv1.ContainerDiskSource{
				Image: "",
			},
			expected: []interface{}{
				map[string]interface{}{
					"image_url": "",
				},
			},
		},
		{
			name: "Valid input",
			input: kubevirtapiv1.ContainerDiskSource{
				Image: "registry.example.com/my-image",
			},
			expected: []interface{}{
				map[string]interface{}{
					"image_url": "registry.example.com/my-image",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flattenContainerDisk(tt.input)
			assert.ElementsMatch(t, tt.expected, result)
		})
	}
}

func TestFlattenCloudInitNoCloud(t *testing.T) {
	tests := []struct {
		name     string
		input    kubevirtapiv1.CloudInitNoCloudSource
		expected []interface{}
	}{
		{
			name: "Empty input",
			input: kubevirtapiv1.CloudInitNoCloudSource{
				UserData: "",
			},
			expected: []interface{}{
				map[string]interface{}{
					"user_data": "",
				},
			},
		},
		{
			name: "Valid input",
			input: kubevirtapiv1.CloudInitNoCloudSource{
				UserData: "user-data-content",
			},
			expected: []interface{}{
				map[string]interface{}{
					"user_data": "user-data-content",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flattenCloudInitNoCloud(tt.input)
			assert.ElementsMatch(t, tt.expected, result)
		})
	}
}

func TestFlattenEphemeral(t *testing.T) {
	tests := []struct {
		name     string
		input    kubevirtapiv1.EphemeralVolumeSource
		expected []interface{}
	}{
		{
			name: "Empty input",
			input: kubevirtapiv1.EphemeralVolumeSource{
				PersistentVolumeClaim: nil,
			},
			expected: []interface{}{
				map[string]interface{}{
					"persistent_volume_claim": nil,
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flattenEphemeral(tt.input)

		})
	}
}

func TestFlattenEmptyDisk(t *testing.T) {
	tests := []struct {
		name     string
		input    kubevirtapiv1.EmptyDiskSource
		expected []interface{}
	}{
		{
			name: "Empty input",
			input: kubevirtapiv1.EmptyDiskSource{
				Capacity: resource.Quantity{},
			},
			expected: []interface{}{
				map[string]interface{}{
					"capacity": "",
				},
			},
		},
		{
			name: "Valid input",
			input: kubevirtapiv1.EmptyDiskSource{
				Capacity: resource.Quantity{},
			},
			expected: []interface{}{
				map[string]interface{}{
					"capacity": "10Gi",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flattenEmptyDisk(tt.input)
		})
	}
}

func TestFlattenConfigMap(t *testing.T) {
	tests := []struct {
		name     string
		input    kubevirtapiv1.ConfigMapVolumeSource
		expected []interface{}
	}{
		{
			name:  "Empty input",
			input: kubevirtapiv1.ConfigMapVolumeSource{},
			expected: []interface{}{
				map[string]interface{}{
					"name": "",
				},
			},
		},
		{
			name:  "Valid input",
			input: kubevirtapiv1.ConfigMapVolumeSource{},
			expected: []interface{}{
				map[string]interface{}{
					"name": "my-config-map",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			flattenConfigMap(tt.input)
		})
	}
}
