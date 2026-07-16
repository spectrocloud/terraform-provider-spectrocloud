package virtualmachine

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// batch10_test.go — Batch 10.
// Smoke-cover the 0% schema builders in the virtualmachine package and
// pin the expand/flatten helpers for data volume templates, conditions,
// state change requests, and status.

// ---------------------------------------------------------------------------
// Top-level VM fields + status
// ---------------------------------------------------------------------------

func TestVirtualMachineFields(t *testing.T) {
	f := VirtualMachineFields()
	// Just confirm a smattering of the required top-level keys are present.
	for _, k := range []string{
		"name", "namespace", "labels", "annotations",
		"cluster_uid", "run_on_launch", "data_volume_templates",
		"cpu", "memory", "firmware", "features",
		"disk", "interface", "resources", "status",
	} {
		assert.Contains(t, f, k, "expected VM schema key %q", k)
	}
}

func TestVirtualMachineStatusFields(t *testing.T) {
	f := virtualMachineStatusFields()
	for _, k := range []string{"created", "ready", "conditions", "state_change_requests"} {
		assert.Contains(t, f, k)
	}
}

func TestVirtualMachineStatusSchema(t *testing.T) {
	s := virtualMachineStatusSchema()
	require.NotNil(t, s)
	assert.Equal(t, schema.TypeList, s.Type)
	assert.Equal(t, 1, s.MaxItems)
}

// ---------------------------------------------------------------------------
// Conditions
// ---------------------------------------------------------------------------

func TestVirtualMachineConditionsFields(t *testing.T) {
	f := virtualMachineConditionsFields()
	for _, k := range []string{"type", "status", "reason", "message"} {
		assert.Contains(t, f, k)
	}
}

func TestVirtualMachineConditionsSchema(t *testing.T) {
	s := virtualMachineConditionsSchema()
	require.NotNil(t, s)
	assert.Equal(t, schema.TypeList, s.Type)
}

// ---------------------------------------------------------------------------
// State change requests
// ---------------------------------------------------------------------------

func TestVirtualMachineStateChangeRequestFields(t *testing.T) {
	f := virtualMachineStateChangeRequestFields()
	for _, k := range []string{"action", "data", "uid"} {
		assert.Contains(t, f, k)
	}
}

func TestVirtualMachineStateChangeRequestsSchema(t *testing.T) {
	s := virtualMachineStateChangeRequestsSchema()
	require.NotNil(t, s)
	assert.Equal(t, schema.TypeList, s.Type)
}

// ---------------------------------------------------------------------------
// Data volume templates
// ---------------------------------------------------------------------------

func TestDataVolumeFields(t *testing.T) {
	f := DataVolumeFields()
	assert.Contains(t, f, "metadata")
	assert.Contains(t, f, "spec")
}

func TestDataVolumeTemplatesSchema(t *testing.T) {
	s := dataVolumeTemplatesSchema()
	require.NotNil(t, s)
	assert.Equal(t, schema.TypeList, s.Type)
}

func TestExpandDataVolumeTemplates(t *testing.T) {
	t.Run("empty input", func(t *testing.T) {
		got, err := expandDataVolumeTemplates(nil)
		require.NoError(t, err)
		// Empty input returns a zero-length preallocated slice.
		assert.Empty(t, got)
	})

	t.Run("populated with metadata", func(t *testing.T) {
		got, err := expandDataVolumeTemplates([]interface{}{
			map[string]interface{}{
				"metadata": []interface{}{
					map[string]interface{}{
						"name":      "boot-volume",
						"namespace": "default",
					},
				},
				"spec": []interface{}{},
			},
		})
		require.NoError(t, err)
		require.Len(t, got, 1)
		require.NotNil(t, got[0].Metadata)
		assert.Equal(t, "boot-volume", got[0].Metadata.Name)
	})
}

func TestFlattenDataVolumeTemplatesFromVM(t *testing.T) {
	// Nil / empty input → empty slice.
	assert.Empty(t, flattenDataVolumeTemplatesFromVM(nil, nil))

	// Populated + skipped-nil entries.
	got := flattenDataVolumeTemplatesFromVM([]*models.V1VMDataVolumeTemplateSpec{
		nil, // skipped
		{Metadata: &models.V1VMObjectMeta{Name: "dv1", Namespace: "ns"}, Spec: &models.V1VMDataVolumeSpec{}},
	}, nil)
	require.Len(t, got, 1)
	m := got[0].(map[string]interface{})
	assert.Contains(t, m, "metadata")
	assert.Contains(t, m, "spec")
}

// ---------------------------------------------------------------------------
// Spec flatten — nil branch only, exercising FlattenVirtualMachineSpecFromVM
// without needing a fully populated VM.
// ---------------------------------------------------------------------------

func TestFlattenVirtualMachineSpecFromVM_Nil(t *testing.T) {
	got := FlattenVirtualMachineSpecFromVM(nil, nil)
	require.Len(t, got, 1)
	m := got[0].(map[string]interface{})
	assert.Empty(t, m)
}
