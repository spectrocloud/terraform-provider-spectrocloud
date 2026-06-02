package spectrocloud

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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
