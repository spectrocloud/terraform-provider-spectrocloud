package datavolume

import (
	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	cdiv1 "kubevirt.io/containerized-data-importer-api/pkg/apis/core/v1beta1"
	"testing"
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

func TestFromResourceData(t *testing.T) {
	resourceData := schema.TestResourceDataRaw(t, DataVolumeFields(), map[string]interface{}{
		"metadata": []interface{}{
			map[string]interface{}{
				"name":      "test-dv",
				"namespace": "default",
			},
		},
		"spec": []interface{}{
			map[string]interface{}{
				"source": map[string]interface{}{
					"blank": map[string]interface{}{},
				},
			},
		},
	})

	dataVolume, err := FromResourceData(resourceData)
	assert.NoError(t, err)
	assert.NotNil(t, dataVolume)
	assert.Equal(t, "test-dv", dataVolume.Name)
	assert.Equal(t, "default", dataVolume.Namespace)
}

func TestToResourceData(t *testing.T) {
	dv := cdiv1.DataVolume{
		TypeMeta:   metav1.TypeMeta{},
		ObjectMeta: metav1.ObjectMeta{},
		Spec: cdiv1.DataVolumeSpec{
			PriorityClassName: "test-pirioirty",
			ContentType:       "bil",
			Checkpoints:       nil,
			FinalCheckpoint:   false,
			Preallocation:     nil,
		},
		Status: cdiv1.DataVolumeStatus{},
	}

	resourceData := schema.TestResourceDataRaw(t, DataVolumeFields(), map[string]interface{}{})
	err := ToResourceData(dv, resourceData)
	assert.NoError(t, err)
}

func TestExpandDataVolumeStatus(t *testing.T) {
	tests := []struct {
		input    []interface{}
		expected cdiv1.DataVolumeStatus
	}{
		{
			input: []interface{}{
				map[string]interface{}{
					"phase":    "Succeeded",
					"progress": "50%",
				},
			},
			expected: cdiv1.DataVolumeStatus{
				Phase:    cdiv1.DataVolumePhase("Succeeded"),
				Progress: cdiv1.DataVolumeProgress("50%"),
			},
		},
		{
			input:    []interface{}{nil},
			expected: cdiv1.DataVolumeStatus{},
		},
		{
			input:    []interface{}{},
			expected: cdiv1.DataVolumeStatus{},
		},
		{
			input: []interface{}{
				map[string]interface{}{
					"phase": "Failed",
				},
			},
			expected: cdiv1.DataVolumeStatus{
				Phase: cdiv1.DataVolumePhase("Failed"),
			},
		},
	}

	for _, tt := range tests {
		t.Run("", func(t *testing.T) {
			result := expandDataVolumeStatus(tt.input)
			assert.Equal(t, tt.expected, result)
		})
	}
}
