package spectrocloud

import (
	"context"
	"testing"

	"github.com/hashicorp/terraform-plugin-sdk/v2/helper/schema"
	"github.com/hashicorp/terraform-plugin-sdk/v2/terraform"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// ---------------------------------------------------------------------
// buildProfilesVariablesBatchEntity — assign_strategy "all" vs "cluster".
//
// Regression for a real bug: the builder used to send Clusters: []
// unconditionally for every variable, so the configured value never
// reached Palette's PATCH /v1/clusterTemplates/{uid}/profiles/variables
// regardless of assign_strategy. "all" must now expand to every cluster in
// attached_cluster; "cluster" must expand to exactly the configured
// cluster_ids, both carrying the variable's value.
// ---------------------------------------------------------------------

func variableWithStrategy(name, value, assignStrategy string, clusterIDs ...string) map[string]interface{} {
	ids := schema.NewSet(schema.HashString, nil)
	for _, id := range clusterIDs {
		ids.Add(id)
	}
	return map[string]interface{}{
		"name":            name,
		"value":           value,
		"assign_strategy": assignStrategy,
		"cluster_ids":     ids,
	}
}

func TestBuildProfilesVariablesBatchEntity_AllStrategyUsesAttachedClusters(t *testing.T) {
	d := resourceClusterConfigTemplate().TestResourceData()
	require.NoError(t, d.Set("attached_cluster", []interface{}{
		map[string]interface{}{"cluster_uid": "cluster-1", "name": "c1"},
		map[string]interface{}{"cluster_uid": "cluster-2", "name": "c2"},
	}))

	varsSet := schema.NewSet(resourceClusterConfigTemplateVariableHash, []interface{}{
		variableWithStrategy("region", "us-west-2", "all"),
	})

	got := buildProfilesVariablesBatchEntity(d, []interface{}{
		map[string]interface{}{"id": "profile-1", "variables": varsSet},
	})

	require.Len(t, got.Profiles, 1)
	require.Len(t, got.Profiles[0].Variables, 1)
	clusters := got.Profiles[0].Variables[0].Clusters
	require.Len(t, clusters, 2)
	seen := map[string]string{}
	for _, c := range clusters {
		seen[c.UID] = c.Value
	}
	assert.Equal(t, "us-west-2", seen["cluster-1"])
	assert.Equal(t, "us-west-2", seen["cluster-2"])
}

func TestBuildProfilesVariablesBatchEntity_ClusterStrategyUsesClusterIDs(t *testing.T) {
	d := resourceClusterConfigTemplate().TestResourceData()
	// attached_cluster deliberately includes a cluster NOT in cluster_ids,
	// to confirm "cluster" strategy ignores it.
	require.NoError(t, d.Set("attached_cluster", []interface{}{
		map[string]interface{}{"cluster_uid": "cluster-1", "name": "c1"},
		map[string]interface{}{"cluster_uid": "cluster-2", "name": "c2"},
	}))

	varsSet := schema.NewSet(resourceClusterConfigTemplateVariableHash, []interface{}{
		variableWithStrategy("region", "us-west-2", "cluster", "cluster-2"),
	})

	got := buildProfilesVariablesBatchEntity(d, []interface{}{
		map[string]interface{}{"id": "profile-1", "variables": varsSet},
	})

	require.Len(t, got.Profiles, 1)
	require.Len(t, got.Profiles[0].Variables, 1)
	clusters := got.Profiles[0].Variables[0].Clusters
	require.Len(t, clusters, 1)
	assert.Equal(t, "cluster-2", clusters[0].UID)
	assert.Equal(t, "us-west-2", clusters[0].Value)
}

// ---------------------------------------------------------------------
// validateClusterConfigTemplateVariableAssignment — CustomizeDiff guard.
// ---------------------------------------------------------------------

func baseClusterConfigTemplateRaw(variables map[string]interface{}) map[string]interface{} {
	return map[string]interface{}{
		"name":       "test-template",
		"cloud_type": "aws",
		"cluster_profile": []interface{}{
			map[string]interface{}{
				"id":        "profile-1",
				"variables": []interface{}{variables},
			},
		},
	}
}

func diffClusterConfigTemplate(t *testing.T, raw map[string]interface{}) error {
	t.Helper()
	res := resourceClusterConfigTemplate()
	config := terraform.NewResourceConfigRaw(raw)
	_, err := res.Diff(context.Background(), nil, config, nil)
	return err
}

func TestValidateClusterConfigTemplateVariableAssignment_RejectsMissingClusterIDs(t *testing.T) {
	err := diffClusterConfigTemplate(t, baseClusterConfigTemplateRaw(map[string]interface{}{
		"name":            "region",
		"value":           "us-west-2",
		"assign_strategy": "cluster",
	}))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "cluster_ids is required")
}

func TestValidateClusterConfigTemplateVariableAssignment_RejectsClusterIDsWithAllStrategy(t *testing.T) {
	err := diffClusterConfigTemplate(t, baseClusterConfigTemplateRaw(map[string]interface{}{
		"name":            "region",
		"value":           "us-west-2",
		"assign_strategy": "all",
		"cluster_ids":     []interface{}{"cluster-1"},
	}))
	require.Error(t, err)
	assert.Contains(t, err.Error(), "must be empty when assign_strategy is")
}

func TestValidateClusterConfigTemplateVariableAssignment_AcceptsValidCombinations(t *testing.T) {
	err := diffClusterConfigTemplate(t, baseClusterConfigTemplateRaw(map[string]interface{}{
		"name":            "region",
		"value":           "us-west-2",
		"assign_strategy": "cluster",
		"cluster_ids":     []interface{}{"cluster-1"},
	}))
	require.NoError(t, err)

	err = diffClusterConfigTemplate(t, baseClusterConfigTemplateRaw(map[string]interface{}{
		"name":            "region",
		"value":           "us-west-2",
		"assign_strategy": "all",
	}))
	require.NoError(t, err)
}
