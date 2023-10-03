package spectrocloud

import (
	"testing"

	"github.com/spectrocloud/hapi/models"

	"github.com/stretchr/testify/assert"
)

func TestFlattenGpuDevice(t *testing.T) {
	tests := []struct {
		name        string
		input       []*models.V1GPUDeviceSpec
		expectedLen int
		validations []func(t *testing.T, result []interface{})
	}{
		{
			name:        "nil input",
			input:       nil,
			expectedLen: 0,
			validations: []func(t *testing.T, result []interface{}){
				func(t *testing.T, result []interface{}) {
					assert.Empty(t, result)
				},
			},
		},
		{
			name: "non-empty input with valid GPU devices",
			input: []*models.V1GPUDeviceSpec{
				{
					Model:     "GTX 1080",
					Vendor:    "NVIDIA",
					Addresses: map[string]string{"GTX 1080": "0x5678"},
				},
				{
					Model:     "RX 570",
					Vendor:    "AMD",
					Addresses: map[string]string{"RX 570": "0xEFGH"},
				},
			},
			expectedLen: 2,
			validations: []func(t *testing.T, result []interface{}){
				func(t *testing.T, result []interface{}) {
					assert.Equal(t, "GTX 1080", result[0].(map[string]interface{})["device_model"], "Unexpected device model")
					assert.Equal(t, "NVIDIA", result[0].(map[string]interface{})["vendor"], "Unexpected vendor")
					assert.Equal(t, map[string]string{"GTX 1080": "0x5678"}, result[0].(map[string]interface{})["addresses"], "Unexpected addresses")
					assert.Equal(t, "RX 570", result[1].(map[string]interface{})["device_model"], "Unexpected device model")
					assert.Equal(t, "AMD", result[1].(map[string]interface{})["vendor"], "Unexpected vendor")
					assert.Equal(t, map[string]string{"RX 570": "0xEFGH"}, result[1].(map[string]interface{})["addresses"], "Unexpected addresses")
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := flattenGpuDevice(tt.input)
			assert.Len(t, result, tt.expectedLen, "Unexpected number of GPU devices")
			for _, validate := range tt.validations {
				validate(t, result)
			}
		})
	}
}

func TestGetGPUDevices(t *testing.T) {
	tests := []struct {
		name        string
		input       interface{}
		expectedLen int
		validations []func(t *testing.T, result []*models.V1GPUDeviceSpec)
	}{
		{
			name:        "nil input",
			input:       nil,
			expectedLen: 0,
			validations: []func(t *testing.T, result []*models.V1GPUDeviceSpec){
				func(t *testing.T, result []*models.V1GPUDeviceSpec) {
					assert.Nil(t, result)
				},
			},
		},
		{
			name:        "empty input",
			input:       []interface{}{},
			expectedLen: 0,
			validations: []func(t *testing.T, result []*models.V1GPUDeviceSpec){
				func(t *testing.T, result []*models.V1GPUDeviceSpec) {
					assert.Empty(t, result)
				},
			},
		},
		{
			name: "valid input with one GPU device",
			input: []interface{}{
				map[string]interface{}{
					"device_model": "GTX 1080",
					"vendor":       "NVIDIA",
					"addresses": map[string]interface{}{
						"address1": "0x1234",
						"address2": "0x5678",
					},
				},
			},
			expectedLen: 1,
			validations: []func(t *testing.T, result []*models.V1GPUDeviceSpec){
				func(t *testing.T, result []*models.V1GPUDeviceSpec) {
					assert.Equal(t, "GTX 1080", result[0].Model)
					assert.Equal(t, "NVIDIA", result[0].Vendor)
					assert.Len(t, result[0].Addresses, 2)
					assert.Equal(t, "0x1234", result[0].Addresses["address1"])
					assert.Equal(t, "0x5678", result[0].Addresses["address2"])
				},
			},
		},
		{
			name: "valid input with multiple GPU devices",
			input: []interface{}{
				map[string]interface{}{
					"device_model": "RX 570",
					"vendor":       "AMD",
					"addresses":    map[string]interface{}{},
				},
				map[string]interface{}{
					"device_model": "GTX 2080",
					"vendor":       "NVIDIA",
					"addresses": map[string]interface{}{
						"address1": "0xABCD",
					},
				},
			},
			expectedLen: 2,
			validations: []func(t *testing.T, result []*models.V1GPUDeviceSpec){
				func(t *testing.T, result []*models.V1GPUDeviceSpec) {
					assert.Equal(t, "RX 570", result[0].Model)
					assert.Equal(t, "AMD", result[0].Vendor)
					assert.Empty(t, result[0].Addresses)
					assert.Equal(t, "GTX 2080", result[1].Model)
					assert.Equal(t, "NVIDIA", result[1].Vendor)
					assert.Len(t, result[1].Addresses, 1)
					assert.Equal(t, "0xABCD", result[1].Addresses["address1"])
				},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := getGPUDevices(tt.input)
			assert.Len(t, result, tt.expectedLen, "Unexpected number of GPU devices")
			for _, validate := range tt.validations {
				validate(t, result)
			}
		})
	}
}
