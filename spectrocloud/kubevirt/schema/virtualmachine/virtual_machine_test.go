package virtualmachine

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFromResourceData(t *testing.T) {
	t.Run("empty ResourceData succeeds", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, VirtualMachineFields(), map[string]interface{}{
			"name":        "test-vm",
			"cluster_uid": "cluster-1",
		})
		got, err := FromResourceData(d)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.NotNil(t, got.Metadata)
		assert.Equal(t, "test-vm", got.Metadata.Name)
		require.NotNil(t, got.Spec)
		require.NotNil(t, got.Status)
	})

	t.Run("populated ResourceData", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, VirtualMachineFields(), map[string]interface{}{
			"name":        "test-vm",
			"namespace":   "custom-ns",
			"cluster_uid": "cluster-1",
			"labels": map[string]interface{}{
				"app": "demo",
			},
			"run_strategy": "Always",
			"hostname":     "my-host",
			"status": []interface{}{
				map[string]interface{}{
					"created": true,
					"ready":   true,
				},
			},
		})

		got, err := FromResourceData(d)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "test-vm", got.Metadata.Name)
		assert.Equal(t, "custom-ns", got.Metadata.Namespace)
		assert.Equal(t, map[string]string{"app": "demo"}, got.Metadata.Labels)
		require.NotNil(t, got.Spec)
		assert.Equal(t, "Always", got.Spec.RunStrategy)
		require.NotNil(t, got.Spec.Template)
		require.NotNil(t, got.Spec.Template.Spec)
		assert.Equal(t, "my-host", got.Spec.Template.Spec.Hostname)
		require.NotNil(t, got.Status)
		assert.True(t, got.Status.Created)
		assert.True(t, got.Status.Ready)
	})
}
