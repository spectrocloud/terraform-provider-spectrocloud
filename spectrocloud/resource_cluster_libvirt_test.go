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
