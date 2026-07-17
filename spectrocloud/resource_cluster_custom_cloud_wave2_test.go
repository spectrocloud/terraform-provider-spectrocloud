package spectrocloud

import (
	"context"
	"strings"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	customCloudType           = "nutanix"
	customCloudConfigUID      = "test-cloud-config-id"
	customCloudAccountUID     = "test-custom-account-id-1"
	customCloudPoolYAML       = "kind: KubeadmControlPlane\nmetadata:\n  name: pool-1\nspec:\n  replicas: 3"
	customCloudPool2YAML      = "kind: MachineDeployment\nmetadata:\n  name: pool-2\nspec:\n  replicas: 2\n  template:\n    spec: {}"
	customCloudClusterProfile = "cluster-profile-import-1"
)

func customCloudMachinePoolSet(pools ...map[string]interface{}) *schema.Set {
	items := make([]interface{}, len(pools))
	for i, p := range pools {
		items[i] = p
	}
	return schema.NewSet(resourceMachinePoolCustomCloudHash, items)
}

func prepareCustomCloudResourceData(t *testing.T) *schema.ResourceData {
	t.Helper()
	d := resourceClusterCustomCloud().TestResourceData()
	require.NoError(t, d.Set("name", "test-custom-cluster"))
	require.NoError(t, d.Set("context", "project"))
	require.NoError(t, d.Set("cloud", customCloudType))
	require.NoError(t, d.Set("cloud_account_id", customCloudAccountUID))
	require.NoError(t, d.Set("cloud_config_id", customCloudConfigUID))
	require.NoError(t, d.Set("cluster_profile", []interface{}{
		map[string]interface{}{"id": customCloudClusterProfile},
	}))
	require.NoError(t, d.Set("cloud_config", []interface{}{
		map[string]interface{}{
			"values": "kind: Cluster\nmetadata:\n  name: test-custom-cluster",
		},
	}))
	require.NoError(t, d.Set("machine_pool", []interface{}{
		map[string]interface{}{
			"control_plane":           true,
			"control_plane_as_worker": true,
			"node_pool_config":        customCloudPoolYAML,
		},
	}))
	return d
}

func TestApplyYamlOverridesPath(t *testing.T) {
	yamlContent := `kind: Cluster
metadata:
  name: old-name`

	out, err := applyYamlOverrides(yamlContent, map[string]interface{}{
		"metadata.name": "new-name",
	})
	require.NoError(t, err)
	assert.Contains(t, out, "new-name")
	assert.NotContains(t, out, "old-name")
}

func TestApplyYamlOverridesDocumentKindPrefix(t *testing.T) {
	yamlContent := `kind: Cluster
metadata:
  name: old-name
---
kind: Secret
metadata:
  name: secret-old`

	out, err := applyYamlOverrides(yamlContent, map[string]interface{}{
		"Cluster.metadata.name": "cluster-new",
	})
	require.NoError(t, err)
	assert.Contains(t, out, "cluster-new")
	assert.Contains(t, out, "secret-old")
}

func TestApplyYamlOverridesWithTemplates(t *testing.T) {
	yamlContent := `kind: Cluster
metadata:
  name: cluster-${region}`

	out, err := applyYamlOverridesWithTemplates(yamlContent, map[string]interface{}{
		"region": "us-west-2",
	})
	require.NoError(t, err)
	assert.Contains(t, out, "cluster-us-west-2")
}

func TestSeparateOverrideTypes(t *testing.T) {
	yamlContent := `kind: Cluster
metadata:
  name: ${region}`

	templates, wildcards, fields, paths := separateOverrideTypes(yamlContent, map[string]interface{}{
		"region":          "us-east-1",
		"*metadata.name":  "wildcard-name",
		"metadata.labels": `{"env":"prod"}`,
		"simple":          "value",
	})

	assert.Equal(t, "us-east-1", templates["region"])
	assert.Equal(t, "wildcard-name", wildcards["*metadata.name"])
	assert.Equal(t, `{"env":"prod"}`, fields["metadata.labels"])
	assert.Equal(t, "value", fields["simple"])
	assert.Empty(t, paths)
}

func TestConvertStringToAppropriateType(t *testing.T) {
	assert.Equal(t, true, convertStringToAppropriateType("true"))
	assert.Equal(t, float64(42), convertStringToAppropriateType("42"))
	assert.Equal(t, float64(3.14), convertStringToAppropriateType("3.14"))
	assert.Equal(t, int64(99), convertStringToAppropriateType("099"))
	assert.Equal(t, "plain", convertStringToAppropriateType("plain"))

	jsonVal := convertStringToAppropriateType(`["a","b"]`)
	slice, ok := jsonVal.([]interface{})
	require.True(t, ok)
	assert.Len(t, slice, 2)
}

func TestIsFieldPatternAndWildcardPattern(t *testing.T) {
	assert.True(t, isFieldPattern("metadata.name"))
	assert.False(t, isFieldPattern("items[0].name"))
	assert.True(t, isWildcardPattern("*metadata.name"))
	assert.False(t, isWildcardPattern("metadata.name"))
}

func TestToCustomCloudConfigAppliesOverrides(t *testing.T) {
	d := resourceClusterCustomCloud().TestResourceData()
	require.NoError(t, d.Set("cloud_config", []interface{}{
		map[string]interface{}{
			"values": `kind: Cluster
metadata:
  name: ${region}`,
			"overrides": map[string]interface{}{
				"region": "eu-central-1",
			},
		},
	}))

	cfg := toCustomCloudConfig(d)
	require.NotNil(t, cfg)
	require.NotNil(t, cfg.Values)
	assert.Contains(t, *cfg.Values, "eu-central-1")
	assert.NotContains(t, *cfg.Values, "${region}")
}

func TestResourceClusterCustomCloudStateUpgradeV2(t *testing.T) {
	raw := map[string]interface{}{
		"machine_pool": []interface{}{
			map[string]interface{}{"name": "pool-1"},
		},
	}
	out, err := resourceClusterCustomCloudStateUpgradeV2(context.Background(), raw, nil)
	require.NoError(t, err)
	pools, ok := out["machine_pool"].([]interface{})
	require.True(t, ok)
	assert.Len(t, pools, 1)
}

func TestResourceClusterCustomCloudStateUpgradeV3(t *testing.T) {
	raw := map[string]interface{}{
		"cluster_profile": []interface{}{
			map[string]interface{}{"id": "profile-1"},
		},
	}
	out, err := resourceClusterCustomCloudStateUpgradeV3(context.Background(), raw, nil)
	require.NoError(t, err)
	profiles, ok := out["cluster_profile"].([]interface{})
	require.True(t, ok)
	assert.Len(t, profiles, 1)
}

func TestApplyTemplateSubstitution(t *testing.T) {
	in := "host: ${host}\nport: {{port}}"
	out := applyTemplateSubstitution(in, map[string]interface{}{
		"host": "api.example.com",
		"port": "443",
	})
	assert.True(t, strings.Contains(out, "api.example.com"))
	assert.True(t, strings.Contains(out, "443"))
	assert.False(t, strings.Contains(out, "${host}"))
}

func TestExtractDocumentKind(t *testing.T) {
	var data interface{} = map[string]interface{}{
		"kind": "Cluster",
	}
	assert.Equal(t, "Cluster", extractDocumentKind(data))
	assert.Equal(t, "", extractDocumentKind(map[string]interface{}{}))
}

func TestParseDocumentSpecificPath(t *testing.T) {
	kind, path := parseDocumentSpecificPath("Cluster.metadata.name")
	assert.Equal(t, "Cluster", kind)
	assert.Equal(t, "metadata.name", path)

	kind, path = parseDocumentSpecificPath("metadata.name")
	assert.Equal(t, "", kind)
	assert.Equal(t, "metadata.name", path)
}

func TestResourceClusterCustomCloudReadWithMock(t *testing.T) {
	d := prepareCustomCloudResourceData(t)
	d.SetId("test-cluster-id")

	diags := resourceClusterCustomCloudRead(context.Background(), d, unitTestMockAPIClient)
	assert.Empty(t, diags)
	assert.Equal(t, customCloudConfigUID, d.Get("cloud_config_id"))
	assert.Equal(t, customCloudAccountUID, d.Get("cloud_account_id"))
}

func TestResourceClusterCustomCloudUpdateCloudConfigWithMock(t *testing.T) {
	d := prepareCustomCloudResourceData(t)
	d.SetId("test-cluster-id")

	require.NoError(t, d.Set("cloud_config", []interface{}{
		map[string]interface{}{
			"values": "kind: Cluster\nmetadata:\n  name: old-name",
		},
	}))
	require.NoError(t, d.Set("cloud_config", []interface{}{
		map[string]interface{}{
			"values": "kind: Cluster\nmetadata:\n  name: new-name",
		},
	}))

	diags := resourceClusterCustomCloudUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.Empty(t, diags)
}

func TestResourceClusterCustomCloudCreateWithMock(t *testing.T) {
	d := prepareCustomCloudResourceData(t)
	require.NoError(t, d.Set("tags", []interface{}{"skip_completion"}))

	diags := resourceClusterCustomCloudCreate(context.Background(), d, unitTestMockAPIClient)
	assert.Empty(t, diags)
	assert.Equal(t, "test-custom-cluster-id", d.Id())
}

func TestApplyWildcardPatternOverrides(t *testing.T) {
	yamlContent := `kind: Cluster
metadata:
  hostname: old-host
  labels:
    env: dev`

	out, err := applyWildcardPatternOverrides(yamlContent, map[string]interface{}{
		"*name": "patched",
	})
	require.NoError(t, err)
	assert.Contains(t, out, "patched")
}

func TestApplyFieldPatternOverrides(t *testing.T) {
	yamlContent := `kind: Cluster
spec:
  rootVolume:
    size: 100
  other: 1`

	out, err := applyFieldPatternOverrides(yamlContent, map[string]interface{}{
		"rootVolume.size": "256",
	})
	require.NoError(t, err)
	assert.Contains(t, out, "256")
	assert.NotContains(t, out, "100")
}

func TestApplyYamlOverridesWithTemplatesWildcardAndField(t *testing.T) {
	yamlContent := `kind: Cluster
metadata:
  hostname: host-${env}
spec:
  rootVolume:
    size: 10`

	out, err := applyYamlOverridesWithTemplates(yamlContent, map[string]interface{}{
		"env":             "prod",
		"rootVolume.size": "512",
	})
	require.NoError(t, err)
	assert.Contains(t, out, "host-prod")
	assert.Contains(t, out, "512")
}

func TestExtractMachinePoolNameFromYAML(t *testing.T) {
	name := extractMachinePoolNameFromYAML(map[string]interface{}{
		"node_pool_config": customCloudPoolYAML,
	})
	assert.Equal(t, "pool-1", name)
}

func TestResourceClusterCustomCloudUpdateMachinePoolWithMock(t *testing.T) {
	d := prepareCustomCloudResourceData(t)
	d.SetId("test-cluster-id")

	oldPool := map[string]interface{}{
		"control_plane":           true,
		"control_plane_as_worker": true,
		"node_pool_config":        customCloudPoolYAML,
	}
	updatedYAML := "kind: KubeadmControlPlane\nmetadata:\n  name: pool-1\nspec:\n  replicas: 5"
	newPool := map[string]interface{}{
		"control_plane":           true,
		"control_plane_as_worker": true,
		"node_pool_config":        updatedYAML,
	}

	require.NoError(t, d.Set("machine_pool", customCloudMachinePoolSet(oldPool)))
	require.NoError(t, d.Set("machine_pool", customCloudMachinePoolSet(newPool)))

	diags := resourceClusterCustomCloudUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.Empty(t, diags)
}

func TestResourceClusterCustomCloudUpdateMachinePoolAddWithMock(t *testing.T) {
	d := prepareCustomCloudResourceData(t)
	d.SetId("test-cluster-id")

	pool1 := map[string]interface{}{
		"control_plane":           true,
		"control_plane_as_worker": true,
		"node_pool_config":        customCloudPoolYAML,
	}
	pool2 := map[string]interface{}{
		"control_plane":           false,
		"control_plane_as_worker": false,
		"node_pool_config":        customCloudPool2YAML,
	}

	require.NoError(t, d.Set("machine_pool", customCloudMachinePoolSet(pool1)))
	require.NoError(t, d.Set("machine_pool", customCloudMachinePoolSet(pool1, pool2)))

	diags := resourceClusterCustomCloudUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.Empty(t, diags)
}

func TestResourceClusterCustomCloudUpdateMachinePoolDeleteWithMock(t *testing.T) {
	d := prepareCustomCloudResourceData(t)
	d.SetId("test-cluster-id")

	pool1 := map[string]interface{}{
		"control_plane":           true,
		"control_plane_as_worker": true,
		"node_pool_config":        customCloudPoolYAML,
	}
	pool2 := map[string]interface{}{
		"control_plane":           false,
		"control_plane_as_worker": false,
		"node_pool_config":        customCloudPool2YAML,
	}

	require.NoError(t, d.Set("machine_pool", customCloudMachinePoolSet(pool1, pool2)))
	require.NoError(t, d.Set("machine_pool", customCloudMachinePoolSet(pool1)))

	diags := resourceClusterCustomCloudUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.Empty(t, diags)
}

func TestResourceClusterCustomCloudUpdateClusterProfileWithMock(t *testing.T) {
	d := prepareCustomCloudResourceData(t)
	d.SetId("test-cluster-id")
	require.NoError(t, d.Set("tags", []interface{}{"skip_apply"}))

	setChangedClusterProfiles(t, d,
		[]interface{}{map[string]interface{}{"id": "cluster-profile-import-2"}},
		[]interface{}{map[string]interface{}{"id": "cluster-profile-import-1"}},
	)

	diags := resourceClusterCustomCloudUpdate(context.Background(), d, unitTestMockAPIClient)
	assert.False(t, diags.HasError())
}

func TestFlattenCloudConfigCustomWithMock(t *testing.T) {
	d := prepareCustomCloudResourceData(t)
	c := mustUnitClient(t, false)

	diags, hasError := flattenCloudConfigCustom(customCloudConfigUID, d, c)
	assert.False(t, hasError)
	assert.Empty(t, diags)

	pools, ok := d.Get("machine_pool").(*schema.Set)
	require.True(t, ok)
	assert.Greater(t, pools.Len(), 0)
}
