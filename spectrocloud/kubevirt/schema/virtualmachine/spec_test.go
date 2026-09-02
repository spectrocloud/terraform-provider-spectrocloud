package virtualmachine

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpandVirtualMachineSpec(t *testing.T) {
	t.Run("empty ResourceData succeeds", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, VirtualMachineFields(), map[string]interface{}{
			"name":        "test-vm",
			"cluster_uid": "cluster-1",
		})
		got, err := ExpandVirtualMachineSpec(d)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.NotNil(t, got.Template)
		require.NotNil(t, got.Template.Spec)
		require.NotNil(t, got.Template.Spec.Domain)
	})

	t.Run("populated ResourceData", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, VirtualMachineFields(), map[string]interface{}{
			"name":                "test-vm",
			"cluster_uid":         "cluster-1",
			"run_strategy":        "Always",
			"hostname":            "my-host",
			"subdomain":           "my-subdomain",
			"dns_policy":          "ClusterFirst",
			"node_selector":       map[string]interface{}{"disktype": "ssd"},
			"priority_class_name": "high-priority",
			"data_volume_templates": []interface{}{
				map[string]interface{}{
					"metadata": []interface{}{
						map[string]interface{}{
							"name":      "boot-volume",
							"namespace": "default",
						},
					},
				},
			},
		})

		got, err := ExpandVirtualMachineSpec(d)
		require.NoError(t, err)
		require.NotNil(t, got)
		assert.Equal(t, "Always", got.RunStrategy)
		require.NotNil(t, got.Template)
		require.NotNil(t, got.Template.Spec)
		assert.Equal(t, "my-host", got.Template.Spec.Hostname)
		assert.Equal(t, "my-subdomain", got.Template.Spec.Subdomain)
		assert.Equal(t, "ClusterFirst", got.Template.Spec.DNSPolicy)
		assert.Equal(t, "high-priority", got.Template.Spec.PriorityClassName)
		require.Len(t, got.DataVolumeTemplates, 1)
		assert.Equal(t, "boot-volume", got.DataVolumeTemplates[0].Metadata.Name)
	})
}
