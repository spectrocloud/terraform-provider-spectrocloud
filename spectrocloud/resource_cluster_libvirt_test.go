package spectrocloud

import (
	"github.com/spectrocloud/hapi/models"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFlattenGpuDevice(t *testing.T) {
	// Test case 1: Empty input
	result := flattenGpuDevice(nil)
	assert.Empty(t, result, "Expected an empty result for nil input")

	// Test case 2: Non-empty input with valid GPU devices
	gpus := []*models.V1GPUDeviceSpec{
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
	}
	result = flattenGpuDevice(gpus)
	assert.Len(t, result, 2, "Expected 2 GPU devices in the result")
	assert.Equal(t, "GTX 1080", result[0].(map[string]interface{})["device_model"], "Unexpected device model")
	assert.Equal(t, "NVIDIA", result[0].(map[string]interface{})["vendor"], "Unexpected vendor")
	assert.Equal(t, map[string]string{"GTX 1080": "0x5678"}, result[0].(map[string]interface{})["addresses"], "Unexpected addresses")
	assert.Equal(t, "RX 570", result[1].(map[string]interface{})["device_model"], "Unexpected device model")
	assert.Equal(t, "AMD", result[1].(map[string]interface{})["vendor"], "Unexpected vendor")
	assert.Equal(t, map[string]string{"RX 570": "0xEFGH"}, result[1].(map[string]interface{})["addresses"], "Unexpected addresses")

}

func TestGetGPUDevices(t *testing.T) {
	// Test case 1: nil input
	result := getGPUDevices(nil)
	assert.Nil(t, result, "Expected nil result for nil input")

	// Test case 2: Empty input
	result = getGPUDevices([]interface{}{})
	assert.Empty(t, result, "Expected empty result for empty input")

	// Test case 3: Valid input with one GPU device
	gpuDevice := []interface{}{
		map[string]interface{}{
			"device_model": "GTX 1080",
			"vendor":       "NVIDIA",
			"addresses": map[string]interface{}{
				"address1": "0x1234",
				"address2": "0x5678",
			},
		},
	}
	result = getGPUDevices(gpuDevice)
	assert.Len(t, result, 1, "Expected 1 GPU device in the result")
	assert.Equal(t, "GTX 1080", result[0].Model, "Unexpected device model")
	assert.Equal(t, "NVIDIA", result[0].Vendor, "Unexpected vendor")
	assert.Len(t, result[0].Addresses, 2, "Unexpected number of addresses")
	assert.Equal(t, "0x1234", result[0].Addresses["address1"], "Unexpected address")
	assert.Equal(t, "0x5678", result[0].Addresses["address2"], "Unexpected address")

	// Test case 4: Valid input with multiple GPU devices
	gpuDevice = []interface{}{
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
	}
	result = getGPUDevices(gpuDevice)
	assert.Len(t, result, 2, "Expected 2 GPU devices in the result")
	assert.Equal(t, "RX 570", result[0].Model, "Unexpected device model")
	assert.Equal(t, "AMD", result[0].Vendor, "Unexpected vendor")
	assert.Empty(t, result[0].Addresses, "Unexpected addresses")
	assert.Equal(t, "GTX 2080", result[1].Model, "Unexpected device model")
	assert.Equal(t, "NVIDIA", result[1].Vendor, "Unexpected vendor")
	assert.Len(t, result[1].Addresses, 1, "Unexpected number of addresses")
	assert.Equal(t, "0xABCD", result[1].Addresses["address1"], "Unexpected address")
}
