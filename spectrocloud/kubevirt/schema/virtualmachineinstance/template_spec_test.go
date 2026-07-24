package virtualmachineinstance

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestExpandVirtualMachineInstanceTemplateSpec(t *testing.T) {
	t.Run("empty ResourceData succeeds", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, vmiSpecTestSchema(), map[string]interface{}{})
		got, err := ExpandVirtualMachineInstanceTemplateSpec(d)
		require.NoError(t, err)
		require.NotNil(t, got)
		require.NotNil(t, got.Metadata)
		require.NotNil(t, got.Spec)
		require.NotNil(t, got.Spec.Domain)
	})

	t.Run("populated ResourceData", func(t *testing.T) {
		d := schema.TestResourceDataRaw(t, vmiSpecTestSchema(), map[string]interface{}{
			"hostname":      "my-host",
			"subdomain":     "my-subdomain",
			"dns_policy":    "ClusterFirst",
			"node_selector": map[string]interface{}{"disktype": "ssd"},
		})
		got, err := ExpandVirtualMachineInstanceTemplateSpec(d)
		require.NoError(t, err)
		require.NotNil(t, got.Spec)
		assert.Equal(t, "my-host", got.Spec.Hostname)
		assert.Equal(t, "my-subdomain", got.Spec.Subdomain)
		assert.Equal(t, "ClusterFirst", got.Spec.DNSPolicy)
		assert.Equal(t, map[string]string{"disktype": "ssd"}, got.Spec.NodeSelector)
	})
}

func TestFlattenVirtualMachineInstanceTemplateSpecFromVM(t *testing.T) {
	t.Run("nil input", func(t *testing.T) {
		got := FlattenVirtualMachineInstanceTemplateSpecFromVM(nil, nil)
		require.Len(t, got, 1)
		m := got[0].(map[string]interface{})
		spec, ok := m["spec"].([]interface{})
		require.True(t, ok)
		require.Len(t, spec, 1)
		assert.Empty(t, spec[0].(map[string]interface{}))
	})

	t.Run("populated input", func(t *testing.T) {
		in := &models.V1VMVirtualMachineInstanceTemplateSpec{
			Metadata: &models.V1VMObjectMeta{Name: "vmi-template"},
			Spec: &models.V1VMVirtualMachineInstanceSpec{
				Hostname: "my-host",
			},
		}
		got := FlattenVirtualMachineInstanceTemplateSpecFromVM(in, nil)
		require.Len(t, got, 1)
		m := got[0].(map[string]interface{})
		spec, ok := m["spec"].([]interface{})
		require.True(t, ok)
		require.Len(t, spec, 1)
		specMap := spec[0].(map[string]interface{})
		assert.Equal(t, "my-host", specMap["hostname"])
	})
}
