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
				{UID: "profile-a", Name: "profile-a-name"},
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
	assert.Equal(t, "profile-a-name", profiles[0].(map[string]interface{})["name"])
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

	t.Run("keeps config values when API values differ only by whitespace", func(t *testing.T) {
		pack := copyPackMap(flattened)
		pack["values"] = "key: value\n\n"
		alignPackStateWithConfig(pack, map[string]interface{}{"name": "heartbeat", "values": "key: value\n"})
		assert.Equal(t, "key: value\n", pack["values"])
	})

	t.Run("keeps API values when content differs from config", func(t *testing.T) {
		pack := copyPackMap(flattened)
		pack["values"] = "key: from-api"
		alignPackStateWithConfig(pack, map[string]interface{}{"name": "heartbeat", "values": "key: from-config"})
		assert.Equal(t, "key: from-api", pack["values"])
	})

	t.Run("name-only config drops every other attribute", func(t *testing.T) {
		pack := copyPackMap(flattened)
		pack["registry_uid"] = "reg-1"
		pack["manifest"] = []interface{}{map[string]interface{}{"name": "m1", "uid": "mu", "content": "c"}}
		alignPackStateWithConfig(pack, map[string]interface{}{
			"name": "heartbeat", "tag": "", "values": "", "uid": "", "type": "",
			"registry_uid": "", "registry_name": "", "manifest": []interface{}{},
		})
		assert.Equal(t, map[string]interface{}{"name": "heartbeat"}, pack)
	})

	t.Run("declared tag and uid keep API values so real drift shows", func(t *testing.T) {
		pack := copyPackMap(flattened)
		alignPackStateWithConfig(pack, map[string]interface{}{"name": "heartbeat", "tag": "2.0.0", "uid": "config-uid"})
		assert.Equal(t, "1.0.0", pack["tag"])
		assert.Equal(t, "6023f4ff4fdd7f7047756474", pack["uid"])
	})

	t.Run("declared registry_name falls back to config when API lookup returned nothing", func(t *testing.T) {
		pack := copyPackMap(flattened)
		alignPackStateWithConfig(pack, map[string]interface{}{"name": "heartbeat", "registry_name": "Public Repo"})
		assert.Equal(t, "Public Repo", pack["registry_name"])
	})

	t.Run("type keeps the config value", func(t *testing.T) {
		pack := copyPackMap(flattened)
		pack["type"] = "helm"
		alignPackStateWithConfig(pack, map[string]interface{}{"name": "heartbeat", "type": "spectro"})
		assert.Equal(t, "spectro", pack["type"])
	})

	t.Run("manifests keep only declared ones in config order with API content", func(t *testing.T) {
		pack := copyPackMap(flattened)
		pack["manifest"] = []interface{}{
			map[string]interface{}{"name": "extra", "uid": "u-extra", "content": "x"},
			map[string]interface{}{"name": "m2", "uid": "u2", "content": "from-api"},
			map[string]interface{}{"name": "m1", "uid": "u1", "content": "kind: A\n\n"},
		}
		alignPackStateWithConfig(pack, map[string]interface{}{
			"name": "heartbeat",
			"manifest": []interface{}{
				map[string]interface{}{"name": "m1", "content": "kind: A\n"},
				map[string]interface{}{"name": "m2", "content": "from-config"},
				map[string]interface{}{"name": "missing", "content": "y"},
			},
		})
		assert.Equal(t, []interface{}{
			map[string]interface{}{"name": "m1", "uid": "u1", "content": "kind: A\n"},
			map[string]interface{}{"name": "m2", "uid": "u2", "content": "from-api"},
		}, pack["manifest"])
	})
}

func TestAlignPacksWithConfigOrderAndMissing(t *testing.T) {
	api := []interface{}{
		map[string]interface{}{"name": "a", "tag": "1"},
		map[string]interface{}{"name": "b", "tag": "2"},
	}
	cfg := []interface{}{
		map[string]interface{}{"name": "b", "tag": "2"},
		map[string]interface{}{"name": "not-on-cluster", "tag": "9"},
		map[string]interface{}{"name": "a", "tag": "1"},
	}
	got := alignPacksWithConfig(api, cfg)
	require.Len(t, got, 2)
	assert.Equal(t, "b", got[0].(map[string]interface{})["name"])
	assert.Equal(t, "a", got[1].(map[string]interface{})["name"])
}

func TestAlignedNameOnlyPackHashMatchesConfig(t *testing.T) {
	r := resourceClusterEks()
	d := r.TestResourceData()
	require.NoError(t, d.Set("cluster_profile", []interface{}{
		map[string]interface{}{
			"id":   "profile-1",
			"pack": []interface{}{map[string]interface{}{"name": "kubernetes-eks"}},
		},
	}))

	fromAPI := []interface{}{map[string]interface{}{
		"id": "profile-1",
		"pack": []interface{}{
			map[string]interface{}{"name": "amazon-linux-eks", "tag": "1.0.0", "values": "x"},
			map[string]interface{}{
				"name": "kubernetes-eks", "tag": "1.33", "type": "spectro", "uid": "u1", "values": "full: yaml\n",
				"manifest": []interface{}{map[string]interface{}{"name": "m1", "uid": "mu", "content": "c"}},
			},
		},
	}}
	alignClusterProfilesStateWithConfig(d, fromAPI)

	// Round-trip through ResourceData so missing keys read back as zero values, as the SDK does
	// before hashing set elements.
	state := r.TestResourceData()
	require.NoError(t, state.Set("cluster_profile", fromAPI))

	hash := schema.HashResource(r.Schema["cluster_profile"].Elem.(*schema.Resource))
	configElem := normalizeInterfaceSliceFromListOrSet(d.Get("cluster_profile"))[0]
	stateElem := normalizeInterfaceSliceFromListOrSet(state.Get("cluster_profile"))[0]
	assert.Equal(t, hash(configElem), hash(stateElem))
}

func TestAlignClusterProfilesStateWithConfigDropsUnconfiguredPacks(t *testing.T) {
	d := resourceClusterEks().TestResourceData()
	require.NoError(t, d.Set("cluster_profile", []interface{}{
		map[string]interface{}{
			"id": "profile-1",
			"pack": []interface{}{
				map[string]interface{}{"name": "kubernetes-eks", "tag": "1.33", "values": "a: 1\n"},
			},
		},
	}))

	clusterProfiles := []interface{}{
		map[string]interface{}{
			"id": "profile-1",
			"pack": []interface{}{
				map[string]interface{}{"name": "amazon-linux-eks", "tag": "1.0.0", "values": "x"},
				map[string]interface{}{"name": "kubernetes-eks", "tag": "1.33", "type": "spectro", "uid": "pack-uid", "values": "a: 1\n\n"},
				map[string]interface{}{"name": "cni-calico", "tag": "3.33.0", "values": "y"},
			},
		},
	}

	alignClusterProfilesStateWithConfig(d, clusterProfiles)

	packs := clusterProfiles[0].(map[string]interface{})["pack"].([]interface{})
	require.Len(t, packs, 1)
	pack := packs[0].(map[string]interface{})
	assert.Equal(t, "kubernetes-eks", pack["name"])
	assert.Equal(t, "a: 1\n", pack["values"])
	_, hasUID := pack["uid"]
	assert.False(t, hasUID)
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

func TestSetClusterProfilesFromAPI(t *testing.T) {
	r := resourceClusterEks()
	d := r.TestResourceData()
	d.SetId("cluster-1")

	cluster := &models.V1SpectroCluster{
		Spec: &models.V1SpectroClusterSpec{
			ClusterProfileTemplates: []*models.V1ClusterProfileTemplate{
				{UID: "api-profile-a"},
				{UID: "api-profile-b"},
			},
		},
	}

	err := setClusterProfilesFromAPI(nil, d, cluster)
	require.NoError(t, err)

	profiles := normalizeInterfaceSliceFromListOrSet(d.Get("cluster_profile"))
	require.Len(t, profiles, 2)
}

func TestRollbackClusterProfileOnUpdateError(t *testing.T) {
	t.Cleanup(func() {
		disableAddonDeploymentResource = false
	})

	r := resourceClusterEks()
	oldProfile := []interface{}{
		map[string]interface{}{"id": "pre-apply-profile"},
	}

	t.Run("flag off restores pre-apply snapshot", func(t *testing.T) {
		disableAddonDeploymentResource = false
		d := r.TestResourceData()
		d.SetId("cluster-1")
		_ = d.Set("cluster_profile", []interface{}{
			map[string]interface{}{"id": "failed-desired-profile"},
		})

		rollbackClusterProfileOnUpdateError(nil, d, oldProfile)

		profiles := normalizeInterfaceSliceFromListOrSet(d.Get("cluster_profile"))
		require.Len(t, profiles, 1)
		assert.Equal(t, "pre-apply-profile", profiles[0].(map[string]interface{})["id"])
	})

	t.Run("flag on without client falls back to pre-apply snapshot", func(t *testing.T) {
		disableAddonDeploymentResource = true
		d := r.TestResourceData()
		d.SetId("cluster-1")
		_ = d.Set("cluster_profile", []interface{}{
			map[string]interface{}{"id": "failed-desired-profile"},
		})

		rollbackClusterProfileOnUpdateError(nil, d, oldProfile)

		profiles := normalizeInterfaceSliceFromListOrSet(d.Get("cluster_profile"))
		require.Len(t, profiles, 1)
		assert.Equal(t, "pre-apply-profile", profiles[0].(map[string]interface{})["id"])
	})

	t.Run("flag on with cluster_template restores pre-apply snapshot", func(t *testing.T) {
		disableAddonDeploymentResource = true
		d := r.TestResourceData()
		d.SetId("cluster-1")
		_ = d.Set("cluster_template", []interface{}{
			map[string]interface{}{"id": "template-1"},
		})
		_ = d.Set("cluster_profile", []interface{}{
			map[string]interface{}{"id": "failed-desired-profile"},
		})

		rollbackClusterProfileOnUpdateError(nil, d, oldProfile)

		profiles := normalizeInterfaceSliceFromListOrSet(d.Get("cluster_profile"))
		require.Len(t, profiles, 1)
		assert.Equal(t, "pre-apply-profile", profiles[0].(map[string]interface{})["id"])
	})
}
