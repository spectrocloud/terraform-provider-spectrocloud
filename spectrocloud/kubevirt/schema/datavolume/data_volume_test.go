package datavolume

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
)

func TestDataVolumeFields(t *testing.T) {
	fields := DataVolumeFields()
	assert.NotNil(t, fields)
	assert.Contains(t, fields, "cluster_uid")
	assert.Contains(t, fields, "cluster_context")
	assert.Contains(t, fields, "vm_name")
	assert.Contains(t, fields, "vm_namespace")
	assert.Contains(t, fields, "add_volume_options")
	assert.Contains(t, fields, "metadata")
	assert.Contains(t, fields, "spec")
	assert.Contains(t, fields, "status")
}

// func TestFromResourceData(t *testing.T) {
// 	resourceData := schema.TestResourceDataRaw(t, DataVolumeFields(), map[string]interface{}{
// 		"metadata": []interface{}{
// 			map[string]interface{}{
// 				"name":      "test-dv",
// 				"namespace": "default",
// 			},
// 		},
// 		"spec": []interface{}{
// 			map[string]interface{}{
// 				"source": map[string]interface{}{
// 					"blank": map[string]interface{}{},
// 				},
// 			},
// 		},
// 	})

// 	dataVolume, err := FromResourceData(resourceData)
// 	assert.NoError(t, err)
// 	assert.NotNil(t, dataVolume)
// 	assert.Equal(t, "test-dv", dataVolume.Name)
// 	assert.Equal(t, "default", dataVolume.Namespace)
// }

func TestToResourceData(t *testing.T) {
	tpl := &models.V1VMDataVolumeTemplateSpec{
		Spec: &models.V1VMDataVolumeSpec{
			PriorityClassName: "test-pirioirty",
			ContentType:       "bil",
			Checkpoints:       nil,
			FinalCheckpoint:   false,
			Preallocation:     false,
		},
	}
	resourceData := schema.TestResourceDataRaw(t, DataVolumeFields(), map[string]interface{}{})
	err := ToResourceDataFromVMTemplate(tpl, resourceData)
	assert.NoError(t, err)
}

// func TestExpandDataVolumeStatus(t *testing.T) {
// 	tests := []struct {
// 		input    []interface{}
// 		expected cdiv1.DataVolumeStatus
// 	}{
// 		{
// 			input: []interface{}{
// 				map[string]interface{}{
// 					"phase":    "Succeeded",
// 					"progress": "50%",
// 				},
// 			},
// 			expected: cdiv1.DataVolumeStatus{
// 				Phase:    cdiv1.DataVolumePhase("Succeeded"),
// 				Progress: cdiv1.DataVolumeProgress("50%"),
// 			},
// 		},
// 		{
// 			input:    []interface{}{nil},
// 			expected: cdiv1.DataVolumeStatus{},
// 		},
// 		{
// 			input:    []interface{}{},
// 			expected: cdiv1.DataVolumeStatus{},
// 		},
// 		{
// 			input: []interface{}{
// 				map[string]interface{}{
// 					"phase": "Failed",
// 				},
// 			},
// 			expected: cdiv1.DataVolumeStatus{
// 				Phase: cdiv1.DataVolumePhase("Failed"),
// 			},
// 		},
// 	}

// 	for _, tt := range tests {
// 		t.Run("", func(t *testing.T) {
// 			result := expandDataVolumeStatus(tt.input)
// 			assert.Equal(t, tt.expected, result)
// 		})
// 	}
// }
