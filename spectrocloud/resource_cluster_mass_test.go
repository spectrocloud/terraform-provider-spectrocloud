package spectrocloud

import (
	"reflect"
	"testing"

	"github.com/spectrocloud/hapi/models"
)

func TestFlattenMachinePoolConfigsMaas(t *testing.T) {
	// Test scenario with nil input
	t.Run("Nil Input", func(t *testing.T) {
		expected := make([]interface{}, 0)
		result := flattenMachinePoolConfigsMaas(nil)

		if !reflect.DeepEqual(expected, result) {
			t.Errorf("Expected empty array for nil input, got: %v", result)
		}
	})

	// Test scenario with valid input
	t.Run("Valid Input", func(t *testing.T) {
		// Create a mock machine pool configuration
		mockMachinePool := []*models.V1MaasMachinePoolConfig{
			// Populate with your mock data
		}

		expected := []interface{}{
			// Populate with the expected output based on your mock data
		}

		result := flattenMachinePoolConfigsMaas(mockMachinePool)

		if !reflect.DeepEqual(expected, result) {
			t.Errorf("Expected output doesn't match actual result.\nExpected: %v\nGot: %v", expected, result)
		}
	})
}
