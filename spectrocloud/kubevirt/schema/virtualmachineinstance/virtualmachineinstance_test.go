package virtualmachineinstance

import (
	"testing"

	"github.com/stretchr/testify/assert"
	v1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/api/resource"
	kubevirtapiv1 "kubevirt.io/api/core/v1"
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
					"user_data_base64":    "",
					"user_data":           "",
					"network_data_base64": "",
					"network_data":        "",
				},
			},
		},
		{
			name: "Valid input with user_data only",
			input: kubevirtapiv1.CloudInitNoCloudSource{
				UserData: "user-data-content",
			},
			expected: []interface{}{
				map[string]interface{}{
					"user_data_base64":    "",
					"user_data":           "user-data-content",
					"network_data_base64": "",
					"network_data":        "",
				},
			},
		},
		{
			name: "Valid input with user_data and network_data",
			input: kubevirtapiv1.CloudInitNoCloudSource{
				UserData:    "user-data-content",
				NetworkData: "network-data-content",
			},
			expected: []interface{}{
				map[string]interface{}{
					"user_data_base64":    "",
					"user_data":           "user-data-content",
					"network_data_base64": "",
					"network_data":        "network-data-content",
				},
			},
		},
		{
			name: "Valid input with base64 encoded data",
			input: kubevirtapiv1.CloudInitNoCloudSource{
				UserDataBase64:    "dXNlci1kYXRhLWNvbnRlbnQ=",
				NetworkDataBase64: "bmV0d29yay1kYXRhLWNvbnRlbnQ=",
			},
			expected: []interface{}{
				map[string]interface{}{
					"user_data_base64":    "dXNlci1kYXRhLWNvbnRlbnQ=",
					"user_data":           "",
					"network_data_base64": "bmV0d29yay1kYXRhLWNvbnRlbnQ=",
					"network_data":        "",
				},
			},
		},
		{
			name: "Valid input with all fields",
			input: kubevirtapiv1.CloudInitNoCloudSource{
				UserData:          "user-data-content",
				UserDataBase64:    "dXNlci1kYXRhLWNvbnRlbnQ=",
				NetworkData:       "network-data-content",
				NetworkDataBase64: "bmV0d29yay1kYXRhLWNvbnRlbnQ=",
			},
			expected: []interface{}{
				map[string]interface{}{
					"user_data_base64":    "dXNlci1kYXRhLWNvbnRlbnQ=",
					"user_data":           "user-data-content",
					"network_data_base64": "bmV0d29yay1kYXRhLWNvbnRlbnQ=",
					"network_data":        "network-data-content",
				},
			},
		},
		{
			name: "Valid input with secret references",
			input: kubevirtapiv1.CloudInitNoCloudSource{
				UserDataSecretRef: &v1.LocalObjectReference{
					Name: "my-user-secret",
				},
				NetworkDataSecretRef: &v1.LocalObjectReference{
					Name: "my-network-secret",
				},
			},
			expected: []interface{}{
				map[string]interface{}{
					"user_data_secret_ref": []interface{}{
						map[string]interface{}{
							"name": "my-user-secret",
						},
					},
					"user_data_base64": "",
					"user_data":        "",
					"network_data_secret_ref": []interface{}{
						map[string]interface{}{
							"name": "my-network-secret",
						},
					},
					"network_data_base64": "",
					"network_data":        "",
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

func TestExpandCloudInitNoCloud(t *testing.T) {
	tests := []struct {
		name     string
		input    []interface{}
		expected *kubevirtapiv1.CloudInitNoCloudSource
	}{
		{
			name:     "Empty input",
			input:    []interface{}{},
			expected: nil,
		},
		{
			name:     "Nil input",
			input:    nil,
			expected: nil,
		},
		{
			name: "Valid input with user_data only",
			input: []interface{}{
				map[string]interface{}{
					"user_data": "user-data-content",
				},
			},
			expected: &kubevirtapiv1.CloudInitNoCloudSource{
				UserData: "user-data-content",
			},
		},
		{
			name: "Valid input with user_data and network_data",
			input: []interface{}{
				map[string]interface{}{
					"user_data":    "user-data-content",
					"network_data": "network-data-content",
				},
			},
			expected: &kubevirtapiv1.CloudInitNoCloudSource{
				UserData:    "user-data-content",
				NetworkData: "network-data-content",
			},
		},
		{
			name: "Valid input with base64 encoded data",
			input: []interface{}{
				map[string]interface{}{
					"user_data_base64":    "dXNlci1kYXRhLWNvbnRlbnQ=",
					"network_data_base64": "bmV0d29yay1kYXRhLWNvbnRlbnQ=",
				},
			},
			expected: &kubevirtapiv1.CloudInitNoCloudSource{
				UserDataBase64:    "dXNlci1kYXRhLWNvbnRlbnQ=",
				NetworkDataBase64: "bmV0d29yay1kYXRhLWNvbnRlbnQ=",
			},
		},
		{
			name: "Valid input with all string fields",
			input: []interface{}{
				map[string]interface{}{
					"user_data":           "user-data-content",
					"user_data_base64":    "dXNlci1kYXRhLWNvbnRlbnQ=",
					"network_data":        "network-data-content",
					"network_data_base64": "bmV0d29yay1kYXRhLWNvbnRlbnQ=",
				},
			},
			expected: &kubevirtapiv1.CloudInitNoCloudSource{
				UserData:          "user-data-content",
				UserDataBase64:    "dXNlci1kYXRhLWNvbnRlbnQ=",
				NetworkData:       "network-data-content",
				NetworkDataBase64: "bmV0d29yay1kYXRhLWNvbnRlbnQ=",
			},
		},
		{
			name: "Valid input with secret references",
			input: []interface{}{
				map[string]interface{}{
					"user_data_secret_ref": []interface{}{
						map[string]interface{}{
							"name": "my-user-secret",
						},
					},
					"network_data_secret_ref": []interface{}{
						map[string]interface{}{
							"name": "my-network-secret",
						},
					},
				},
			},
			expected: &kubevirtapiv1.CloudInitNoCloudSource{
				UserDataSecretRef: &v1.LocalObjectReference{
					Name: "my-user-secret",
				},
				NetworkDataSecretRef: &v1.LocalObjectReference{
					Name: "my-network-secret",
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := expandCloudInitNoCloud(tt.input)
			assert.Equal(t, tt.expected, result)
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
