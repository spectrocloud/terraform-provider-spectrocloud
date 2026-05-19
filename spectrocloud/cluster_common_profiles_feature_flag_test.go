package spectrocloud

import (
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/spectrocloud/palette-sdk-go/api/models"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFlattenClusterProfilesFromCluster(t *testing.T) {
	cluster := &models.V1SpectroCluster{
		Spec: &models.V1SpectroClusterSpec{
			ClusterProfileTemplates: []*models.V1ClusterProfileTemplate{
				{UID: "profile-a"},
				{UID: "profile-b"},
				nil,
				{UID: ""},
			},
		},
	}

	profiles, err := flattenClusterProfilesFromCluster(cluster)
	require.NoError(t, err)
	require.Len(t, profiles, 2)
	assert.Equal(t, "profile-a", profiles[0].(map[string]interface{})["id"])
	_, hasVariables := profiles[0].(map[string]interface{})["variables"]
	assert.False(t, hasVariables)
	assert.Equal(t, []interface{}{}, profiles[0].(map[string]interface{})["pack"])
	assert.Equal(t, "profile-b", profiles[1].(map[string]interface{})["id"])
	_, hasVariables = profiles[1].(map[string]interface{})["variables"]
	assert.False(t, hasVariables)
	assert.Equal(t, []interface{}{}, profiles[1].(map[string]interface{})["pack"])
}

func TestAlignPackStateWithConfig(t *testing.T) {
	flattened := map[string]interface{}{
		"name":   "heartbeat",
		"tag":    "1.0.0",
		"type":   "spectro",
		"uid":    "6023f4ff4fdd7f7047756474",
		"values": "test",
	}

	t.Run("strips uid and type when omitted from config", func(t *testing.T) {
		pack := copyPackMap(flattened)
		alignPackStateWithConfig(pack, map[string]interface{}{
			"name": "heartbeat",
			"tag":  "1.0.0",
		})
		_, hasUID := pack["uid"]
		_, hasType := pack["type"]
		assert.False(t, hasUID)
		assert.False(t, hasType)
		assert.Equal(t, "heartbeat", pack["name"])
	})

	t.Run("keeps uid and type when present in config", func(t *testing.T) {
		pack := copyPackMap(flattened)
		alignPackStateWithConfig(pack, map[string]interface{}{
			"name": "heartbeat",
			"tag":  "1.0.0",
			"type": "spectro",
			"uid":  "6023f4ff4fdd7f7047756474",
		})
		assert.Equal(t, "6023f4ff4fdd7f7047756474", pack["uid"])
		assert.Equal(t, "spectro", pack["type"])
	})
}

func copyPackMap(src map[string]interface{}) map[string]interface{} {
	dst := make(map[string]interface{}, len(src))
	for k, v := range src {
		dst[k] = v
	}
	return dst
}

func TestShouldSyncClusterProfilesFromAPI(t *testing.T) {
	t.Cleanup(func() {
		disableAddonDeploymentResource = false
	})

	baseSchema := &schema.Resource{
		Schema: map[string]*schema.Schema{
			"cluster_profile": {
				Type:     schema.TypeSet,
				Optional: true,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {Type: schema.TypeString, Required: true},
					},
				},
			},
			"cluster_template": {
				Type:     schema.TypeList,
				Optional: true,
				MaxItems: 1,
				Elem: &schema.Resource{
					Schema: map[string]*schema.Schema{
						"id": {Type: schema.TypeString, Required: true},
					},
				},
			},
		},
	}

	t.Run("false when feature flag is off", func(t *testing.T) {
		disableAddonDeploymentResource = false
		d := baseSchema.TestResourceData()
		assert.False(t, shouldSyncClusterProfilesFromAPI(d))
	})

	t.Run("true when feature flag is on and cluster_template is not used", func(t *testing.T) {
		disableAddonDeploymentResource = true
		d := baseSchema.TestResourceData()
		assert.True(t, shouldSyncClusterProfilesFromAPI(d))
	})

	t.Run("false when feature flag is on but cluster_template is set", func(t *testing.T) {
		disableAddonDeploymentResource = true
		d := baseSchema.TestResourceData()
		_ = d.Set("cluster_template", []interface{}{
			map[string]interface{}{"id": "template-1"},
		})
		assert.False(t, shouldSyncClusterProfilesFromAPI(d))
	})
}

func TestSyncClusterProfilesFromAPIWhenAddonDeploymentDisabled(t *testing.T) {
	t.Cleanup(func() {
		disableAddonDeploymentResource = false
	})
	disableAddonDeploymentResource = true

	r := resourceClusterEks()
	d := r.TestResourceData()
	d.SetId("cluster-1")
	_ = d.Set("cluster_profile", []interface{}{
		map[string]interface{}{"id": "old-profile"},
	})

	cluster := &models.V1SpectroCluster{
		Spec: &models.V1SpectroClusterSpec{
			ClusterProfileTemplates: []*models.V1ClusterProfileTemplate{
				{UID: "profile-from-api"},
			},
		},
	}

	diags := syncClusterProfilesFromAPIWhenAddonDeploymentDisabled(nil, d, cluster)
	require.Empty(t, diags)

	profiles := normalizeInterfaceSliceFromListOrSet(d.Get("cluster_profile"))
	require.Len(t, profiles, 1)
	assert.Equal(t, "profile-from-api", profiles[0].(map[string]interface{})["id"])
}
